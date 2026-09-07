package v4

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/command"
	"github.com/pact-foundation/pact-go/v2/internal/native"
	logging "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
)

// Builder 1: Async with no plugin/transport
// Builder 2: Async with plugin content no transport
// Builder 3: Async with plugin content + transport

// AsynchronousMessageBuilder builds a single, one-way message interaction,
// starting from AddAsynchronousMessage.
type AsynchronousMessageBuilder struct {
	messageHandle *native.Message
	pact          *AsynchronousPact

	// Type to Marshal content into when sending back to the consumer
	// Defaults to interface{}
	Type any

	// The handler for this message
	handler AsynchronousConsumer
}

// Given specifies a provider state. Optional.
func (m *AsynchronousMessageBuilder) Given(state string) *AsynchronousMessageBuilder {
	m.messageHandle.Given(state)

	return m
}

// GivenWithParameter specifies a provider state along with parameters
// used to set it up. Optional.
func (m *AsynchronousMessageBuilder) GivenWithParameter(state models.ProviderState) *AsynchronousMessageBuilder {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return m
}

// AddExternalReference records a reference to an external resource (such as a ticket or
// pull request) against the interaction. References appear under
// comments.references[group][name] in the Pact file. May be called multiple times.
func (m *AsynchronousMessageBuilder) AddExternalReference(group, name, value string) *AsynchronousMessageBuilder {
	m.messageHandle.WithReference(group, name, value)

	return m
}

// ExpectsToReceive specifies the content it is expecting to be
// given from the Provider. The function must be able to handle this
// message for the interaction to succeed.
func (m *AsynchronousMessageBuilder) ExpectsToReceive(description string) *UnconfiguredAsynchronousMessageBuilder {
	m.messageHandle.ExpectsToReceive(description)

	return &UnconfiguredAsynchronousMessageBuilder{
		rootBuilder: m,
	}
}

// UnconfiguredAsynchronousMessageBuilder is a message interaction with its
// expected description set, ready to configure content directly or via a
// plugin.
type UnconfiguredAsynchronousMessageBuilder struct {
	rootBuilder *AsynchronousMessageBuilder
}

// UsingPlugin enables a plugin for use in the current test case.
func (m *UnconfiguredAsynchronousMessageBuilder) UsingPlugin(config PluginConfig) *AsynchronousMessageWithPlugin {
	err := m.rootBuilder.pact.messageserver.UsingPlugin(config.Plugin, config.Version)
	if err != nil {
		log.Println("[ERROR] failed to add plugin:", err)
		panic(err)
	}

	return &AsynchronousMessageWithPlugin{
		rootBuilder: m.rootBuilder,
	}
}

// AsynchronousMessageWithPlugin is a message interaction with a plugin
// loaded via UsingPlugin, ready to have its contents set by that plugin.
type AsynchronousMessageWithPlugin struct {
	rootBuilder *AsynchronousMessageBuilder
}

// WithContents sets the contents of this message from contents,
// interpreted by the loaded plugin according to contentType (e.g. a
// protobuf message type).
func (s *AsynchronousMessageWithPlugin) WithContents(contents string, contentType string) *AsynchronousMessageWithPluginContents {
	err := s.rootBuilder.messageHandle.WithPluginInteractionContents(native.INTERACTION_PART_REQUEST, contentType, contents)
	if err != nil {
		log.Println("[ERROR] failed to get plugin content from message handle:", err)
		panic(err)
	}

	return &AsynchronousMessageWithPluginContents{
		rootBuilder: s.rootBuilder,
	}
}

// AsynchronousMessageWithPluginContents is a plugin-backed message
// interaction with its contents set, ready to run directly via
// ExecuteTest or to start a transport first via StartTransport.
type AsynchronousMessageWithPluginContents struct {
	rootBuilder *AsynchronousMessageBuilder
}

// ExecuteTest runs integrationTest against the reified message, then
// writes the pact file if successful.
func (s *AsynchronousMessageWithPluginContents) ExecuteTest(t *testing.T, integrationTest func(m AsynchronousMessage) error) error {
	t.Helper()
	defer s.rootBuilder.pact.messageserver.CleanupPlugins()
	message, err := getAsynchronousMessageWithReifiedContents(s.rootBuilder.messageHandle, s.rootBuilder.Type)
	if err != nil {
		return err
	}

	fmt.Println()
	err = integrationTest(message)
	if err != nil {
		return err
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return s.rootBuilder.pact.messageserver.WritePactFile(s.rootBuilder.pact.config.PactDir, false)
}

// StartTransport starts a plugin-provided mock server for transport on
// address, for tests that need a live connection rather than reading the
// message contents directly.
func (s *AsynchronousMessageWithPluginContents) StartTransport(transport string, address string, config map[string][]any) *AsynchronousMessageWithTransport {
	port, err := s.rootBuilder.pact.messageserver.StartTransport(transport, address, 0, make(map[string][]any))
	if err != nil {
		log.Fatalln("unable to start plugin transport:", err)
	}

	return &AsynchronousMessageWithTransport{
		rootBuilder: s.rootBuilder,
		transport: TransportConfig{
			Port:    port,
			Address: address,
		},
	}
}

// AsynchronousMessageWithTransport is a plugin-backed message interaction
// with a live transport running, ready to run via ExecuteTest.
type AsynchronousMessageWithTransport struct {
	rootBuilder *AsynchronousMessageBuilder
	transport   TransportConfig
}

// ExecuteTest runs integrationTest against the transport started by
// StartTransport, then verifies the interaction and writes the pact file
// if successful.
func (s *AsynchronousMessageWithTransport) ExecuteTest(t *testing.T, integrationTest func(tc TransportConfig, m AsynchronousMessage) error) error {
	t.Helper()
	defer s.rootBuilder.pact.messageserver.CleanupMockServer(s.transport.Port)
	defer s.rootBuilder.pact.messageserver.CleanupPlugins()
	message, err := getAsynchronousMessageWithReifiedContents(s.rootBuilder.messageHandle, s.rootBuilder.Type)
	if err != nil {
		return err
	}

	err = integrationTest(s.transport, message)
	if err != nil {
		return err
	}

	mismatches := s.rootBuilder.pact.messageserver.MockServerMismatchedRequests(s.transport.Port)

	if len(mismatches) > 0 {
		return fmt.Errorf("pact validation failed: %+v", mismatches)
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return s.rootBuilder.pact.messageserver.WritePactFileForServer(s.transport.Port, s.rootBuilder.pact.config.PactDir, false)
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content.
func (m *UnconfiguredAsynchronousMessageBuilder) WithMetadata(metadata map[string]string) *UnconfiguredAsynchronousMessageBuilder {
	m.rootBuilder.messageHandle.WithMetadata(metadata)

	return m
}

// AsynchronousMessageWithContents is a message interaction with its
// contents set, ready to optionally narrow the decoded type via AsType
// and attach a consumer handler via ConsumedBy.
type AsynchronousMessageWithContents struct {
	rootBuilder *AsynchronousMessageBuilder
}

// WithContent specifies the payload in bytes that the consumer expects to receive.
func (m *UnconfiguredAsynchronousMessageBuilder) WithContent(contentType string, body []byte) *AsynchronousMessageWithContents {
	m.rootBuilder.messageHandle.WithContents(native.INTERACTION_PART_REQUEST, contentType, body)

	return &AsynchronousMessageWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed.
func (m *UnconfiguredAsynchronousMessageBuilder) WithJSONContent(content any) *AsynchronousMessageWithContents {
	m.rootBuilder.messageHandle.WithRequestJSONContents(content)

	return &AsynchronousMessageWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// AsType specifies that the content sent through to the
// consumer handler should be sent as the given type.
func (m *AsynchronousMessageWithContents) AsType(t any) *AsynchronousMessageWithContents {
	log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	m.rootBuilder.Type = t

	return m
}

// ConsumedBy sets the function that will consume the message.
func (m *AsynchronousMessageWithContents) ConsumedBy(handler AsynchronousConsumer) *AsynchronousMessageWithConsumer {
	m.rootBuilder.handler = handler

	return &AsynchronousMessageWithConsumer{
		rootBuilder: m.rootBuilder,
	}
}

// AsynchronousMessageWithConsumer is a fully configured message
// interaction, ready to run via Verify.
type AsynchronousMessageWithConsumer struct {
	rootBuilder *AsynchronousMessageBuilder
}

// Verify runs the configured consumer handler against the message and
// writes the pact file if successful.
func (m *AsynchronousMessageWithConsumer) Verify(t *testing.T) error {
	t.Helper()
	return m.rootBuilder.pact.Verify(t, m.rootBuilder, m.rootBuilder.handler)
}

// AsynchronousPact is a one-way message pact between a consumer and
// provider, built via AddAsynchronousMessage.
type AsynchronousPact struct {
	config Config

	// Reference to the native rust handle
	messageserver *native.MessageServer
}

// NewAsynchronousPact creates a new asynchronous (one-way) message pact
// for the consumer/provider pair described by config.
func NewAsynchronousPact(config Config) (*AsynchronousPact, error) {
	provider := &AsynchronousPact{
		config: config,
	}
	provider.validateConfig()

	native.Init(string(logging.LogLevel()))

	return provider, nil
}

// AddMessage creates a new asynchronous consumer expectation
//
// Deprecated: use AddAsynchronousMessage() instead.
func (p *AsynchronousPact) AddMessage() *AsynchronousMessageBuilder {
	return p.AddAsynchronousMessage()
}

// AddAsynchronousMessage creates a new asynchronous consumer expectation.
func (p *AsynchronousPact) AddAsynchronousMessage() *AsynchronousMessageBuilder {
	log.Println("[DEBUG] add message")

	message := p.messageserver.NewAsyncMessageInteraction("")

	return &AsynchronousMessageBuilder{
		messageHandle: message,
		pact:          p,
	}
}

// Verify is a test convenience function for verifyMessageConsumerRaw,
// accepting an instance of `*testing.T`.
func (p *AsynchronousPact) Verify(t *testing.T, message *AsynchronousMessageBuilder, handler AsynchronousConsumer) error {
	t.Helper()
	err := p.verifyMessageConsumerRaw(message, handler)
	if err != nil {
		t.Errorf("VerifyMessageConsumer failed: %v", err)
	}

	return err
}

// validateConfig applies the defaults for any unset consumer test configuration.
func (p *AsynchronousPact) validateConfig() {
	log.Println("[DEBUG] pact message validate config")
	dir, _ := os.Getwd()

	if p.config.PactDir == "" {
		p.config.PactDir = filepath.Join(dir, "pacts")
	}

	p.messageserver = native.NewMessageServer(p.config.Consumer, p.config.Provider)
	p.messageserver.WithSpecificationVersion(native.SPECIFICATION_VERSION_V4)
	p.messageserver.WithMetadata("pact-go", "version", strings.TrimPrefix(command.Version, "v"))
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
// A Message Consumer is analogous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (p *AsynchronousPact) verifyMessageConsumerRaw(messageToVerify *AsynchronousMessageBuilder, handler AsynchronousConsumer) error {
	log.Printf("[DEBUG] verify message")

	m, err := getAsynchronousMessageWithReifiedContents(messageToVerify.messageHandle, messageToVerify.Type)
	if err != nil {
		return err
	}

	// Yield message, and send through handler function
	err = handler(m)
	if err != nil {
		return err
	}

	//nolint:wrapcheck // MessageServer.WritePactFile returns internal/native's
	// sentinels (ErrUnableToWritePactFile, ErrMockServerPanic, ...) whose text
	// mirrors the pact_ffi return code, and this is the last frame before the
	// user's test output.
	return p.messageserver.WritePactFile(p.config.PactDir, false)
}

func getAsynchronousMessageWithContents(message *native.Message) (AsynchronousMessage, error) {
	var m AsynchronousMessage

	contents, err := message.GetMessageRequestContents()
	if err != nil {
		//nolint:wrapcheck // native.Message's contents accessors return
		// internal/native's FFI-derived error; the callers of this helper add the
		// "bug in the framework" context.
		return m, err
	}

	return AsynchronousMessage{
		Contents: contents,
	}, nil
}

func getAsynchronousMessageWithReifiedContents(message *native.Message, reifiedType any) (AsynchronousMessage, error) {
	var m AsynchronousMessage
	var err error

	m, err = getAsynchronousMessageWithContents(message)
	if err != nil {
		return m, fmt.Errorf("unexpected response from message server, this is a bug in the framework: %w", err)
	}
	log.Println("[DEBUG] reified body raw", string(m.Contents))

	// // 1. Strip out the matchers
	// // Reify the message back to its "example/generated" form
	// body, err := message.GetMessageRequestContents()
	// if err != nil {
	// 	return m, fmt.Errorf("unexpected response from message server, this is a bug in the framework: %v", err)
	// }

	log.Println("[DEBUG] unmarshalled into an AsynchronousMessage", m)

	// 2. Convert to an actual type (to avoid wrapping if needed/requested)
	t := reflect.TypeOf(reifiedType)
	if t != nil && t.Name() != "interface" {
		err = json.Unmarshal(m.Contents, &reifiedType)
		if err != nil {
			return m, fmt.Errorf("unable to narrow type to %v: %w", t.Name(), err)
		}

		m.Body = reifiedType
	}

	return m, nil
}
