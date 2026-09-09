// Package v3 implements Pact V3 asynchronous (one-way) message pacts,
// built through the fluent AsynchronousPact API.
package v3

type (
	// Body is the reified content of a message, decoded into whatever
	// type the consumer handler expects.
	Body any
	// Metadata holds message-implementation-specific metadata, such as
	// queue or topic headers, sent alongside a message's content.
	Metadata map[string]any
)

// AsynchronousConsumer receives a message and must be able to parse
// the content.
type AsynchronousConsumer func(MessageContents) error

// MessageContents is a V3 message (asynchronous only).
type MessageContents struct {
	// Message Body
	Content Body `json:"contents"`

	// Message metadata
	Metadata Metadata `json:"metadata"`
}

// Config identifies the consumer, provider and pact output directory for
// an AsynchronousPact.
type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
