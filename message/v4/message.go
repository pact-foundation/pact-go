package v4

type Metadata map[string]interface{}

// AsynchronousMessage is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// AsynchronousMessage is the main implementation of the Pact AsynchronousMessage interface.
type AsynchronousMessage MessageContents

// AsynchronousConsumer receives a message and must be able to parse
// the content
type AsynchronousConsumer func(AsynchronousMessage) error

// V3 Message (Asynchronous only)
type MessageContents struct {
	// Message Body
	Contents []byte

	// Body is the attempt to reify the message body back into a specified type
	// Not populated for synchronous  messages
	Body interface{} `json:"contents"`

	// Message metadata. Currently not populated for synchronous messages
	// Metadata Metadata `json:"metadata"`
}

type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
