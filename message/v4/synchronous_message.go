package v4

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/command"
	"github.com/pact-foundation/pact-go/v2/internal/native"
	logging "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
)

// SynchronousPact is a request/response message pact between a consumer
// and provider, built via AddSynchronousMessage.
type SynchronousPact struct {
	config Config

	// Reference to the native rust handle
	mockserver *native.MessageServer
}

// SynchronousMessage contains a req/res message
// It is currently an empty struct to allow future expansion.
type SynchronousMessage struct {
	// TODO: should we pass this in? Probably need to be able to reify the message
	//       in these cases
	Request MessageContents

	// Currently only support a single response, but support may be added for multiple
	// responses to be given in the future
	Response []MessageContents
}

// SynchronousMessageBuilder is a representation of a single, bidirectional message.
type SynchronousMessageBuilder struct{}

// Given specifies a provider state.
func (m *UnconfiguredSynchronousMessageBuilder) Given(state string) *UnconfiguredSynchronousMessageBuilder {
	m.messageHandle.Given(state)

	return &UnconfiguredSynchronousMessageBuilder{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// GivenWithParameter specifies a provider state along with parameters
// used to set it up.
func (m *UnconfiguredSynchronousMessageBuilder) GivenWithParameter(state models.ProviderState) *UnconfiguredSynchronousMessageBuilder {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return &UnconfiguredSynchronousMessageBuilder{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// UnconfiguredSynchronousMessageBuilder builds a single synchronous
// message interaction, starting from AddSynchronousMessage and continuing
// via Given/UsingPlugin/WithRequest.
type UnconfiguredSynchronousMessageBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// AddExternalReference records a reference to an external resource (such as a ticket or
// pull request) against the interaction. References appear under
// comments.references[group][name] in the Pact file. May be called multiple times.
func (m *UnconfiguredSynchronousMessageBuilder) AddExternalReference(group, name, value string) *UnconfiguredSynchronousMessageBuilder {
	m.messageHandle.WithReference(group, name, value)

	return m
}

// UsingPlugin enables a plugin for use in the current test case.
func (m *UnconfiguredSynchronousMessageBuilder) UsingPlugin(config PluginConfig) *SynchronousMessageWithPlugin {
	err := m.pact.mockserver.UsingPlugin(config.Plugin, config.Version)
	if err != nil {
		log.Println("[ERROR] failed to add plugin:", err)
		panic(err)
	}

	return &SynchronousMessageWithPlugin{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// UsingPlugin enables a plugin for use in the current test case.
func (m *SynchronousMessageWithPlugin) UsingPlugin(config PluginConfig) *SynchronousMessageWithPlugin {
	err := m.pact.mockserver.UsingPlugin(config.Plugin, config.Version)
	if err != nil {
		log.Println("[ERROR] failed to add plugin:", err)
		panic(err)
	}

	return m
}

// WithRequest configures the request half of this message via r, then
// moves on to configuring the response.
func (m *UnconfiguredSynchronousMessageBuilder) WithRequest(r RequestBuilderFunc) *SynchronousMessageWithRequest {
	r(&SynchronousMessageWithRequestBuilder{
		messageHandle: m.messageHandle,
		pact:          m.pact,
	})

	return &SynchronousMessageWithRequest{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// SynchronousMessageWithRequest is a message interaction whose request
// has been configured, ready to configure a response via WithResponse.
type SynchronousMessageWithRequest struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// RequestBuilderFunc configures the request half of a synchronous
// message, as passed to UnconfiguredSynchronousMessageBuilder.WithRequest.
type RequestBuilderFunc func(*SynchronousMessageWithRequestBuilder)

// SynchronousMessageWithRequestBuilder configures the request contents
// and metadata of a synchronous message.
type SynchronousMessageWithRequestBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content.
func (m *SynchronousMessageWithRequestBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithRequestMetadata(metadata)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive.
func (m *SynchronousMessageWithRequestBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithContents(native.INTERACTION_PART_REQUEST, contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed.
func (m *SynchronousMessageWithRequestBuilder) WithJSONContent(content any) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithRequestJSONContents(content)

	return m
}

// WithResponse configures the response half of this message via builder,
// then moves on to running the test.
func (m *SynchronousMessageWithRequest) WithResponse(builder ResponseBuilderFunc) *SynchronousMessageWithResponse {
	builder(&SynchronousMessageWithResponseBuilder{
		messageHandle: m.messageHandle,
		pact:          m.pact,
	})

	return &SynchronousMessageWithResponse{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// SynchronousMessageWithResponse is a message interaction whose request
// and response have both been configured, ready to run via ExecuteTest.
type SynchronousMessageWithResponse struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// ResponseBuilderFunc configures the response half of a synchronous
// message, as passed to SynchronousMessageWithRequest.WithResponse.
type ResponseBuilderFunc func(*SynchronousMessageWithResponseBuilder)

// SynchronousMessageWithResponseBuilder configures the response contents
// and metadata of a synchronous message.
type SynchronousMessageWithResponseBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content.
func (m *SynchronousMessageWithResponseBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithResponseMetadata(metadata)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
// May be called multiple times, with each call appeding a new response to the interaction.
func (m *SynchronousMessageWithResponseBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithContents(native.INTERACTION_PART_RESPONSE, contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed.
func (m *SynchronousMessageWithResponseBuilder) WithJSONContent(content any) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithResponseJSONContents(content)

	return m
}

// SynchronousMessageWithPlugin is a message interaction with a plugin
// loaded via UsingPlugin, ready to have its contents set by that plugin.
type SynchronousMessageWithPlugin struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// WithContents sets the request contents of this message from contents,
// interpreted by the loaded plugin according to contentType (e.g. a
// protobuf message type).
func (m *SynchronousMessageWithPlugin) WithContents(contents string, contentType string) *SynchronousMessageWithPluginContents {
	_ = m.messageHandle.WithPluginInteractionContents(native.INTERACTION_PART_REQUEST, contentType, contents)

	return &SynchronousMessageWithPluginContents{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// SynchronousMessageWithPluginContents is a plugin-backed message
// interaction with its contents set, ready to run directly via
// ExecuteTest or to start a transport (e.g. gRPC) first via StartTransport.
type SynchronousMessageWithPluginContents struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful.
func (m *SynchronousMessageWithPluginContents) ExecuteTest(t *testing.T, integrationTest func(m SynchronousMessage) error) error {
	t.Helper()
	defer m.pact.mockserver.CleanupPlugins()
	message, err := getSynchronousMessageWithContents(m.messageHandle)
	if err != nil {
		return err
	}

	err = integrationTest(message)
	if err != nil {
		return err
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return m.pact.mockserver.WritePactFile(m.pact.config.PactDir, false)
}

// StartTransport starts a plugin-provided mock server for transport (e.g.
// "grpc") on address, for tests that need a live connection rather than
// reading the message contents directly.
func (m *SynchronousMessageWithPluginContents) StartTransport(transport string, address string, config map[string][]any) *SynchronousMessageWithTransport {
	port, err := m.pact.mockserver.StartTransport(transport, address, 0, make(map[string][]any))
	if err != nil {
		log.Fatalln("unable to start plugin transport:", err)
	}

	return &SynchronousMessageWithTransport{
		pact:          m.pact,
		messageHandle: m.messageHandle,
		transport: TransportConfig{
			Port:    port,
			Address: address,
		},
	}
}

// SynchronousMessageWithTransport is a plugin-backed message interaction
// with a live transport (e.g. gRPC) running, ready to run via ExecuteTest.
type SynchronousMessageWithTransport struct {
	messageHandle *native.Message
	pact          *SynchronousPact
	transport     TransportConfig
}

// ExecuteTest runs integrationTest against the transport started by
// StartTransport, then verifies the interaction and writes the pact file
// if successful.
func (s *SynchronousMessageWithTransport) ExecuteTest(t *testing.T, integrationTest func(tc TransportConfig, m SynchronousMessage) error) error {
	t.Helper()
	defer s.pact.mockserver.CleanupMockServer(s.transport.Port)
	defer s.pact.mockserver.CleanupPlugins()
	message, err := getSynchronousMessageWithContents(s.messageHandle)
	if err != nil {
		return err
	}

	err = integrationTest(s.transport, message)

	// matched := s.pact.mockserver.MockServerMatched(s.transport.Port)
	// log.Println("MATHED??????????", matched)
	mismatches := s.pact.mockserver.MockServerMismatchedRequests(s.transport.Port)

	if len(mismatches) > 0 {
		return fmt.Errorf("pact validation failed: %+v", mismatches)
	}

	if err != nil {
		return err
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return s.pact.mockserver.WritePactFileForServer(s.transport.Port, s.pact.config.PactDir, false)
}

// PluginConfig identifies a pact_ffi plugin to load via UsingPlugin.
type PluginConfig struct {
	Plugin  string
	Version string
}

// NewSynchronousPact creates a new synchronous (request/response) message
// pact for the consumer/provider pair described by config.
func NewSynchronousPact(config Config) (*SynchronousPact, error) {
	provider := &SynchronousPact{
		config: config,
	}
	provider.validateConfig()

	native.Init(string(logging.LogLevel()))

	return provider, nil
}

// AddSynchronousMessage starts building a new request/response message
// interaction, described by description.
func (m *SynchronousPact) AddSynchronousMessage(description string) *UnconfiguredSynchronousMessageBuilder {
	log.Println("[DEBUG] add sync message")

	message := m.mockserver.NewSyncMessageInteraction(description)

	return &UnconfiguredSynchronousMessageBuilder{
		messageHandle: message,
		pact:          m,
	}
}

// validateConfig applies the defaults for any unset consumer test configuration.
func (m *SynchronousPact) validateConfig() {
	log.Println("[DEBUG] pact synchronous message validate config")
	dir, _ := os.Getwd()

	if m.config.PactDir == "" {
		m.config.PactDir = filepath.Join(dir, "pacts")
	}

	m.mockserver = native.NewMessageServer(m.config.Consumer, m.config.Provider)
	m.mockserver.WithSpecificationVersion(native.SPECIFICATION_VERSION_V4)
	m.mockserver.WithMetadata("pact-go", "version", strings.TrimPrefix(command.Version, "v"))
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful.
func (m *SynchronousMessageWithResponse) ExecuteTest(t *testing.T, integrationTest func(md SynchronousMessage) error) error {
	t.Helper()
	message, err := getSynchronousMessageWithContents(m.messageHandle)
	if err != nil {
		return err
	}

	err = integrationTest(message)
	if err != nil {
		return err
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return m.pact.mockserver.WritePactFile(m.pact.config.PactDir, false)
}

func getSynchronousMessageWithContents(message *native.Message) (SynchronousMessage, error) {
	var m SynchronousMessage

	contents, err := message.GetMessageRequestContents()
	if err != nil {
		//nolint:wrapcheck // native.Message's contents accessors return
		// internal/native's FFI-derived error; the callers of this helper add the
		// "bug in the framework" context.
		return m, err
	}

	responses, err := message.GetMessageResponseContents()
	if err != nil {
		//nolint:wrapcheck // native.Message's contents accessors return
		// internal/native's FFI-derived error; the callers of this helper add the
		// "bug in the framework" context.
		return m, err
	}

	response := make([]MessageContents, len(responses))
	for i, r := range responses {
		response[i] = MessageContents{
			Contents: r,
		}
	}

	return SynchronousMessage{
		Request: MessageContents{
			Contents: contents,
		},
		Response: response,
	}, nil
}
