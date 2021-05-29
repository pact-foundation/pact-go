package v3

// ProviderStateV3 allows parameters and a description to be passed to the verification process
type ProviderStateV3 struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"params,omitempty"`
}

// ProviderStateV3Response may return values in the state setup
// for the "value from provider state" feature
type ProviderStateV3Response map[string]interface{}

// InteractionV3 sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type InteractionV3 struct {
	Interaction
}

// Given specifies a provider state. Optional.
func (i *InteractionV3) Given(state ProviderStateV3) *InteractionV3 {
	i.Interaction.interaction.GivenWithParameter(state.Name, state.Parameters)

	return i
}
