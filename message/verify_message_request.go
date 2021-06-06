package message

import "github.com/pact-foundation/pact-go/provider"

// VerifyMessageRequest contains the verification params.
type VerifyMessageRequest struct {
	provider.VerifyRequest
	// MessageHandlers contains a mapped list of message handlers for a provider
	// that will be rable to produce the correct message format for a given
	// consumer interaction
	MessageHandlers MessageHandlers
}
