package consumer

import (
	"log"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// V4HTTPMockProvider is the entrypoint for V3 http consumer tests
// This object is not thread safe
type V4HTTPMockProvider struct {
	*httpMockProvider
}

// NewV4Pact configures a new V3 HTTP Mock Provider for consumer tests
func NewV4Pact(config MockHTTPProviderConfig) (*V4HTTPMockProvider, error) {
	provider := &V4HTTPMockProvider{
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
func (p *V4HTTPMockProvider) AddInteraction() *UnconfiguredV4Interaction {
	log.Println("[DEBUG] pact add v3 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &UnconfiguredV4Interaction{
		interaction: &Interaction{
			specificationVersion: models.V4,
			interaction:          interaction,
		},
		provider: p,
	}

	return i
}

// V4Interaction sets up an expected request/response on a mock server
// and is replayed on the provider side for verification
type V4Interaction struct {
}

type UnconfiguredV4Interaction struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *UnconfiguredV4Interaction) Given(state string) *UnconfiguredV4Interaction {
	i.interaction.interaction.Given(state)

	return i
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *UnconfiguredV4Interaction) GivenWithParameter(state models.V3ProviderState) *UnconfiguredV4Interaction {
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

type RequestBuilder func(*V4InteractionWithRequestBuilder)

type V4InteractionWithRequestBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *UnconfiguredV4Interaction) UponReceiving(description string) *UnconfiguredV4Interaction {
	i.interaction.UponReceiving(description)

	return i
}

// WithRequest provides a builder for the request interface
// It specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *UnconfiguredV4Interaction) WithRequest(method Method, path string, builders ...RequestBuilder) *V4InteractionWithRequest {
	i.interaction.WithRequest(method, matchers.String(path))

	for _, builder := range builders {
		builder(&V4InteractionWithRequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}
func (i *UnconfiguredV4Interaction) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...RequestBuilder) *V4InteractionWithRequest {
	i.interaction.WithRequest(method, path)

	for _, builder := range builders {
		builder(&V4InteractionWithRequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

func (i *V4InteractionWithRequestBuilder) Query(key string, values ...matchers.Matcher) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithRequestBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithRequestBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *V4InteractionWithRequestBuilder) JSONBody(body interface{}) *V4InteractionWithRequestBuilder {
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

func (i *V4InteractionWithRequestBuilder) BinaryBody(body []byte) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

func (i *V4InteractionWithRequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

func (i *V4InteractionWithRequestBuilder) Body(contentType string, body []byte) *V4InteractionWithRequestBuilder {
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

func (i *V4InteractionWithRequestBuilder) BodyMatch(body interface{}) *V4InteractionWithRequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// WithRequest provides a builder for the request interface
func (i *V4InteractionWithRequest) WillRespondWith(status int, builders ...ResponseBuilder) *V4InteractionWithResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {

		builder(&V4InteractionWithResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V4InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type ResponseBuilder func(*V4InteractionWithResponseBuilder)

type V4InteractionWithResponseBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

type V4InteractionWithResponse struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

func (i *V4InteractionWithResponseBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithResponseBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *V4InteractionWithResponseBuilder) JSONBody(body interface{}) *V4InteractionWithResponseBuilder {
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

func (i *V4InteractionWithResponseBuilder) BinaryBody(body []byte) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

func (i *V4InteractionWithResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

func (i *V4InteractionWithResponseBuilder) Body(contentType string, body []byte) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

func (i *V4InteractionWithResponseBuilder) BodyMatch(body interface{}) *V4InteractionWithResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V4InteractionWithResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}

////////////
// Plugin //
///////////

type PluginConfig struct {
	Plugin  string
	Version string
}

type V4InteractionWithPlugin struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *UnconfiguredV4Interaction) UsingPlugin(config PluginConfig) *V4InteractionWithPlugin {
	res := i.provider.mockserver.UsingPlugin(config.Plugin, config.Version)
	if res != nil {
		log.Fatal("pact setup failed:", res)
	}

	return &V4InteractionWithPlugin{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
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

type PluginRequestBuilder func(*V4InteractionWithPluginRequestBuilder)

type V4InteractionWithPluginRequestBuilder struct {
	interaction *Interaction
	provider    *V4HTTPMockProvider
}

// WithRequest provides a builder for the request interface
func (i *V4InteractionWithPlugin) WithRequest(method Method, path string, builders ...PluginRequestBuilder) *V4InteractionWithPluginRequest {
	i.interaction.WithRequest(method, matchers.String(path))

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

// WithRequest provides a builder for the request interface
func (i *V4InteractionWithPlugin) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...PluginRequestBuilder) *V4InteractionWithPluginRequest {
	i.interaction.WithRequest(method, path)

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

// WithRequest provides a builder for the request interface
func (i *V4InteractionWithPluginRequest) WillRespondWith(status int, builders ...PluginResponseBuilder) *V4InteractionWithPluginResponse {
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

type PluginResponseBuilder func(*V4InteractionWithPluginResponseBuilder)

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

func (i *V4InteractionWithPluginRequestBuilder) Query(key string, values ...matchers.Matcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithPluginRequestBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithPluginRequestBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *V4InteractionWithPluginRequestBuilder) PluginContents(contents string, contentType string) *V4InteractionWithPluginRequestBuilder {

	return i
}

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

func (i *V4InteractionWithPluginRequestBuilder) BinaryBody(body []byte) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

func (i *V4InteractionWithPluginRequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

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

func (i *V4InteractionWithPluginRequestBuilder) BodyMatch(body interface{}) *V4InteractionWithPluginRequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) Header(key string, values ...matchers.Matcher) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) Headers(headers matchers.HeadersMatcher) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) PluginContents(contents string, contentType string) *V4InteractionWithPluginResponseBuilder {

	return i
}

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

func (i *V4InteractionWithPluginResponseBuilder) BinaryBody(body []byte) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) Body(contentType string, body []byte) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

func (i *V4InteractionWithPluginResponseBuilder) BodyMatch(body interface{}) *V4InteractionWithPluginResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}
