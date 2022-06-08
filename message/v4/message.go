package v4

import (
	"github.com/pact-foundation/pact-go/v2/models"
)

type Body interface{}
type Metadata map[string]interface{}

// Handler is a provider function that generates a
// message for a Consumer given a Message context (state, description etc.)
type Handler func([]models.V3ProviderState) (Body, Metadata, error)
type Producer Handler

// Handlers is a list of handlers ordered by description
type Handlers map[string]Handler

// AsynchronousConsumer receives a message and must be able to parse
// the content
// type AsynchronousConsumer func(AsynchronousMessage) error

// // type SynchronousConsumer func(SynchronousMessage) error

// // V3 Message (Asynchronous only)
// type AsynchronousMessage struct {
// 	// Message Body
// 	Content interface{} `json:"contents"`

// 	// Provider state to be written into the Pact file
// 	States []models.V3ProviderState `json:"providerStates"`

// 	// Message metadata
// 	Metadata matchers.MetadataMatcher `json:"metadata"`

// 	// Description to be written into the Pact file
// 	Description string `json:"description"`
// }

type Config struct {
	Consumer string
	Provider string
	PactDir  string
}
