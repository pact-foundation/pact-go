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
func (p *V2HTTPMockProvider) AddInteraction() *InteractionV2 {
	log.Println("[DEBUG] pact add v2 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &InteractionV2{
		Interaction: Interaction{
			specificationVersion: models.V2,
			interaction:          interaction,
		},
	}

	return i
}

// InteractionV2 sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type InteractionV2 struct {
	Interaction
}

// Given specifies a provider state. Optional.
func (i *InteractionV2) Given(state string) *InteractionV2 {
	i.interaction.Given(state)

	return i
}
