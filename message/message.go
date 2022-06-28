package message

import (
	"github.com/pact-foundation/pact-go/v2/models"
)

type Body interface{}
type Metadata map[string]interface{}

// Handler is a provider function that generates a
// message for a Consumer given a Message context (state, description etc.)
type Handler func([]models.ProviderState) (Body, Metadata, error)
type Producer Handler

// Handlers is a list of handlers ordered by description
type Handlers map[string]Handler
