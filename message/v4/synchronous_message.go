package v4

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/models"
)

type SynchronousPact struct {
	config Config

	// Reference to the native rust handle
	mockserver *native.MessageServer
}

// SynchronousMessage contains a req/res message
// It is currently an empty struct to allow future expansion
type SynchronousMessage struct {
	// TODO: should we pass this in? Probably need to be able to reify the message
	//       in these cases
	// Request  MessageContents
	// Response []MessageContents
}

// SynchronousMessageBuilder is a representation of a single, bidirectional message
type SynchronousMessageBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// Given specifies a provider state
func (m *UnconfiguredSynchronousMessageBuilder) Given(state string) *UnconfiguredSynchronousMessageBuilder {
	m.messageHandle.Given(state)

	return &UnconfiguredSynchronousMessageBuilder{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// Given specifies a provider state
func (m *UnconfiguredSynchronousMessageBuilder) GivenWithParameter(state models.ProviderState) *UnconfiguredSynchronousMessageBuilder {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return &UnconfiguredSynchronousMessageBuilder{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

type UnconfiguredSynchronousMessageBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// UsingPlugin enables a plugin for use in the current test case
func (m *UnconfiguredSynchronousMessageBuilder) UsingPlugin(config PluginConfig) *SynchronousMessageWithPlugin {
	m.pact.mockserver.UsingPlugin(config.Plugin, config.Version)

	return &SynchronousMessageWithPlugin{
		pact:          m.pact,
		messageHandle: m.messageHandle,
	}
}

// UsingPlugin enables a plugin for use in the current test case
func (m *SynchronousMessageWithPlugin) UsingPlugin(config PluginConfig) *SynchronousMessageWithPlugin {
	m.pact.mockserver.UsingPlugin(config.Plugin, config.Version)

	return m
}

// AddMessage creates a new asynchronous consumer expectation
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

type SynchronousMessageWithRequest struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

type RequestBuilderFunc func(*SynchronousMessageWithRequestBuilder)

type SynchronousMessageWithRequestBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *SynchronousMessageWithRequestBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithMetadata(metadata)

	return m
}

// WithBinaryContent accepts a binary payload
func (m *SynchronousMessageWithRequestBuilder) WithBinaryContent(contentType string, body []byte) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *SynchronousMessageWithRequestBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *SynchronousMessageWithRequestBuilder) WithJSONContent(content interface{}) *SynchronousMessageWithRequestBuilder {
	m.messageHandle.WithJSONContents(content)

	return m
}

// AddMessage creates a new asynchronous consumer expectation
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

type SynchronousMessageWithResponse struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

type ResponseBuilderFunc func(*SynchronousMessageWithResponseBuilder)

type SynchronousMessageWithResponseBuilder struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *SynchronousMessageWithResponseBuilder) WithMetadata(metadata map[string]string) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithMetadata(metadata)

	return m
}

// WithBinaryContent accepts a binary payload
func (m *SynchronousMessageWithResponseBuilder) WithBinaryContent(contentType string, body []byte) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *SynchronousMessageWithResponseBuilder) WithContent(contentType string, body []byte) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *SynchronousMessageWithResponseBuilder) WithJSONContent(content interface{}) *SynchronousMessageWithResponseBuilder {
	m.messageHandle.WithJSONContents(content)

	return m
}

type SynchronousMessageWithPlugin struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

func (s *SynchronousMessageWithPlugin) WithContents(contents string, contentType string) *SynchronousMessageWithPluginContents {
	s.messageHandle.WithPluginInteractionContents(native.INTERACTION_PART_REQUEST, contentType, contents)

	return &SynchronousMessageWithPluginContents{
		pact:          s.pact,
		messageHandle: s.messageHandle,
	}
}

type SynchronousMessageWithPluginContents struct {
	messageHandle *native.Message
	pact          *SynchronousPact
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
// NOTE: currently, this function is not very useful because without a transport,
//       there is no useful way to test your actual code (because the message isn't passed back in)
//       Use at your own risk ;)
func (m *SynchronousMessageWithPluginContents) ExecuteTest(t *testing.T, integrationTest func(m SynchronousMessage) error) error {
	message := SynchronousMessage{}

	err := integrationTest(message)

	if err != nil {
		return err
	}

	return m.pact.mockserver.WritePactFile(m.pact.config.PactDir, false)
}

func (s *SynchronousMessageWithPluginContents) StartTransport(transport string, address string, config map[string][]interface{}) *SynchronousMessageWithTransport {
	port, err := s.pact.mockserver.StartTransport(transport, address, 0, make(map[string][]interface{}))

	if err != nil {
		log.Fatalln("unable to start plugin transport:", err)
	}

	return &SynchronousMessageWithTransport{
		pact:          s.pact,
		messageHandle: s.messageHandle,
		transport: TransportConfig{
			Port:    port,
			Address: address,
		},
	}
}

type SynchronousMessageWithTransport struct {
	messageHandle *native.Message
	pact          *SynchronousPact
	transport     TransportConfig
}

func (s *SynchronousMessageWithTransport) ExecuteTest(t *testing.T, integrationTest func(tc TransportConfig, m SynchronousMessage) error) error {
	message := SynchronousMessage{}

	defer s.pact.mockserver.CleanupMockServer(s.transport.Port)

	err := integrationTest(s.transport, message)

	if err != nil {
		return err
	}

	mismatches := s.pact.mockserver.MockServerMismatchedRequests(s.transport.Port)

	if len(mismatches) > 0 {
		return fmt.Errorf("pact validation failed: %+v", mismatches)
	}

	return s.pact.mockserver.WritePactFileForServer(s.transport.Port, s.pact.config.PactDir, false)
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
	log.Println("[DEBUG] pact synchronous message validate config")
	dir, _ := os.Getwd()

	if m.config.PactDir == "" {
		m.config.PactDir = filepath.Join(dir, "pacts")
	}

	m.mockserver = native.NewMessageServer(m.config.Consumer, m.config.Provider)

	return nil
}

func (m *SynchronousPact) AddSynchronousMessage(description string) *UnconfiguredSynchronousMessageBuilder {
	log.Println("[DEBUG] add sync message")

	message := m.mockserver.NewSyncMessageInteraction(description)

	return &UnconfiguredSynchronousMessageBuilder{
		messageHandle: message,
		pact:          m,
	}
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (m *SynchronousMessageWithResponse) ExecuteTest(t *testing.T, integrationTest func(md SynchronousMessage) error) error {
	message := SynchronousMessage{}

	err := integrationTest(message)

	if err != nil {
		return err
	}

	return m.pact.mockserver.WritePactFile(m.pact.config.PactDir, false)
}
