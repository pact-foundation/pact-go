package v4

type Body interface{}
type Metadata map[string]interface{}

// AsynchronousConsumer receives a message and must be able to parse
// the content
type AsynchronousConsumer func(MessageContents) error

// V3 Message (Asynchronous only)
type MessageContents struct {
	// Message Body
	Content Body `json:"contents"`

	// Message metadata
	Metadata Metadata `json:"metadata"`
}

type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
