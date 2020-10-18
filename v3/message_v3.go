package v3

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type MessageConfig struct {
	Consumer             string
	Provider             string
	PactDir              string
	SpecificationVersion SpecificationVersion
}

type MessagePactV3 struct {
	messages   []*Message
	config     MessageConfig
	pactWriter pactFileV3ReaderWriter
}

func NewMessagePactV3(config MessageConfig) (*MessagePactV3, error) {
	provider := &MessagePactV3{
		config:     config,
		pactWriter: defaultPactFileV3ReaderWriter(),
	}
	err := provider.validateConfig()

	if err != nil {
		return nil, err
	}

	return provider, err
}

// validateConfig validates the configuration for the consumer test
func (p *MessagePactV3) validateConfig() error {
	log.Println("[DEBUG] pact message validate config")
	dir, _ := os.Getwd()

	if p.config.PactDir == "" {
		p.config.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	}

	return nil
}

// AddMessage creates a new asynchronous consumer expectation
func (p *MessagePactV3) AddMessage() *Message {
	log.Println("[DEBUG] pact add message")

	m := &Message{}
	p.messages = append(p.messages, m)
	return m
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
//
// A Message Consumer is analagous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (p *MessagePactV3) VerifyMessageConsumerRaw(message *Message, handler MessageConsumer) error {
	log.Printf("[DEBUG] verify message")

	// 1. Strip out the matchers
	// Reify the message back to its "example/generated" form
	body, _, _ := buildPart(message.Content)

	// 2. Convert to an actual type (to avoid wrapping if needed/requested)
	// 3. Invoke the message handler
	// 4. write the pact file
	// - fail if http interactions are already present
	// - merge interactions
	// - only append interactions if a general structure already present

	t := reflect.TypeOf(message.Type)
	if t != nil && t.Name() != "interface" {
		reified, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("unable to generate message for type: %+v", message.Type)
		}
		err = json.Unmarshal(reified, &message.Type)

		if err != nil {
			return fmt.Errorf("unable to narrow type to %v: %v", t.Name(), err)
		}
	}

	// Yield message, and send through handler function
	generatedMessage :=
		Message{
			Content:     message.Type,
			States:      message.States,
			Description: message.Description,
			Metadata:    message.Metadata,
		}

	err := handler(generatedMessage)
	if err != nil {
		return err
	}

	// // If no errors, update Message Pact
	// Generate interactions for Pact file
	serialisedPact := newPactFileV3(p.config.Consumer, p.config.Provider, nil, p.messages)

	return p.pactWriter.update(p.config.PactDir, serialisedPact)
}

// VerifyMessageConsumer is a test convience function for VerifyMessageConsumerRaw,
// accepting an instance of `*testing.T`
func (p *MessagePactV3) VerifyMessageConsumer(t *testing.T, message *Message, handler MessageConsumer) error {
	err := p.VerifyMessageConsumerRaw(message, handler)

	if err != nil {
		t.Errorf("VerifyMessageConsumer failed: %v", err)
	}

	return err
}
