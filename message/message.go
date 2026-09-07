// Package message implements Pact message pact verification: matching an
// incoming message description to a Handler that produces the message a
// consumer expects, and returning its body and metadata to the verifier.
package message

import (
	"github.com/pact-foundation/pact-go/v2/models"
)

type (
	// Body is the reified content of a message, decoded into whatever
	// type the consumer handler expects.
	Body any
	// Metadata holds message-implementation-specific metadata, such as
	// queue or topic headers, sent alongside a message's content.
	Metadata map[string]any
)

// Handler is a provider function that generates a
// message for a Consumer given a Message context (state, description etc.)
type (
	Handler func([]models.ProviderState) (Body, Metadata, error)
	// Producer has the same underlying type as Handler, but is a
	// distinct defined type: converting between the two requires an
	// explicit conversion (e.g. Producer(h) or Handler(p)).
	Producer Handler
)

// Handlers is a list of handlers ordered by description.
type Handlers map[string]Handler
