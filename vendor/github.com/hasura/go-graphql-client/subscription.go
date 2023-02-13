package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// SubscriptionProtocolType represents the protocol specification enum of the subscription
type SubscriptionProtocolType String

const (
	SubscriptionsTransportWS SubscriptionProtocolType = "subscriptions-transport-ws"
	GraphQLWS                SubscriptionProtocolType = "graphql-ws"

	// Receiving a message of a type or format which is not specified in this document
	// The <error-message> can be vaguely descriptive on why the received message is invalid.
	StatusInvalidMessage websocket.StatusCode = 4400
	// if the connection is not acknowledged, the socket will be closed immediately with the event 4401: Unauthorized
	StatusUnauthorized websocket.StatusCode = 4401
	// Connection initialisation timeout
	StatusConnectionInitialisationTimeout websocket.StatusCode = 4408
	// Subscriber for <generated-id> already exists
	StatusSubscriberAlreadyExists websocket.StatusCode = 4409
	// Too many initialisation requests
	StatusTooManyInitialisationRequests websocket.StatusCode = 4429
)

// OperationMessageType represents a subscription message enum type
type OperationMessageType string

const (
	// Unknown operation type, for logging only
	GQLUnknown OperationMessageType = "unknown"
	// Internal status, for logging only
	GQLInternal OperationMessageType = "internal"

	// @deprecated: use GQLUnknown instead
	GQL_UNKNOWN = GQLUnknown
	// @deprecated: use GQLInternal instead
	GQL_INTERNAL = GQLInternal
)

// ErrSubscriptionStopped a special error which forces the subscription stop
var ErrSubscriptionStopped = errors.New("subscription stopped")

// OperationMessage represents a subscription operation message
type OperationMessage struct {
	ID      string               `json:"id,omitempty"`
	Type    OperationMessageType `json:"type"`
	Payload json.RawMessage      `json:"payload,omitempty"`
}

// String overrides the default Stringer to return json string for debugging
func (om OperationMessage) String() string {
	bs, _ := json.Marshal(om)

	return string(bs)
}

// WebsocketHandler abstracts WebSocket connection functions
// ReadJSON and WriteJSON data of a frame from the WebSocket connection.
// Close the WebSocket connection.
type WebsocketConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
	// SetReadLimit sets the maximum size in bytes for a message read from the peer. If a
	// message exceeds the limit, the connection sends a close message to the peer
	// and returns ErrReadLimit to the application.
	SetReadLimit(limit int64)
}

// SubscriptionProtocol abstracts the life-cycle of subscription protocol implementation for a specific transport protocol
type SubscriptionProtocol interface {
	// GetSubprotocols returns subprotocol names of the subscription transport
	// The graphql server depends on the Sec-WebSocket-Protocol header to return the correct message specification
	GetSubprotocols() []string
	// ConnectionInit sends a initial request to establish a connection within the existing socket
	ConnectionInit(ctx *SubscriptionContext, connectionParams map[string]interface{}) error
	// Subscribe requests an graphql operation specified in the payload message
	Subscribe(ctx *SubscriptionContext, id string, sub *Subscription) error
	// Unsubscribe sends a request to stop listening and complete the subscription
	Unsubscribe(ctx *SubscriptionContext, id string) error
	// OnMessage listens ongoing messages from server
	OnMessage(ctx *SubscriptionContext, subscription *Subscription, message OperationMessage)
	// Close terminates all subscriptions of the current websocket
	Close(ctx *SubscriptionContext) error
}

// SubscriptionContext represents a shared context for protocol implementations with the websocket connection inside
type SubscriptionContext struct {
	context.Context
	WebsocketConn
	OnConnected      func()
	onDisconnected   func()
	cancel           context.CancelFunc
	subscriptions    map[string]*Subscription
	disabledLogTypes []OperationMessageType
	log              func(args ...interface{})
	acknowledged     int64
	exitStatusCodes  []int
	mutex            sync.Mutex
}

// Log prints condition logging with message type filters
func (sc *SubscriptionContext) Log(message interface{}, source string, opType OperationMessageType) {
	if sc == nil || sc.log == nil {
		return
	}
	for _, ty := range sc.disabledLogTypes {
		if ty == opType {
			return
		}
	}

	sc.log(message, source)
}

// GetWebsocketConn get the current websocket connection
func (sc *SubscriptionContext) GetWebsocketConn() WebsocketConn {
	return sc.WebsocketConn
}

// SetWebsocketConn set the current websocket connection
func (sc *SubscriptionContext) SetWebsocketConn(conn WebsocketConn) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	sc.WebsocketConn = conn
}

// GetSubscription get the subscription state by id
func (sc *SubscriptionContext) GetSubscription(id string) *Subscription {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	if sc.subscriptions == nil {
		return nil
	}
	sub, _ := sc.subscriptions[id]
	return sub
}

// GetSubscription get all available subscriptions in the context
func (sc *SubscriptionContext) GetSubscriptions() map[string]*Subscription {
	newMap := make(map[string]*Subscription)
	for k, v := range sc.subscriptions {
		newMap[k] = v
	}
	return newMap
}

// SetSubscription set the input subscription state into the context
// if subscription is nil, removes the subscription from the map
func (sc *SubscriptionContext) SetSubscription(id string, sub *Subscription) {
	sc.mutex.Lock()
	if sub == nil {
		delete(sc.subscriptions, id)
	} else {
		sc.subscriptions[id] = sub
	}
	sc.mutex.Unlock()
}

// GetAcknowledge get the acknowledge status
func (sc *SubscriptionContext) GetAcknowledge() bool {
	return atomic.LoadInt64(&sc.acknowledged) > 0
}

// SetAcknowledge set the acknowledge status
func (sc *SubscriptionContext) SetAcknowledge(value bool) {
	if value {
		atomic.StoreInt64(&sc.acknowledged, 1)
	} else {
		atomic.StoreInt64(&sc.acknowledged, 0)
	}
}

// Close closes the context and the inner websocket connection if exists
func (sc *SubscriptionContext) Close() error {
	if conn := sc.GetWebsocketConn(); conn != nil {
		err := conn.Close()
		sc.SetWebsocketConn(nil)
		if err != nil {
			return err
		}
	}
	if sc.cancel != nil {
		sc.cancel()
	}

	return nil
}

// Send emits a message to the graphql server
func (sc *SubscriptionContext) Send(message interface{}, opType OperationMessageType) error {
	if conn := sc.GetWebsocketConn(); conn != nil {
		sc.Log(message, "client", opType)
		return conn.WriteJSON(message)
	}
	return nil
}

type handlerFunc func(data []byte, err error) error

// Subscription stores the subscription declaration and its state
type Subscription struct {
	payload GraphQLRequestPayload
	handler func(data []byte, err error)
	started bool
}

// GetPayload returns the graphql request payload
func (s Subscription) GetPayload() GraphQLRequestPayload {
	return s.payload
}

// GetStarted a public getter for the started status
func (s Subscription) GetStarted() bool {
	return s.started
}

// SetStarted a public getter for the started status
func (s *Subscription) SetStarted(value bool) {
	s.started = value
}

// GetHandler a public getter for the subscription handler
func (s Subscription) GetHandler() func(data []byte, err error) {
	return s.handler
}

// SubscriptionClient is a GraphQL subscription client.
type SubscriptionClient struct {
	url                string
	context            *SubscriptionContext
	connectionParams   map[string]interface{}
	connectionParamsFn func() map[string]interface{}
	protocol           SubscriptionProtocol
	websocketOptions   WebsocketOptions
	timeout            time.Duration
	isRunning          int64
	readLimit          int64 // max size of response message. Default 10 MB
	createConn         func(sc *SubscriptionClient) (WebsocketConn, error)
	retryTimeout       time.Duration
	onError            func(sc *SubscriptionClient, err error) error
	errorChan          chan error
}

// NewSubscriptionClient constructs new subscription client
func NewSubscriptionClient(url string) *SubscriptionClient {
	return &SubscriptionClient{
		url:          url,
		timeout:      time.Minute,
		readLimit:    10 * 1024 * 1024, // set default limit 10MB
		createConn:   newWebsocketConn,
		retryTimeout: time.Minute,
		errorChan:    make(chan error),
		protocol:     &subscriptionsTransportWS{},
		context: &SubscriptionContext{
			subscriptions: make(map[string]*Subscription),
		},
	}
}

// GetURL returns GraphQL server's URL
func (sc *SubscriptionClient) GetURL() string {
	return sc.url
}

// GetTimeout returns write timeout of websocket client
func (sc *SubscriptionClient) GetTimeout() time.Duration {
	return sc.timeout
}

// GetContext returns current context of subscription client
func (sc *SubscriptionClient) GetContext() context.Context {
	return sc.context.Context
}

// WithWebSocket replaces customized websocket client constructor
// In default, subscription client uses https://github.com/nhooyr/websocket
func (sc *SubscriptionClient) WithWebSocket(fn func(sc *SubscriptionClient) (WebsocketConn, error)) *SubscriptionClient {
	sc.createConn = fn
	return sc
}

// WithProtocol changes the subscription protocol implementation
// By default the subscription client uses the subscriptions-transport-ws protocol
func (sc *SubscriptionClient) WithProtocol(protocol SubscriptionProtocolType) *SubscriptionClient {

	switch protocol {
	case GraphQLWS:
		sc.protocol = &graphqlWS{}
	case SubscriptionsTransportWS:
		sc.protocol = &subscriptionsTransportWS{}
	default:
		panic(fmt.Sprintf("unknown subscription protocol %s", protocol))
	}
	return sc
}

// WithWebSocketOptions provides options to the websocket client
func (sc *SubscriptionClient) WithWebSocketOptions(options WebsocketOptions) *SubscriptionClient {
	sc.websocketOptions = options
	return sc
}

// WithConnectionParams updates connection params for sending to server through GQL_CONNECTION_INIT event
// It's usually used for authentication handshake
func (sc *SubscriptionClient) WithConnectionParams(params map[string]interface{}) *SubscriptionClient {
	sc.connectionParams = params
	return sc
}

// WithConnectionParamsFn set a function that returns connection params for sending to server through GQL_CONNECTION_INIT event
// It's suitable for short-lived access tokens that need to be refreshed frequently
func (sc *SubscriptionClient) WithConnectionParamsFn(fn func() map[string]interface{}) *SubscriptionClient {
	sc.connectionParamsFn = fn
	return sc
}

// WithTimeout updates write timeout of websocket client
func (sc *SubscriptionClient) WithTimeout(timeout time.Duration) *SubscriptionClient {
	sc.timeout = timeout
	return sc
}

// WithRetryTimeout updates reconnecting timeout. When the websocket server was stopped, the client will retry connecting every second until timeout
// The zero value means unlimited timeout
func (sc *SubscriptionClient) WithRetryTimeout(timeout time.Duration) *SubscriptionClient {
	sc.retryTimeout = timeout
	return sc
}

// WithLog sets logging function to print out received messages. By default, nothing is printed
func (sc *SubscriptionClient) WithLog(logger func(args ...interface{})) *SubscriptionClient {
	sc.context.log = logger
	return sc
}

// WithoutLogTypes these operation types won't be printed
func (sc *SubscriptionClient) WithoutLogTypes(types ...OperationMessageType) *SubscriptionClient {
	sc.context.disabledLogTypes = types
	return sc
}

// WithReadLimit set max size of response message
func (sc *SubscriptionClient) WithReadLimit(limit int64) *SubscriptionClient {
	sc.readLimit = limit
	return sc
}

// OnError event is triggered when there is any connection error. This is bottom exception handler level
// If this function is empty, or returns nil, the error is ignored
// If returns error, the websocket connection will be terminated
func (sc *SubscriptionClient) OnError(onError func(sc *SubscriptionClient, err error) error) *SubscriptionClient {
	sc.onError = onError
	return sc
}

// OnConnected event is triggered when the websocket connected to GraphQL server successfully
func (sc *SubscriptionClient) OnConnected(fn func()) *SubscriptionClient {
	sc.context.OnConnected = fn
	return sc
}

// OnDisconnected event is triggered when the websocket client was disconnected
func (sc *SubscriptionClient) OnDisconnected(fn func()) *SubscriptionClient {
	sc.context.onDisconnected = fn
	return sc
}

// set the running atomic lock status
func (sc *SubscriptionClient) setIsRunning(value bool) {
	if value {
		atomic.StoreInt64(&sc.isRunning, 1)
	} else {
		atomic.StoreInt64(&sc.isRunning, 0)
	}
}

// initializes the websocket connection
func (sc *SubscriptionClient) init() error {

	now := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	sc.context.Context = ctx
	sc.context.cancel = cancel

	for {
		var err error
		var conn WebsocketConn
		// allow custom websocket client
		if sc.context.GetWebsocketConn() == nil {
			conn, err = sc.createConn(sc)
			if err == nil {
				sc.context.SetWebsocketConn(conn)
			}
		}

		if err == nil {
			sc.context.SetReadLimit(sc.readLimit)
			// send connection init event to the server
			connectionParams := sc.connectionParams
			if sc.connectionParamsFn != nil {
				connectionParams = sc.connectionParamsFn()
			}
			err = sc.protocol.ConnectionInit(sc.context, connectionParams)
		}

		if err == nil {
			return nil
		}

		if sc.retryTimeout > 0 && now.Add(sc.retryTimeout).Before(time.Now()) {
			if sc.context.onDisconnected != nil {
				sc.context.onDisconnected()
			}
			return err
		}
		sc.context.Log(fmt.Sprintf("%s. retry in second...", err.Error()), "client", GQLInternal)
		time.Sleep(time.Second)
	}
}

// Subscribe sends start message to server and open a channel to receive data.
// The handler callback function will receive raw message data or error. If the call return error, onError event will be triggered
// The function returns subscription ID and error. You can use subscription ID to unsubscribe the subscription
func (sc *SubscriptionClient) Subscribe(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...Option) (string, error) {
	return sc.do(v, variables, handler, options...)
}

// NamedSubscribe sends start message to server and open a channel to receive data, with operation name
//
// Deprecated: this is the shortcut of Subscribe method, with NewOperationName option
func (sc *SubscriptionClient) NamedSubscribe(name string, v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...Option) (string, error) {
	return sc.do(v, variables, handler, append(options, OperationName(name))...)
}

// SubscribeRaw sends start message to server and open a channel to receive data, with raw query
// Deprecated: use Exec instead
func (sc *SubscriptionClient) SubscribeRaw(query string, variables map[string]interface{}, handler func(message []byte, err error) error) (string, error) {
	return sc.doRaw(query, variables, handler)
}

// Exec sends start message to server and open a channel to receive data, with raw query
func (sc *SubscriptionClient) Exec(query string, variables map[string]interface{}, handler func(message []byte, err error) error) (string, error) {
	return sc.doRaw(query, variables, handler)
}

func (sc *SubscriptionClient) do(v interface{}, variables map[string]interface{}, handler func(message []byte, err error) error, options ...Option) (string, error) {
	query, err := ConstructSubscription(v, variables, options...)
	if err != nil {
		return "", err
	}

	return sc.doRaw(query, variables, handler)
}

func (sc *SubscriptionClient) doRaw(query string, variables map[string]interface{}, handler func(message []byte, err error) error) (string, error) {
	id := uuid.New().String()

	sub := Subscription{
		payload: GraphQLRequestPayload{
			Query:     query,
			Variables: variables,
		},
		handler: sc.wrapHandler(handler),
	}

	// if the websocket client is running, start subscription immediately
	if atomic.LoadInt64(&sc.isRunning) > 0 {
		if err := sc.protocol.Subscribe(sc.context, id, &sub); err != nil {
			return "", err
		}
	}

	sc.context.SetSubscription(id, &sub)

	return id, nil
}

func (sc *SubscriptionClient) wrapHandler(fn handlerFunc) func(data []byte, err error) {
	return func(data []byte, err error) {
		if errValue := fn(data, err); errValue != nil {
			sc.errorChan <- errValue
		}
	}
}

// Unsubscribe sends stop message to server and close subscription channel
// The input parameter is subscription ID that is returned from Subscribe function
func (sc *SubscriptionClient) Unsubscribe(id string) error {
	return sc.protocol.Unsubscribe(sc.context, id)
}

// Run start websocket client and subscriptions. If this function is run with goroutine, it can be stopped after closed
func (sc *SubscriptionClient) Run() error {
	if err := sc.init(); err != nil {
		return fmt.Errorf("retry timeout. exiting...")
	}

	sc.setIsRunning(true)
	go func() {
		for atomic.LoadInt64(&sc.isRunning) > 0 {
			select {
			case <-sc.context.Done():
				return
			default:
				if sc.context == nil || sc.context.GetWebsocketConn() == nil {
					return
				}

				var message OperationMessage
				if err := sc.context.ReadJSON(&message); err != nil {
					// manual EOF check
					if err == io.EOF || strings.Contains(err.Error(), "EOF") {
						if err = sc.Reset(); err != nil {
							sc.errorChan <- err
							return
						}
					}
					closeStatus := websocket.CloseStatus(err)
					switch closeStatus {
					case websocket.StatusNormalClosure, websocket.StatusAbnormalClosure:
						// close event from websocket client, exiting...
						return
					case StatusConnectionInitialisationTimeout, StatusTooManyInitialisationRequests, StatusSubscriberAlreadyExists, StatusUnauthorized:
						sc.context.Log(err, "server", GQLError)
						return
					}

					if closeStatus != -1 && closeStatus < 3000 && closeStatus > 4999 {
						sc.context.Log(fmt.Sprintf("%s. Retry connecting...", err), "client", GQLInternal)
						if err = sc.Reset(); err != nil {
							sc.errorChan <- err
							return
						}
					}

					if sc.onError != nil {
						if err = sc.onError(sc, err); err != nil {
							// end the subscription if the callback return error
							sc.Close()
							return
						}
					}
					continue
				}

				sub := sc.context.GetSubscription(message.ID)
				go sc.protocol.OnMessage(sc.context, sub, message)
			}
		}
	}()

	for atomic.LoadInt64(&sc.isRunning) > 0 {
		select {
		case <-sc.context.Done():
			return nil
		case e := <-sc.errorChan:
			// stop the subscription if the error has stop message
			if e == ErrSubscriptionStopped {
				return nil
			}

			if sc.onError != nil {
				if err := sc.onError(sc, e); err != nil {
					return err
				}
			}
		}
	}
	// if the running status is false, stop retrying
	if atomic.LoadInt64(&sc.isRunning) == 0 {
		return nil
	}

	return sc.Reset()
}

// Reset restart websocket connection and subscriptions
func (sc *SubscriptionClient) Reset() error {
	sc.context.SetAcknowledge(false)
	isRunning := atomic.LoadInt64(&sc.isRunning) == 0

	for id, sub := range sc.context.GetSubscriptions() {
		sub.SetStarted(false)
		if isRunning {
			_ = sc.protocol.Unsubscribe(sc.context, id)
			sc.context.SetSubscription(id, sub)
		}
	}

	if sc.context.GetWebsocketConn() != nil {
		_ = sc.protocol.Close(sc.context)
		_ = sc.context.Close()
		sc.context.SetWebsocketConn(nil)
	}

	return sc.Run()
}

// Close closes all subscription channel and websocket as well
func (sc *SubscriptionClient) Close() (err error) {
	sc.setIsRunning(false)
	for id := range sc.context.GetSubscriptions() {
		if err = sc.protocol.Unsubscribe(sc.context, id); err != nil {
			sc.context.cancel()
			return
		}
	}

	if sc.context != nil {
		_ = sc.protocol.Close(sc.context)
		err = sc.context.Close()
		sc.context.SetWebsocketConn(nil)
		if sc.context.onDisconnected != nil {
			sc.context.onDisconnected()
		}
	}

	return
}

// the reusable function for sending connection init message.
// The payload format of both subscriptions-transport-ws and graphql-ws are the same
func connectionInit(conn *SubscriptionContext, connectionParams map[string]interface{}) error {
	var bParams []byte = nil
	var err error
	if connectionParams != nil {
		bParams, err = json.Marshal(connectionParams)
		if err != nil {
			return err
		}
	}

	// send connection_init event to the server
	msg := OperationMessage{
		Type:    GQLConnectionInit,
		Payload: bParams,
	}

	return conn.Send(msg, GQLConnectionInit)
}

// default websocket handler implementation using https://github.com/nhooyr/websocket
type WebsocketHandler struct {
	ctx     context.Context
	timeout time.Duration
	*websocket.Conn
}

// WriteJSON implements the function to encode and send message in json format to the server
func (wh *WebsocketHandler) WriteJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()

	return wsjson.Write(ctx, wh.Conn, v)
}

// ReadJSON implements the function to decode the json message from the server
func (wh *WebsocketHandler) ReadJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()
	return wsjson.Read(ctx, wh.Conn, v)
}

// Close implements the function to close the websocket connection
func (wh *WebsocketHandler) Close() error {
	return wh.Conn.Close(websocket.StatusNormalClosure, "close websocket")
}

// the default constructor function to create a websocket client
// which uses https://github.com/nhooyr/websocket library
func newWebsocketConn(sc *SubscriptionClient) (WebsocketConn, error) {

	options := &websocket.DialOptions{
		Subprotocols: sc.protocol.GetSubprotocols(),
		HTTPClient:   sc.websocketOptions.HTTPClient,
	}

	c, _, err := websocket.Dial(sc.GetContext(), sc.GetURL(), options)
	if err != nil {
		return nil, err
	}

	return &WebsocketHandler{
		ctx:     sc.GetContext(),
		Conn:    c,
		timeout: sc.GetTimeout(),
	}, nil
}

// WebsocketOptions allows implementation agnostic configuration of the websocket client
type WebsocketOptions struct {
	// HTTPClient is used for the connection.
	HTTPClient *http.Client
}
