package consumer

import (
	"log"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
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

// AddInteraction to the pact
func (p *V3HTTPMockProvider) AddInteraction() *V3UnconfiguredInteraction {
	log.Println("[DEBUG] pact add V3 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &V3UnconfiguredInteraction{
		interaction: &Interaction{
			specificationVersion: models.V3,
			interaction:          interaction,
		},
		provider: p,
	}

	return i
}

type V3UnconfiguredInteraction struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *V3UnconfiguredInteraction) Given(state string) *V3UnconfiguredInteraction {
	i.interaction.interaction.Given(state)

	return i
}

// GivenWithParameter specifies a provider state with parameters, may be called multiple times. Optional.
func (i *V3UnconfiguredInteraction) GivenWithParameter(state models.ProviderState) *V3UnconfiguredInteraction {
	if len(state.Parameters) > 0 {
		i.interaction.interaction.GivenWithParameter(state.Name, state.Parameters)
	} else {
		i.interaction.interaction.Given(state.Name)
	}

	return i
}

type V3InteractionWithRequest struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

type V3RequestBuilderFunc func(*V3RequestBuilder)

type V3RequestBuilder struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *V3UnconfiguredInteraction) UponReceiving(description string) *V3UnconfiguredInteraction {
	i.interaction.interaction.UponReceiving(description)

	return i
}

// WithRequest provides a builder for the expected request
func (i *V3UnconfiguredInteraction) WithCompleteRequest(request Request) *V3InteractionWithCompleteRequest {
	i.interaction.WithCompleteRequest(request)

	return &V3InteractionWithCompleteRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V3InteractionWithCompleteRequest struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

// WithRequest provides a builder for the expected request
func (i *V3InteractionWithCompleteRequest) WithCompleteResponse(response Response) *V3InteractionWithResponse {
	i.interaction.WithCompleteResponse(response)

	return &V3InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WithRequest provides a builder for the expected request
func (i *V3UnconfiguredInteraction) WithRequest(method Method, path string, builders ...V3RequestBuilderFunc) *V3InteractionWithRequest {
	return i.WithRequestPathMatcher(method, matchers.String(path), builders...)
}

// WithRequestPathMatcher allows a matcher in the expected request path
func (i *V3UnconfiguredInteraction) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...V3RequestBuilderFunc) *V3InteractionWithRequest {
	i.interaction.interaction.WithRequest(string(method), path)

	for _, builder := range builders {
		builder(&V3RequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V3InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// Query specifies any query string on the expect request
func (i *V3RequestBuilder) Query(key string, values ...matchers.Matcher) *V3RequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Header adds a header to the expected request
func (i *V3RequestBuilder) Header(key string, values ...matchers.Matcher) *V3RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected request
func (i *V3RequestBuilder) Headers(headers matchers.HeadersMatcher) *V3RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected request
func (i *V3RequestBuilder) JSONBody(body interface{}) *V3RequestBuilder {
	// TODO: Don't like panic, but not sure if there is a better builder experience?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] request body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}

	i.interaction.interaction.WithJSONRequestBody(body)

	return i
}

// BinaryBody adds a binary body to the expected request
func (i *V3RequestBuilder) BinaryBody(body []byte) *V3RequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected request
func (i *V3RequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V3RequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V3RequestBuilder) Body(contentType string, body []byte) *V3RequestBuilder {
	// Check if someone tried to add an object as a string representation
	// as per original allowed implementation, e.g.
	// { "foo": "bar", "baz": like("bat") }
	if utils.IsJSONFormattedObject(string(body)) {
		log.Println("[WARN] request body appears to be a JSON formatted object, " +
			"no matching will occur. Support for structured strings has been" +
			"deprecated as of 0.13.0. Please use the JSON() method instead")
	}

	i.interaction.interaction.WithRequestBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V3RequestBuilder) BodyMatch(body interface{}) *V3RequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// WillRespondWith sets the expected status and provides a response builder
func (i *V3InteractionWithRequest) WillRespondWith(status int, builders ...V3ResponseBuilderFunc) *V3InteractionWithResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {

		builder(&V3ResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V3InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V3ResponseBuilderFunc func(*V3ResponseBuilder)

type V3ResponseBuilder struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

type V3InteractionWithResponse struct {
	interaction *Interaction
	provider    *V3HTTPMockProvider
}

// Header adds a header to the expected response
func (i *V3ResponseBuilder) Header(key string, values ...matchers.Matcher) *V3ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected response
func (i *V3ResponseBuilder) Headers(headers matchers.HeadersMatcher) *V3ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected response
func (i *V3ResponseBuilder) JSONBody(body interface{}) *V3ResponseBuilder {
	// TODO: Don't like panic, how to build a better builder here - nil return + log?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] response body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}
	i.interaction.interaction.WithJSONResponseBody(body)

	return i
}

// BinaryBody adds a binary body to the expected response
func (i *V3ResponseBuilder) BinaryBody(body []byte) *V3ResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected response
func (i *V3ResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V3ResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V3ResponseBuilder) Body(contentType string, body []byte) *V3ResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V3ResponseBuilder) BodyMatch(body interface{}) *V3ResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V3InteractionWithResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}
