package consumer

import (
	"log"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// V2HTTPMockProvider is the entrypoint for V2 http consumer tests
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

// AddInteraction to the pact
func (p *V2HTTPMockProvider) AddInteraction() *V2UnconfiguredInteraction {
	log.Println("[DEBUG] pact add V2 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &V2UnconfiguredInteraction{
		interaction: &Interaction{
			specificationVersion: models.V2,
			interaction:          interaction,
		},
		provider: p,
	}

	return i
}

type V2UnconfiguredInteraction struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *V2UnconfiguredInteraction) Given(state string) *V2UnconfiguredInteraction {
	i.interaction.interaction.Given(state)

	return i
}

type V2InteractionWithRequest struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

type V2RequestBuilderFunc func(*V2RequestBuilder)

type V2RequestBuilder struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *V2UnconfiguredInteraction) UponReceiving(description string) *V2UnconfiguredInteraction {
	i.interaction.interaction.UponReceiving(description)

	return i
}

// WithRequest provides a builder for the expected request
func (i *V2UnconfiguredInteraction) WithCompleteRequest(request Request) *V2InteractionWithCompleteRequest {
	i.interaction.WithCompleteRequest(request)

	return &V2InteractionWithCompleteRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V2InteractionWithCompleteRequest struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// WithRequest provides a builder for the expected request
func (i *V2InteractionWithCompleteRequest) WithCompleteResponse(response Response) *V2InteractionWithResponse {
	i.interaction.WithCompleteResponse(response)

	return &V2InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WithRequest provides a builder for the expected request
func (i *V2UnconfiguredInteraction) WithRequest(method Method, path string, builders ...V2RequestBuilderFunc) *V2InteractionWithRequest {
	return i.WithRequestPathMatcher(method, matchers.String(path), builders...)
}

// WithRequestPathMatcher allows a matcher in the expected request path
func (i *V2UnconfiguredInteraction) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...V2RequestBuilderFunc) *V2InteractionWithRequest {
	i.interaction.interaction.WithRequest(string(method), path)

	for _, builder := range builders {
		builder(&V2RequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V2InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// Query specifies any query string on the expect request
func (i *V2RequestBuilder) Query(key string, values ...matchers.Matcher) *V2RequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Header adds a header to the expected request
func (i *V2RequestBuilder) Header(key string, values ...matchers.Matcher) *V2RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected request
func (i *V2RequestBuilder) Headers(headers matchers.HeadersMatcher) *V2RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected request
func (i *V2RequestBuilder) JSONBody(body interface{}) *V2RequestBuilder {
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
func (i *V2RequestBuilder) BinaryBody(body []byte) *V2RequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected request
func (i *V2RequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V2RequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V2RequestBuilder) Body(contentType string, body []byte) *V2RequestBuilder {
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
func (i *V2RequestBuilder) BodyMatch(body interface{}) *V2RequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// WillRespondWith sets the expected status and provides a response builder
func (i *V2InteractionWithRequest) WillRespondWith(status int, builders ...V2ResponseBuilderFunc) *V2InteractionWithResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {

		builder(&V2ResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V2InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V2ResponseBuilderFunc func(*V2ResponseBuilder)

type V2ResponseBuilder struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

type V2InteractionWithResponse struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// Header adds a header to the expected response
func (i *V2ResponseBuilder) Header(key string, values ...matchers.Matcher) *V2ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected response
func (i *V2ResponseBuilder) Headers(headers matchers.HeadersMatcher) *V2ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected response
func (i *V2ResponseBuilder) JSONBody(body interface{}) *V2ResponseBuilder {
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
func (i *V2ResponseBuilder) BinaryBody(body []byte) *V2ResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected response
func (i *V2ResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V2ResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V2ResponseBuilder) Body(contentType string, body []byte) *V2ResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V2ResponseBuilder) BodyMatch(body interface{}) *V2ResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V2InteractionWithResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}
