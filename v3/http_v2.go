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

// TODO: this needs to be plumbed into the new native interface
// We'll need a reference to the underlying object so that we can cross the FFI boundary
// with each additional modification to the Interaction object
func (p *HTTPMockProviderV2) AddInteraction() *InteractionV2 {
	log.Println("[DEBUG] pact add v2 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &InteractionV2{
		Interaction: Interaction{
			specificationVersion: V2,
			interaction:          interaction,
		},
	}

	p.httpMockProvider.v2Interactions = append(p.httpMockProvider.v2Interactions, i)

	return i
}

// SetMatchingConfig allows specific contract file serialisation adjustments
// TODO: review if this is even used now we've moved to FFI
// func (p *HTTPMockProviderV2) SetMatchingConfig(config PactSerialisationOptionsV2) *HTTPMockProviderV2 {
// 	p.config.matchingConfig = config

// 	return p
// }
