package dsl

import (
	"log"
	"reflect"
)

// StateHandler is a provider function that sets up a given state before
// the provider interaction is validated
type StateHandler func(State) error

// StateHandlers is a list of StateHandler's
type StateHandlers map[string]StateHandler

// MessageHandler is a provider function that generates a
// message for a Consumer given a Message context (state, description etc.)
type MessageHandler func(Message) (interface{}, error)

// MessageHandlers is a list of handlers ordered by description
type MessageHandlers map[string]MessageHandler

// MessageConsumer receives a message and must be able to parse
// the content
type MessageConsumer func(Message) error

// Message is a representation of a single, unidirectional message
// e.g. MQ, pub/sub, Websocket, Lambda
// Message is the main implementation of the Pact Message interface.
type Message struct {
	// Message Body
	Content interface{} `json:"contents,omitempty"`

	// Message Body as a Raw JSON string
	ContentRaw interface{} `json:"-"`

	// Provider state to be written into the Pact file
	States []State `json:"providerStates,omitempty"`

	// Message metadata
	Metadata MapMatcher `json:"metadata,omitempty"`

	// Description to be written into the Pact file
	Description string `json:"description"`

	// Type to Marshall content into when sending back to the consumer
	// Defaults to interface{}
	Type interface{}

	Args []string `json:"-"`
}

// State specifies how the system should be configured when
// verified. e.g. "user A exists"
type State struct {
	Name   string                 `json:"name"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Given specifies a provider state. Optional.
func (p *Message) Given(state string) *Message {
	p.States = []State{State{Name: state}}

	return p
}

// ExpectsToReceive specifies the content it is expecting to be
// given from the Provider. The function must be able to handle this
// message for the interaction to succeed.
func (p *Message) ExpectsToReceive(description string) *Message {
	p.Description = description
	return p
}

// WithMetadata specifies message-implementation specific metadata
// to go with the content
func (p *Message) WithMetadata(metadata MapMatcher) *Message {
	p.Metadata = metadata
	return p
}

// WithContent specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (p *Message) WithContent(content interface{}) *Message {
	p.Content = content

	return p
}

// AsType specifies that the content sent through to the
// consumer handler should be sent as the given type
func (p *Message) AsType(t interface{}) *Message {
	log.Println("[DEBUG] setting Message decoding to type:", reflect.TypeOf(t))
	p.Type = t

	return p
}
