package graphql

import (
	"encoding/json"
	"fmt"
)

// Subscription transport follow Apollo's subscriptions-transport-ws protocol specification
// https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md

const (
	// The server may responses with this message to the GQL_CONNECTION_INIT from client, indicates the server rejected the connection.
	GQLConnectionError OperationMessageType = "connection_error"
	// Client sends this message to execute GraphQL operation
	GQLStart OperationMessageType = "start"
	// Client sends this message in order to stop a running GraphQL operation execution (for example: unsubscribe)
	GQLStop OperationMessageType = "stop"
	// Client sends this message in order to stop a running GraphQL operation execution (for example: unsubscribe)
	GQLData OperationMessageType = "data"
	// Server message that should be sent right after each GQL_CONNECTION_ACK processed and then periodically to keep the client connection alive.
	// The client starts to consider the keep alive message only upon the first received keep alive message from the server.
	GQLConnectionKeepAlive OperationMessageType = "ka"
	// Client sends this message to terminate the connection.
	GQLConnectionTerminate OperationMessageType = "connection_terminate"

	// Client sends this message after plain websocket connection to start the communication with the server
	// @deprecated: use GQLConnectionInit instead
	GQL_CONNECTION_INIT = GQLConnectionInit
	// The server may responses with this message to the GQL_CONNECTION_INIT from client, indicates the server rejected the connection.
	// @deprecated: use GQLConnectionError instead
	GQL_CONNECTION_ERROR = GQLConnectionError
	// Client sends this message to execute GraphQL operation
	// @deprecated: use GQLStart instead
	GQL_START = GQLStart
	// Client sends this message in order to stop a running GraphQL operation execution (for example: unsubscribe)
	// @deprecated: use GQLStop instead
	GQL_STOP = GQLStop
	// Server sends this message upon a failing operation, before the GraphQL execution, usually due to GraphQL validation errors (resolver errors are part of GQL_DATA message, and will be added as errors array)
	// @deprecated: use GQLError instead
	GQL_ERROR = GQLError
	// The server sends this message to transfer the GraphQL execution result from the server to the client, this message is a response for GQL_START message.
	// @deprecated: use GQLData instead
	GQL_DATA = GQLData
	// Server sends this message to indicate that a GraphQL operation is done, and no more data will arrive for the specific operation.
	// @deprecated: use GQLComplete instead
	GQL_COMPLETE = GQLComplete
	// Server message that should be sent right after each GQL_CONNECTION_ACK processed and then periodically to keep the client connection alive.
	// The client starts to consider the keep alive message only upon the first received keep alive message from the server.
	// @deprecated: use GQLConnectionKeepAlive instead
	GQL_CONNECTION_KEEP_ALIVE = GQLConnectionKeepAlive
	// The server may responses with this message to the GQL_CONNECTION_INIT from client, indicates the server accepted the connection. May optionally include a payload.
	// @deprecated: use GQLConnectionAck instead
	GQL_CONNECTION_ACK = GQLConnectionAck
	// Client sends this message to terminate the connection.
	// @deprecated: use GQLConnectionTerminate instead
	GQL_CONNECTION_TERMINATE = GQLConnectionTerminate
)

type subscriptionsTransportWS struct {
}

// GetSubprotocols returns subprotocol names of the subscription transport
func (stw subscriptionsTransportWS) GetSubprotocols() []string {
	return []string{"graphql-ws"}
}

// ConnectionInit sends a initial request to establish a connection within the existing socket
func (stw *subscriptionsTransportWS) ConnectionInit(ctx *SubscriptionContext, connectionParams map[string]interface{}) error {
	return connectionInit(ctx, connectionParams)
}

// Subscribe requests an graphql operation specified in the payload message
func (stw *subscriptionsTransportWS) Subscribe(ctx *SubscriptionContext, id string, sub Subscription) error {
	if sub.GetStarted() {
		return nil
	}
	payload, err := json.Marshal(sub.GetPayload())
	if err != nil {
		return err
	}
	// send start message to the server
	msg := OperationMessage{
		ID:      id,
		Type:    GQLStart,
		Payload: payload,
	}

	if err := ctx.Send(msg, GQLStart); err != nil {
		return err
	}

	sub.SetStarted(true)
	ctx.SetSubscription(id, &sub)

	return nil
}

// Unsubscribe sends stop message to server and close subscription channel
// The input parameter is subscription ID that is returned from Subscribe function
func (stw *subscriptionsTransportWS) Unsubscribe(ctx *SubscriptionContext, id string) error {
	if ctx == nil || ctx.GetWebsocketConn() == nil {
		return nil
	}
	sub := ctx.GetSubscription(id)

	if sub == nil {
		return fmt.Errorf("subscription id %s doesn't not exist", id)
	}

	ctx.SetSubscription(id, nil)

	// send stop message to the server
	msg := OperationMessage{
		ID:   id,
		Type: GQLStop,
	}

	err := ctx.Send(msg, GQLStop)

	// close the client if there is no running subscription
	if ctx.GetSubscriptionsLength() == 0 {
		ctx.Log("no running subscription. exiting...", "client", GQLInternal)
		return ctx.Close()
	}

	return err
}

// OnMessage listens ongoing messages from server
func (stw *subscriptionsTransportWS) OnMessage(ctx *SubscriptionContext, subscription Subscription, message OperationMessage) {

	switch message.Type {
	case GQLError:
		ctx.Log(message, "server", GQLError)
	case GQLData:
		ctx.Log(message, "server", GQLData)
		var out struct {
			Data   *json.RawMessage
			Errors Errors
		}

		err := json.Unmarshal(message.Payload, &out)
		if err != nil {
			subscription.handler(nil, err)
			return
		}
		if len(out.Errors) > 0 {
			subscription.handler(nil, out.Errors)
			return
		}

		var outData []byte
		if out.Data != nil && len(*out.Data) > 0 {
			outData = *out.Data
		}

		subscription.handler(outData, nil)
	case GQLConnectionError, "conn_err":
		ctx.Log(message, "server", GQLConnectionError)
		_ = stw.Close(ctx)
		_ = ctx.Close()
	case GQLComplete:
		ctx.Log(message, "server", GQLComplete)
		_ = stw.Unsubscribe(ctx, message.ID)
	case GQLConnectionKeepAlive:
		ctx.Log(message, "server", GQLConnectionKeepAlive)
	case GQLConnectionAck:
		// Expected response to the ConnectionInit message from the client acknowledging a successful connection with the server.
		// The client is now ready to request subscription operations.
		ctx.Log(message, "server", GQLConnectionAck)
		ctx.SetAcknowledge(true)
		subscriptions := ctx.GetSubscriptions()
		for id, sub := range subscriptions {
			if err := stw.Subscribe(ctx, id, sub); err != nil {
				_ = stw.Unsubscribe(ctx, id)
				return
			}
		}
		if ctx.OnConnected != nil {
			ctx.OnConnected()
		}
	default:
		ctx.Log(message, "server", GQLUnknown)
	}
}

// Close terminates all subscriptions of the current websocket
func (stw *subscriptionsTransportWS) Close(ctx *SubscriptionContext) error {
	// send terminate message to the server
	msg := OperationMessage{
		Type: GQLConnectionTerminate,
	}

	return ctx.Send(msg, GQLConnectionTerminate)
}
