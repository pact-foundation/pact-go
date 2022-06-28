package consumer

import (
	"log"

	"github.com/pact-foundation/pact-go/v2/models"
)

// V2HTTPMockProvider is the entrypoint for V3 http consumer tests
//
// This object is not thread safe
type V2HTTPMockProvider struct {
	*httpMockProvider
}

// NewV2Pact configures a new V2 HTTP Mock Provider for consumer tests
func NewV2Pact(config MockHTTPProviderConfig) (*V2HTTPMockProvider, error) {
	provider := &V2HTTPMockProvider{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: models.V2,
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
func (p *V2HTTPMockProvider) AddInteraction() *V2Interaction {
	log.Println("[DEBUG] pact add v2 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &V2Interaction{
		Interaction: Interaction{
			specificationVersion: models.V2,
			interaction:          interaction,
		},
	}

	return i
}

// V2Interaction sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type V2Interaction struct {
	Interaction
}

// Given specifies a provider state. Optional.
func (i *V2Interaction) Given(state string) *V2Interaction {
	i.interaction.Given(state)

	return i
}
