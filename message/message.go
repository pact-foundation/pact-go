package message

import (
	"log"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native/mockserver"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
)

// MessageHandler is a provider function that generates a
// message for a Consumer given a Message context (state, description etc.)
type MessageHandler func([]models.ProviderStateV3) (interface{}, error)
type MessageProducer MessageHandler

// MessageHandlers is a list of handlers ordered by description
type MessageHandlers map[string]MessageHandler

// MessageConsumer receives a message and must be able to parse
// the content
type MessageConsumer func(AsynchronousMessage) error

// Message is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// Message is the main implementation of the Pact Message interface.
type Message struct {
	messageHandle *mockserver.Message
	messagePactV3 *MessagePactV3

	// Type to Marshall content into when sending back to the consumer
	// Defaults to interface{}
	Type interface{}

	// The handler for this message
	handler MessageConsumer
}

// V3 Message (Asynchronous only)
type AsynchronousMessage struct {
	// Message Body
	Content interface{} `json:"contents"`

	// Provider state to be written into the Pact file
	States []models.ProviderStateV3 `json:"providerStates"`

	// Message metadata
	Metadata matchers.MetadataMatcher `json:"metadata"`

	// Description to be written into the Pact file
	Description string `json:"description"`
}

// Given specifies a provider state. Optional.
func (m *Message) Given(state models.ProviderStateV3) *Message {
	m.messageHandle.GivenWithParameter(state.Name, state.Parameters)

	return m
}

// ExpectsToReceive specifies the content it is expecting to be
// given from the Provider. The function must be able to handle this
// message for the interaction to succeed.
func (m *Message) ExpectsToReceive(description string) *Message {
	m.messageHandle.ExpectsToReceive(description)

	return m
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
// func (m *Message) WithMetadata(metadata MapMatcher) *Message {
func (m *Message) WithMetadata(metadata map[string]string) *Message {
	m.messageHandle.WithMetadata(metadata)

	return m
}

// WithBinaryContent accepts a binary payload
func (m *Message) WithBinaryContent(contentType string, body []byte) *Message {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithContent specifies the payload in bytes that the consumer expects to receive
func (m *Message) WithContent(contentType string, body []byte) *Message {
	m.messageHandle.WithContents(contentType, body)

	return m
}

// WithJSONContent specifies the payload as an object (to be marshalled to WithJSONContent) that
// is expected to be consumed
func (m *Message) WithJSONContent(content interface{}) *Message {
	m.messageHandle.WithJSONContents(content)

	return m
}

// // AsType specifies that the content sent through to the
// consumer handler should be sent as the given type
func (m *Message) AsType(t interface{}) *Message {
	log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	m.Type = t

	return m
}

// The function that will consume the message
func (m *Message) ConsumedBy(handler MessageConsumer) *Message {
	m.handler = handler

	return m
}

// The function that will consume the message
func (m *Message) Verify(t *testing.T) error {
	return m.messagePactV3.Verify(t, m, m.handler)
}
