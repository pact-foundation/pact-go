package v3

import (
	"log"
)

// HTTPMockProviderV2 is the entrypoint for V3 http consumer tests
type HTTPMockProviderV2 struct {
	*httpMockProvider
}

// NewHTTPMockProviderV2 configures a new V2 HTTP Mock Provider for consumer tests
func NewHTTPMockProviderV2(config MockHTTPProviderConfigV2) (*HTTPMockProviderV2, error) {
	provider := &HTTPMockProviderV2{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: V2,
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
	i := &InteractionV2{}
	p.httpMockProvider.v2Interactions = append(p.httpMockProvider.v2Interactions, i)
	return i
}

// SetMatchingConfig allows specific contract file serialisation adjustments
func (p *HTTPMockProviderV2) SetMatchingConfig(config PactSerialisationOptionsV2) *HTTPMockProviderV2 {
	p.config.matchingConfig = config

	return p
}
