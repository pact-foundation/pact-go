package consumer

import (
	"log"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// V4HTTPMockProvider is the entrypoint for V4 http consumer tests
// This object is not thread safe
type V4HTTPMockProvider struct {
	*httpMockProvider
}

// NewV4Pact configures a new V4 HTTP Mock Provider for consumer tests
func NewV4Pact(config MockHTTPProviderConfig) (*V4HTTPMockProvider, error) {
	provider := &V4HTTPMockProvider{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: models.V4,
		},
	}
	err := provider.configure()

	if err != nil {
		return nil, err
	}

	return provider, err
}

// AddInteraction to the pact
func (p *V4HTTPMockProvider) AddInteraction() *V4UnconfiguredInteraction {
	log.Println("[DEBUG] pact add V4 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &V4UnconfiguredInteraction{
		interaction: &Interaction{
			specificationVersion: models.V4,
			interaction:          interaction,
		},
		provider: p,
	}

	return i
}

type V4UnconfiguredInteraction struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *V4UnconfiguredInteraction) Given(state string) *V4UnconfiguredInteraction {
	i.interaction.interaction.Given(state)

	return i
}

// GivenWithParameter specifies a provider state with parameters, may be called multiple times. Optional.
func (i *V4UnconfiguredInteraction) GivenWithParameter(state models.ProviderState) *V4UnconfiguredInteraction {
	if len(state.Parameters) > 0 {
		i.interaction.interaction.GivenWithParameter(state.Name, state.Parameters)
	} else {
		i.interaction.interaction.Given(state.Name)
	}

	return i
}

type V4InteractionWithRequest struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

type V4RequestBuilderFunc func(*V4RequestBuilder)

type V4RequestBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *V4UnconfiguredInteraction) UponReceiving(description string) *V4UnconfiguredInteraction {
	i.interaction.interaction.UponReceiving(description)

	return i
}

// WithRequest provides a builder for the expected request
func (i *V4UnconfiguredInteraction) WithCompleteRequest(request Request) *V4InteractionWithCompleteRequest {
	i.interaction.WithCompleteRequest(request)

	return &V4InteractionWithCompleteRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V4InteractionWithCompleteRequest struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// WithRequest provides a builder for the expected request
func (i *V4InteractionWithCompleteRequest) WithCompleteResponse(response Response) *V4InteractionWithResponse {
	i.interaction.WithCompleteResponse(response)

	return &V4InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WithRequest provides a builder for the expected request
func (i *V4UnconfiguredInteraction) WithRequest(method Method, path string, builders ...V4RequestBuilderFunc) *V4InteractionWithRequest {
	return i.WithRequestPathMatcher(method, matchers.String(path), builders...)
}

// WithRequestPathMatcher allows a matcher in the expected request path
func (i *V4UnconfiguredInteraction) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...V4RequestBuilderFunc) *V4InteractionWithRequest {
	i.interaction.interaction.WithRequest(string(method), path)

	for _, builder := range builders {
		builder(&V4RequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// Query specifies any query string on the expect request
func (i *V4RequestBuilder) Query(key string, values ...matchers.Matcher) *V4RequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Header adds a header to the expected request
func (i *V4RequestBuilder) Header(key string, values ...matchers.Matcher) *V4RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected request
func (i *V4RequestBuilder) Headers(headers matchers.HeadersMatcher) *V4RequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected request
func (i *V4RequestBuilder) JSONBody(body interface{}) *V4RequestBuilder {
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
func (i *V4RequestBuilder) BinaryBody(body []byte) *V4RequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected request
func (i *V4RequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4RequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V4RequestBuilder) Body(contentType string, body []byte) *V4RequestBuilder {
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
func (i *V4RequestBuilder) BodyMatch(body interface{}) *V4RequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// WillRespondWith sets the expected status and provides a response builder
func (i *V4InteractionWithRequest) WillRespondWith(status int, builders ...V4ResponseBuilderFunc) *V4InteractionWithResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {

		builder(&V4ResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V4ResponseBuilderFunc func(*V4ResponseBuilder)

type V4ResponseBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

type V4InteractionWithResponse struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// Header adds a header to the expected response
func (i *V4ResponseBuilder) Header(key string, values ...matchers.Matcher) *V4ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected response
func (i *V4ResponseBuilder) Headers(headers matchers.HeadersMatcher) *V4ResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected response
func (i *V4ResponseBuilder) JSONBody(body interface{}) *V4ResponseBuilder {
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
func (i *V4ResponseBuilder) BinaryBody(body []byte) *V4ResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected response
func (i *V4ResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4ResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V4ResponseBuilder) Body(contentType string, body []byte) *V4ResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V4ResponseBuilder) BodyMatch(body interface{}) *V4ResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V4InteractionWithResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}

////////////
// Plugin //
////////////

type PluginConfig struct {
	Plugin  string
	Version string
}

type V4InteractionWithPlugin struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// UsingPlugin specifies the current interaction relies on one or more plugins for operation
// If the plugin is not correctly installed, this method will terminate the test immediately with a non-zero status
func (i *V4UnconfiguredInteraction) UsingPlugin(config PluginConfig) *V4InteractionWithPlugin {
	res := i.provider.mockserver.UsingPlugin(config.Plugin, config.Version)
	if res != nil {
		log.Fatal("pact setup failed:", res)
	}

	return &V4InteractionWithPlugin{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// UsingPlugin specifies the current interaction relies on one or more plugins for operation
// If the plugin is not correctly installed, this method will terminate the test immediately with a non-zero status
func (i *V4InteractionWithPlugin) UsingPlugin(config PluginConfig) *V4InteractionWithPlugin {
	res := i.provider.mockserver.UsingPlugin(config.Plugin, config.Version)
	if res != nil {
		log.Fatal("pact setup failed:", res)
	}

	return i
}

type V4InteractionWithPluginRequest struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

type PluginRequestBuilderFunc func(*V4InteractionWithPluginRequestBuilder)

type V4InteractionWithPluginRequestBuilder struct {
	interaction *Interaction
}

// WithRequest provides a builder for the expected request
func (i *V4InteractionWithPlugin) WithRequest(method Method, path string, builders ...PluginRequestBuilderFunc) *V4InteractionWithPluginRequest {
	i.interaction.interaction.WithRequest(string(method), matchers.String(path))

	for _, builder := range builders {
		builder(&V4InteractionWithPluginRequestBuilder{
			interaction: i.interaction,
		})
	}

	return &V4InteractionWithPluginRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WithRequestPathMatcher allows a matcher in the expected request path
func (i *V4InteractionWithPlugin) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...PluginRequestBuilderFunc) *V4InteractionWithPluginRequest {
	i.interaction.interaction.WithRequest(string(method), path)

	for _, builder := range builders {
		builder(&V4InteractionWithPluginRequestBuilder{
			interaction: i.interaction,
		})
	}

	return &V4InteractionWithPluginRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WillResponseWithContent provides a builder for the expected response
func (i *V4InteractionWithPluginRequest) WillRespondWith(status int, builders ...PluginResponseBuilderFunc) *V4InteractionWithPluginResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {
		builder(&V4InteractionWithPluginResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithPluginResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type PluginResponseBuilderFunc func(*V4InteractionWithPluginResponseBuilder)

type V4InteractionWithPluginResponseBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

type V4InteractionWithPluginResponse struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V4InteractionWithPluginResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}

// Query specifies any query string on the expect request
func (i *V4InteractionWithPluginRequestBuilder) Query(key string, values ...matchers.Matcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Header adds a header to the expected request
func (i *V4InteractionWithPluginRequestBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected request
func (i *V4InteractionWithPluginRequestBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// PluginContents configures a plugin. This may be called once per plugin registered.
func (i *V4InteractionWithPluginRequestBuilder) PluginContents(contentType string, contents string) *V4InteractionWithPluginRequestBuilder {
	err := i.interaction.interaction.WithPluginInteractionContents(native.INTERACTION_PART_REQUEST, contentType, contents)
	if err != nil {
		log.Println("[ERROR] failed to get plugin content for interaction:", err)
		panic(err)
	}

	return i
}

// JSONBody adds a JSON body to the expected request
func (i *V4InteractionWithPluginRequestBuilder) JSONBody(body interface{}) *V4InteractionWithPluginRequestBuilder {
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
func (i *V4InteractionWithPluginRequestBuilder) BinaryBody(body []byte) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected request
func (i *V4InteractionWithPluginRequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V4InteractionWithPluginRequestBuilder) Body(contentType string, body []byte) *V4InteractionWithPluginRequestBuilder {
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
func (i *V4InteractionWithPluginRequestBuilder) BodyMatch(body interface{}) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// Header adds a header to the expected response
func (i *V4InteractionWithPluginResponseBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected response
func (i *V4InteractionWithPluginResponseBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// PluginContents configures a plugin. This may be called once per plugin registered.
func (i *V4InteractionWithPluginResponseBuilder) PluginContents(contentType string, contents string) *V4InteractionWithPluginResponseBuilder {
	err := i.interaction.interaction.WithPluginInteractionContents(native.INTERACTION_PART_RESPONSE, contentType, contents)
	if err != nil {
		log.Println("[ERROR] failed to get plugin content for interaction:", err)
		panic(err)
	}

	return i
}

// JSONBody adds a JSON body to the expected response
func (i *V4InteractionWithPluginResponseBuilder) JSONBody(body interface{}) *V4InteractionWithPluginResponseBuilder {
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
func (i *V4InteractionWithPluginResponseBuilder) BinaryBody(body []byte) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected response
func (i *V4InteractionWithPluginResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V4InteractionWithPluginResponseBuilder) Body(contentType string, body []byte) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V4InteractionWithPluginResponseBuilder) BodyMatch(body interface{}) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}
