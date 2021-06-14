package consumer

import "github.com/pact-foundation/pact-go/v2/models"

// InteractionV3 sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type InteractionV3 struct {
	Interaction
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *InteractionV3) Given(state models.ProviderStateV3) *InteractionV3 {
	i.Interaction.interaction.GivenWithParameter(state.Name, state.Parameters)

	return i
}
