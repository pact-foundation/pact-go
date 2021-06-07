package consumer

import (
	"log"

	"github.com/pact-foundation/pact-go/v2/models"
)

// HTTPMockProviderV2 is the entrypoint for V3 http consumer tests
//
// This object is not thread safe
type HTTPMockProviderV2 struct {
	*httpMockProvider
}

// NewV2Pact configures a new V2 HTTP Mock Provider for consumer tests
func NewV2Pact(config MockHTTPProviderConfig) (*HTTPMockProviderV2, error) {
	provider := &HTTPMockProviderV2{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: models.V2,
		},
	}
	err := provider.validateConfig()

	if err != nil {
		return nil, err
	}

	return provider, err
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *HTTPMockProviderV2) AddInteraction() *InteractionV2 {
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
