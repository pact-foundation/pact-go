package graphql

import (
	"encoding/json"
	"fmt"
)

// This package implements GraphQL over WebSocket Protocol (graphql-ws)
// https://github.com/enisdenjo/graphql-ws/blob/master/PROTOCOL.md

const (
	// Indicates that the client wants to establish a connection within the existing socket.
	// This connection is not the actual WebSocket communication channel, but is rather a frame within it asking the server to allow future operation requests.
	GQLConnectionInit OperationMessageType = "connection_init"
	// Expected response to the ConnectionInit message from the client acknowledging a successful connection with the server.
	GQLConnectionAck OperationMessageType = "connection_ack"
	// The Ping message can be sent at any time within the established socket.
	GQLPing OperationMessageType = "ping"
	// The response to the Ping message. Must be sent as soon as the Ping message is received.
	GQLPong OperationMessageType = "pong"
	// Requests an operation specified in the message payload. This message provides a unique ID field to connect published messages to the operation requested by this message.
	GQLSubscribe OperationMessageType = "subscribe"
	// Operation execution result(s) from the source stream created by the binding Subscribe message. After all results have been emitted, the Complete message will follow indicating stream completion.
	GQLNext OperationMessageType = "next"
	// Operation execution error(s) in response to the Subscribe message.
	// This can occur before execution starts, usually due to validation errors, or during the execution of the request.
	GQLError OperationMessageType = "error"
	// indicates that the requested operation execution has completed. If the server dispatched the Error message relative to the original Subscribe message, no Complete message will be emitted.
	GQLComplete OperationMessageType = "complete"
)

type graphqlWS struct {
}

// GetSubprotocols returns subprotocol names of the subscription transport
func (gws graphqlWS) GetSubprotocols() []string {
	return []string{"graphql-transport-ws"}
}

// ConnectionInit sends a initial request to establish a connection within the existing socket
func (gws *graphqlWS) ConnectionInit(ctx *SubscriptionContext, connectionParams map[string]interface{}) error {
	return connectionInit(ctx, connectionParams)
}

// Subscribe requests an graphql operation specified in the payload message
func (gws *graphqlWS) Subscribe(ctx *SubscriptionContext, id string, sub Subscription) error {
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
		Type:    GQLSubscribe,
		Payload: payload,
	}

	if err := ctx.Send(msg, GQLSubscribe); err != nil {
		return err
	}

	sub.SetStarted(true)
	ctx.SetSubscription(id, &sub)

	return nil
}

// Unsubscribe sends stop message to server and close subscription channel
// The input parameter is subscription ID that is returned from Subscribe function
func (gws *graphqlWS) Unsubscribe(ctx *SubscriptionContext, id string) error {
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
		Type: GQLComplete,
	}

	err := ctx.Send(msg, GQLComplete)
	// close the client if there is no running subscription
	if ctx.GetSubscriptionsLength() == 0 {
		ctx.Log("no running subscription. exiting...", "client", GQLInternal)
		return ctx.Close()
	}

	return err
}

// OnMessage listens ongoing messages from server
func (gws *graphqlWS) OnMessage(ctx *SubscriptionContext, subscription Subscription, message OperationMessage) {

	switch message.Type {
	case GQLError:
		ctx.Log(message, "server", message.Type)
	case GQLNext:
		ctx.Log(message, "server", message.Type)
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
	case GQLComplete:
		ctx.Log(message, "server", message.Type)
		_ = gws.Unsubscribe(ctx, message.ID)
	case GQLPing:
		ctx.Log(message, "server", GQLPing)
		// send pong response message back to the server
		msg := OperationMessage{
			Type:    GQLPong,
			Payload: message.Payload,
		}

		if err := ctx.Send(msg, GQLPong); err != nil {
			ctx.Log(err, "client", GQLInternal)
		}
	case GQLConnectionAck:
		// Expected response to the ConnectionInit message from the client acknowledging a successful connection with the server.
		// The client is now ready to request subscription operations.
		ctx.Log(message, "server", GQLConnectionAck)
		ctx.SetAcknowledge(true)
		for id, sub := range ctx.GetSubscriptions() {
			if err := gws.Subscribe(ctx, id, sub); err != nil {
				gws.Unsubscribe(ctx, id)
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
func (gws *graphqlWS) Close(conn *SubscriptionContext) error {
	return nil
}
