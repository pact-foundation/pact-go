package v3

// InteractionV2 sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type InteractionV2 struct {
	Interaction

	// Provider state to be written into the Pact file
	State string `json:"providerState,omitempty"`
}

// Given specifies a provider state. Optional.
func (i *InteractionV2) Given(state string) *InteractionV2 {
	i.State = state

	i.interaction.Given(state)

	return i
}
