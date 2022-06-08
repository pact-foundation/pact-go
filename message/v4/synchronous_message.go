package v4

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
)

type SynchronousPact struct {
	config Config

	// Reference to the native rust handle
	// mockserver *native.MockServer
	mockserver *native.MessageServer
}

// SynchronousMessage contains a req/res message
type SynchronousMessageDetail struct {
	// TODO: what goes here?
	// Message Body
	Content interface{} `json:"contents"`

	// Provider state to be written into the Pact file
	States []models.V3ProviderState `json:"providerStates"`

	// Message metadata
	Metadata matchers.MetadataMatcher `json:"metadata"`

	// Description to be written into the Pact file
	Description string `json:"description"`
}

type SynchronousConsumer func(SynchronousMessageDetail) error

type UnconfiguredSynchronousMessage struct {
}

// AddMessage creates a new asynchronous consumer expectation
func (m *UnconfiguredSynchronousMessage) UsingPlugin(config PluginConfig) *SynchronousMessageWithPlugin {
	// m.Pact.mockserver.UsingPlugin(config.Plugin, config.Version)

	return &SynchronousMessageWithPlugin{}
}

type SynchronousMessageWithRequest struct {
}

type RequestBuilder func(*SynchronousMessageWithRequestBuilder)

// AddMessage creates a new asynchronous consumer expectation
func (m *UnconfiguredSynchronousMessage) WithRequest(r RequestBuilder) *SynchronousMessageWithRequest {
	// m.Pact.mockserver.UsingPlugin(config.Plugin, config.Version)

	return &SynchronousMessageWithRequest{}
}

type SynchronousMessageWithRequestBuilder struct {
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *SynchronousMessageWithRequestBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithRequestBuilder {
	// m.messageHandle.WithMetadata(metadata)

	return m
}

// WithBinaryContent accepts a binary payload
func (m *SynchronousMessageWithRequestBuilder) WithBinaryContent(contentType string, body []byte) *SynchronousMessageWithRequestBuilder {
	// m.messageHandle.WithContents(contentType, body)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *SynchronousMessageWithRequestBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithRequestBuilder {
	// m.messageHandle.WithContents(contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *SynchronousMessageWithRequestBuilder) WithJSONContent(content interface{}) *SynchronousMessageWithRequestBuilder {
	// m.messageHandle.WithJSONContents(content)

	return m
}

// AsType specifies that the content sent through to the
// consumer handler should be sent as the given type
func (m *SynchronousMessageWithRequestBuilder) AsType(t interface{}) *SynchronousMessageWithRequestBuilder {
	// log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	// m.Type = t

	return m
}

// AddMessage creates a new asynchronous consumer expectation
func (m *SynchronousMessageWithRequest) WithResponse(builder ResponseBuilder) *SynchronousMessageWithResponse {
	// m.Pact.mockserver.UsingPlugin(config.Plugin, config.Version)

	return &SynchronousMessageWithResponse{}
}

type SynchronousMessageWithResponse struct {
}

type ResponseBuilder func(*SynchronousMessageWithResponseBuilder)

type SynchronousMessageWithResponseBuilder struct {
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *SynchronousMessageWithResponseBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithResponseBuilder {
	// m.messageHandle.WithMetadata(metadata)

	return m
}

// WithBinaryContent accepts a binary payload
func (m *SynchronousMessageWithResponseBuilder) WithBinaryContent(contentType string, body []byte) *SynchronousMessageWithResponseBuilder {
	// m.messageHandle.WithContents(contentType, body)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *SynchronousMessageWithResponseBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithResponseBuilder {
	// m.messageHandle.WithContents(contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *SynchronousMessageWithResponseBuilder) WithJSONContent(content interface{}) *SynchronousMessageWithResponseBuilder {
	// m.messageHandle.WithJSONContents(content)

	return m
}

// AsType specifies that the content sent through to the
// consumer handler should be sent as the given type
func (m *SynchronousMessageWithResponseBuilder) AsType(t interface{}) *SynchronousMessageWithResponseBuilder {
	// log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	// m.Type = t

	return m
}

type SynchronousMessageWithPlugin struct {
}

func (s *SynchronousMessageWithPlugin) WithContents() *SynchronousMessageWithPluginContents {
	return &SynchronousMessageWithPluginContents{}
}

type SynchronousMessageWithPluginContents struct {
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (m *SynchronousMessageWithPluginContents) ExecuteTest(t *testing.T, integrationTest func(TransportConfig) error) error {
	return nil
}

func (s *SynchronousMessageWithPluginContents) StartTransport() *SynchronousMessageWithTransport {
	return &SynchronousMessageWithTransport{}
}

type SynchronousMessageWithPluginTransport struct {
}

func (s *SynchronousMessageWithPluginTransport) StartTransport() *SynchronousMessageWithTransport {
	return &SynchronousMessageWithTransport{}
}

type SynchronousMessageWithTransport struct {
}

func (s *SynchronousMessageWithTransport) ExecuteTest(t *testing.T, integrationTest func(TransportConfig) error) error {
	return nil
}

// SynchronousMessage is a representation of a single, bidirectional message
type SynchronousMessage struct {
	messageHandle *native.Message
	Pact          *SynchronousPact

	// Type to Marshal content into when sending back to the consumer
	// Defaults to interface{}
	Type interface{}

	// The handler for this message
	handler SynchronousConsumer
}

// Given specifies a provider state. Optional.
func (m *SynchronousMessage) Given(state string) *UnconfiguredSynchronousMessage {
	m.messageHandle.Given(state)

	return &UnconfiguredSynchronousMessage{}
}

// Given specifies a provider state. Optional.
func (m *SynchronousMessage) GivenWithParameter(state models.V3ProviderState) *UnconfiguredSynchronousMessage {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return &UnconfiguredSynchronousMessage{}
}

// The function that will consume the message
// func (m *SynchronousMessage) ConsumedBy(handler SynchronousConsumer) *SynchronousMessage {
// 	m.handler = handler

// 	return m
// }

// The function that will consume the message
func (m *SynchronousMessage) Verify(t *testing.T) error {
	return m.Pact.Verify(t, m, m.handler)
}

type PluginConfig struct {
	Plugin  string
	Version string
}

func NewSynchronousPact(config Config) (*SynchronousPact, error) {
	provider := &SynchronousPact{
		config: config,
	}
	err := provider.validateConfig()

	if err != nil {
		return nil, err
	}

	native.Init()

	return provider, err
}

// validateConfig validates the configuration for the consumer test
func (m *SynchronousPact) validateConfig() error {
	log.Println("[DEBUG] pact message validate config")
	dir, _ := os.Getwd()

	if m.config.PactDir == "" {
		m.config.PactDir = filepath.Join(dir, "pacts")
	}

	m.mockserver = native.NewMessageServer(m.config.Consumer, m.config.Provider)

	return nil
}

func (m *SynchronousPact) AddSynchronousMessage(description string) *SynchronousMessage {
	log.Println("[DEBUG] add sync message")

	message := m.mockserver.NewSyncMessageInteraction(description)

	return &SynchronousMessage{
		messageHandle: message,
		Pact:          m,
	}
}

// TODO
// func (m *Pact) AddAsynchronousMessage(description string) *AsynchronousMessage {
// 	log.Println("[DEBUG] add async message")

// 	message := m.mockserver.NewAsyncMessageInteraction("")

// 	return &SynchronousMessage{
// 		messageHandle: message,
// 		Pact:          m,
// 	}
// }

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (m *SynchronousMessageWithResponse) ExecuteTest(t *testing.T, integrationTest func(TransportConfig) error) error {
	// log.Println("[DEBUG] pact verify")

	// var err error
	// if p.config.AllowedMockServerPorts != "" && p.config.Port <= 0 {
	// 	p.config.Port, err = utils.FindPortInRange(p.config.AllowedMockServerPorts)
	// } else if p.config.Port <= 0 {
	// 	p.config.Port, err = 0, nil
	// }

	// if err != nil {
	// 	return fmt.Errorf("error: unable to find free port, mock server will fail to start")
	// }

	// p.config.Port, err = p.mockserver.Start(fmt.Sprintf("%s:%d", p.config.Host, p.config.Port), p.config.TLS)
	// defer p.reset()
	// if err != nil {
	// 	return err
	// }

	// // Run the integration test
	// err = integrationTest(MockServerConfig{
	// 	Port:      p.config.Port,
	// 	Host:      p.config.Host,
	// 	TLSConfig: GetTLSConfigForTLSMockServer(),
	// })

	// res, mismatches := p.mockserver.Verify(p.config.Port, p.config.PactDir)
	// p.displayMismatches(t, mismatches)

	// if err != nil {
	// 	return err
	// }

	// if !res {
	// 	return fmt.Errorf("pact validation failed: %+v %+v", res, mismatches)
	// }

	// if len(mismatches) > 0 {
	// 	return fmt.Errorf("pact validation failed: %+v", mismatches)
	// }

	// return p.writePact()
	return nil
}

// Verifymessage.AsynchronousConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
//
// A Message Consumer is analagous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (m *SynchronousPact) verifySynchronousConsumerRaw(message *SynchronousMessage, handler SynchronousConsumer) error {
	return nil
}

// Verifymessage.AsynchronousConsumer is a test convience function for Verifymessage.AsynchronousConsumerRaw,
// accepting an instance of `*testing.T`
func (m *SynchronousPact) Verify(t *testing.T, message *SynchronousMessage, handler SynchronousConsumer) error {
	err := m.verifySynchronousConsumerRaw(message, handler)

	if err != nil {
		t.Errorf("Verifymessage.AsynchronousConsumer failed: %v", err)
	}

	return err
}
