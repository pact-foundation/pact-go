// Package v4 implements Pact V4 message pacts: asynchronous (one-way) and
// synchronous (request/response) message interactions, built through the
// fluent AsynchronousPact and SynchronousPact APIs.
package v4

// Metadata holds message-implementation-specific metadata, such as queue
// or topic headers, sent alongside a message's content.
type Metadata map[string]any

// AsynchronousMessage is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// AsynchronousMessage is the main implementation of the Pact AsynchronousMessage interface.
type AsynchronousMessage MessageContents

// AsynchronousConsumer receives a message and must be able to parse
// the content.
type AsynchronousConsumer func(AsynchronousMessage) error

// MessageContents holds a V4 message's body and metadata, used for both
// AsynchronousMessage (as the whole message) and SynchronousMessage's
// separate Request and Responses.
type MessageContents struct {
	// Message Body
	Contents []byte

	// Body is the attempt to reify the message body back into a specified type
	// Not populated for synchronous  messages
	Body any `json:"contents"`

	// Message metadata. Currently not populated for synchronous messages
	// Metadata field (type Metadata): `json:"metadata"`
}

// Config identifies the consumer, provider and pact output directory for
// an AsynchronousPact or SynchronousPact.
type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
