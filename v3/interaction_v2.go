package v3

// InteractionV2 is the main implementation of the Pact interface.
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
