package v3

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	mockserver "github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/models"
)

// TODO: make a builder?

// AsynchronousMessageBuilder is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// AsynchronousMessageBuilder is the main implementation of the Pact AsynchronousMessageBuilder interface.
type AsynchronousMessageBuilder struct {
	messageHandle *mockserver.Message
	messagePactV3 *AsynchronousPact

	// Type to Marshal content into when sending back to the consumer
	// Defaults to interface{}
	Type interface{}

	// The handler for this message
	handler AsynchronousConsumer
}

type UnconfiguredAsynchronousMessageBuilder struct {
	rootBuilder *AsynchronousMessageBuilder
}

// Given specifies a provider state. Optional.
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
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *UnconfiguredAsynchronousMessageBuilder) WithMetadata(metadata map[string]string) *UnconfiguredAsynchronousMessageBuilder {
	m.rootBuilder.messageHandle.WithMetadata(metadata)

	return m
}

type AsynchronousMessageBuilderWithContents struct {
	rootBuilder *AsynchronousMessageBuilder
}

// WithBinaryContent accepts a binary payload
func (m *UnconfiguredAsynchronousMessageBuilder) WithBinaryContent(contentType string, body []byte) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithContents(mockserver.INTERACTION_PART_REQUEST, contentType, body)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *UnconfiguredAsynchronousMessageBuilder) WithContent(contentType string, body []byte) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithContents(mockserver.INTERACTION_PART_REQUEST, contentType, body)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *UnconfiguredAsynchronousMessageBuilder) WithJSONContent(content interface{}) *AsynchronousMessageBuilderWithContents {
	m.rootBuilder.messageHandle.WithRequestJSONContents(content)

	return &AsynchronousMessageBuilderWithContents{
		rootBuilder: m.rootBuilder,
	}
}

// // AsType specifies that the content sent through to the
// consumer handler should be sent as the given type
func (m *AsynchronousMessageBuilderWithContents) AsType(t interface{}) *AsynchronousMessageBuilderWithContents {
	log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	m.rootBuilder.Type = t

	return m
}

type AsynchronousMessageBuilderWithConsumer struct {
	rootBuilder *AsynchronousMessageBuilder
}

// The function that will consume the message
func (m *AsynchronousMessageBuilderWithContents) ConsumedBy(handler AsynchronousConsumer) *AsynchronousMessageBuilderWithConsumer {
	m.rootBuilder.handler = handler

	return &AsynchronousMessageBuilderWithConsumer{
		rootBuilder: m.rootBuilder,
	}
}

// The function that will consume the message
func (m *AsynchronousMessageBuilderWithConsumer) Verify(t *testing.T) error {
	return m.rootBuilder.messagePactV3.Verify(t, m.rootBuilder, m.rootBuilder.handler)
}

type AsynchronousPact struct {
	config Config

	// Reference to the native rust handle
	messageserver *mockserver.MessageServer
}

// Deprecated: use NewAsynchronousPact
var NewMessagePact = NewAsynchronousPact

func NewAsynchronousPact(config Config) (*AsynchronousPact, error) {
	provider := &AsynchronousPact{
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
func (p *AsynchronousPact) validateConfig() error {
	log.Println("[DEBUG] pact message validate config")
	dir, _ := os.Getwd()

	if p.config.PactDir == "" {
		p.config.PactDir = filepath.Join(dir, "pacts")
	}

	p.messageserver = mockserver.NewMessageServer(p.config.Consumer, p.config.Provider)

	return nil
}

// AddMessage creates a new asynchronous consumer expectation
// Deprecated: use AddAsynchronousMessage() instead
func (p *AsynchronousPact) AddMessage() *AsynchronousMessageBuilder {
	return p.AddAsynchronousMessage()
}

// AddMessage creates a new asynchronous consumer expectation
func (p *AsynchronousPact) AddAsynchronousMessage() *AsynchronousMessageBuilder {
	log.Println("[DEBUG] add message")

	message := p.messageserver.NewMessage()

	m := &AsynchronousMessageBuilder{
		messageHandle: message,
		messagePactV3: p,
	}

	return m
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
//
// A Message Consumer is analagous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (p *AsynchronousPact) verifyMessageConsumerRaw(messageToVerify *AsynchronousMessageBuilder, handler AsynchronousConsumer) error {
	log.Printf("[DEBUG] verify message")

	// 1. Strip out the matchers
	// Reify the message back to its "example/generated" form
	body := messageToVerify.messageHandle.ReifyMessage()

	log.Println("[DEBUG] reified message raw", body)

	var m MessageContents
	err := json.Unmarshal([]byte(body), &m)
	if err != nil {
		return fmt.Errorf("unexpected response from message server, this is a bug in the framework")
	}
	log.Println("[DEBUG] unmarshalled into an AsynchronousMessage", m)

	// 2. Convert to an actual type (to avoid wrapping if needed/requested)
	// 3. Invoke the message handler
	// 4. write the pact file
	t := reflect.TypeOf(messageToVerify.Type)
	if t != nil && t.Name() != "interface" {
		s, err := json.Marshal(m.Content)
		if err != nil {
			return fmt.Errorf("unable to generate message for type: %+v", messageToVerify.Type)
		}
		err = json.Unmarshal(s, &messageToVerify.Type)

		if err != nil {
			return fmt.Errorf("unable to narrow type to %v: %v", t.Name(), err)
		}

		m.Content = messageToVerify.Type
	}

	// Yield message, and send through handler function
	err = handler(m)

	if err != nil {
		return err
	}

	return p.messageserver.WritePactFile(p.config.PactDir, false)
}

// VerifyMessageConsumer is a test convience function for VerifyMessageConsumerRaw,
// accepting an instance of `*testing.T`
func (p *AsynchronousPact) Verify(t *testing.T, message *AsynchronousMessageBuilder, handler AsynchronousConsumer) error {
	err := p.verifyMessageConsumerRaw(message, handler)

	if err != nil {
		t.Errorf("VerifyMessageConsumer failed: %v", err)
	}

	return err
}
