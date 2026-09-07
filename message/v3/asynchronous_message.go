package v3

import (
	"encoding/json"
	"errors"
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

// TODO: make a builder?

// AsynchronousMessageBuilder is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// AsynchronousMessageBuilder is the main implementation of the Pact AsynchronousMessageBuilder interface.
type AsynchronousMessageBuilder struct {
	messageHandle *native.Message
	messagePactV3 *AsynchronousPact

	// Type to Marshal content into when sending back to the consumer
	// Defaults to interface{}
	Type any

	// The handler for this message
	handler AsynchronousConsumer
}

// UnconfiguredAsynchronousMessageBuilder is a message interaction with its
// expected description set, ready to configure content.
type UnconfiguredAsynchronousMessageBuilder struct {
	rootBuilder *AsynchronousMessageBuilder
}

// GivenWithParameter specifies a provider state along with parameters
// used to set it up. Optional.
func (m *AsynchronousMessageBuilder) GivenWithParameter(state models.ProviderState) *AsynchronousMessageBuilder {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return m
}

// Given specifies a provider state. Optional.
func (m *AsynchronousMessageBuilder) Given(state string) *AsynchronousMessageBuilder {
	m.messageHandle.Given(state)

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

// WithMetadata specifies message-implementation specific metadata
// to go with the content.
func (m *UnconfiguredAsynchronousMessageBuilder) WithMetadata(metadata map[string]string) *UnconfiguredAsynchronousMessageBuilder {
	m.rootBuilder.messageHandle.WithMetadata(metadata)

	return m
}

// AsynchronousMessageBuilderWithContents is a message interaction with
// its contents set, ready to optionally narrow the decoded type via
// AsType and attach a consumer handler via ConsumedBy.
type AsynchronousMessageBuilderWithContents struct {
	rootBuilder *AsynchronousMessageBuilder
}

// WithBinaryContent accepts a binary payload.
func (m *UnconfiguredAsynchronousMessageBuilder) WithBinaryContent(contentType string, body []byte) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithContents(native.INTERACTION_PART_REQUEST, contentType, body)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// WithContent specifies the payload in bytes that the consumer expects to receive.
func (m *UnconfiguredAsynchronousMessageBuilder) WithContent(contentType string, body []byte) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithContents(native.INTERACTION_PART_REQUEST, contentType, body)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed.
func (m *UnconfiguredAsynchronousMessageBuilder) WithJSONContent(content any) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithRequestJSONContents(content)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// AsType specifies that the content sent through to the
// consumer handler should be sent as the given type.
func (m *AsynchronousMessageBuilderWithContents) AsType(t any) *AsynchronousMessageBuilderWithContents {
	log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	m.rootBuilder.Type = t

	return m
}

// AsynchronousMessageBuilderWithConsumer is a fully configured message
// interaction, ready to run via Verify.
type AsynchronousMessageBuilderWithConsumer struct {
	rootBuilder *AsynchronousMessageBuilder
}

// ConsumedBy sets the function that will consume the message.
func (m *AsynchronousMessageBuilderWithContents) ConsumedBy(handler AsynchronousConsumer) *AsynchronousMessageBuilderWithConsumer {
	m.rootBuilder.handler = handler

	return &AsynchronousMessageBuilderWithConsumer{
		rootBuilder: m.rootBuilder,
	}
}

// Verify runs the configured consumer handler against the message and
// writes the pact file if successful.
func (m *AsynchronousMessageBuilderWithConsumer) Verify(t *testing.T) error {
	t.Helper()
	return m.rootBuilder.messagePactV3.Verify(t, m.rootBuilder, m.rootBuilder.handler)
}

// AsynchronousPact is a one-way message pact between a consumer and
// provider, built via AddAsynchronousMessage.
type AsynchronousPact struct {
	config Config

	// Reference to the native rust handle
	messageserver *native.MessageServer
}

// NewMessagePact creates a new asynchronous message pact.
//
// Deprecated: use NewAsynchronousPact.
var NewMessagePact = NewAsynchronousPact

// NewAsynchronousPact creates a new asynchronous (one-way) message pact
// for the consumer/provider pair described by config.
func NewAsynchronousPact(config Config) (*AsynchronousPact, error) {
	provider := &AsynchronousPact{
		config: config,
	}
	err := provider.validateConfig()
	if err != nil {
		return nil, err
	}

	native.Init(string(logging.LogLevel()))

	return provider, err
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

	m := &AsynchronousMessageBuilder{
		messageHandle: message,
		messagePactV3: p,
	}

	return m
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

// validateConfig validates the configuration for the consumer test.
func (p *AsynchronousPact) validateConfig() error {
	log.Println("[DEBUG] pact message validate config")
	dir, _ := os.Getwd()

	if p.config.PactDir == "" {
		p.config.PactDir = filepath.Join(dir, "pacts")
	}

	p.messageserver = native.NewMessageServer(p.config.Consumer, p.config.Provider)
	p.messageserver.WithMetadata("pact-go", "version", strings.TrimPrefix(command.Version, "v"))

	return nil
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
// A Message Consumer is analogous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (p *AsynchronousPact) verifyMessageConsumerRaw(messageToVerify *AsynchronousMessageBuilder, handler AsynchronousConsumer) error {
	log.Printf("[DEBUG] verify message")

	// 1. Strip out the matchers
	// Reify the message back to its "example/generated" form
	body, err := messageToVerify.messageHandle.GetMessageRequestContents()
	log.Println("[DEBUG] reified message raw", string(body))
	if err != nil {
		return errors.New("unexpected response from message server, this is a bug in the framework")
	}

	log.Println("[DEBUG] reified message raw", string(body))

	var m MessageContents
	// err = json.Unmarshal(body, &m)
	// if err != nil {
	// 	return fmt.Errorf("unexpected response from message server, this is a bug in the framework")
	// }
	// log.Println("[DEBUG] unmarshalled into an AsynchronousMessage", m)

	// 2. Convert to an actual type (to avoid wrapping if needed/requested)
	// 3. Invoke the message handler
	// 4. write the pact file
	t := reflect.TypeOf(messageToVerify.Type)
	if t != nil && t.Name() != "interface" {
		// s, err := json.Marshal()
		// if err != nil {
		// 	return fmt.Errorf("unable to generate message for type: %+v", messageToVerify.Type)
		// }
		err = json.Unmarshal(body, &messageToVerify.Type)
		if err != nil {
			return fmt.Errorf("unable to narrow type to %v: %w", t.Name(), err)
		}

		m.Content = messageToVerify.Type
	}

	// TODO: extract metadata from FFI
	// m.Metadata =

	// Yield message, and send through handler function
	err = handler(m)
	if err != nil {
		return err
	}

	return p.messageserver.WritePactFile(p.config.PactDir, false)
}
