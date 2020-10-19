package v3

// VerifyMessageRequest contains the verification params.
type VerifyMessageRequest struct {
	VerifyRequest
	// MessageHandlers contains a mapped list of message handlers for a provider
	// that will be rable to produce the correct message format for a given
	// consumer interaction
	MessageHandlers MessageHandlers
}
