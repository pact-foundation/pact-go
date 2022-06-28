package consumer

import (
	"log"

	"github.com/pact-foundation/pact-go/v2/models"
)

// V3HTTPMockProvider is the entrypoint for V3 http consumer tests
// This object is not thread safe
type V3HTTPMockProvider struct {
	*httpMockProvider
}

// NewV3Pact configures a new V3 HTTP Mock Provider for consumer tests
func NewV3Pact(config MockHTTPProviderConfig) (*V3HTTPMockProvider, error) {
	provider := &V3HTTPMockProvider{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: models.V3,
		},
	}
	err := provider.configure()

	if err != nil {
		return nil, err
	}

	return provider, err
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *V3HTTPMockProvider) AddInteraction() *V3Interaction {
	log.Println("[DEBUG] pact add v3 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &V3Interaction{
		Interaction: Interaction{
			specificationVersion: models.V3,
			interaction:          interaction,
		},
	}

	return i
}

// V3Interaction sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type V3Interaction struct {
	Interaction
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *V3Interaction) Given(state string) *V3Interaction {
	i.Interaction.interaction.Given(state)

	return i
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *V3Interaction) GivenWithParameter(state models.V3ProviderState) *V3Interaction {
	if len(state.Parameters) > 0 {
		i.Interaction.interaction.GivenWithParameter(state.Name, state.Parameters)
	} else {
		i.Interaction.interaction.Given(state.Name)
	}

	return i
}
