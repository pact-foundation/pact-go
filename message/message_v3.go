package message

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native/mockserver"
)

type MessageConfig struct {
	Consumer string
	Provider string
	PactDir  string
}

type MessagePactV3 struct {
	config MessageConfig

	// Reference to the native rust handle
	messageserver *mockserver.MessageServer
}

func NewMessagePactV3(config MessageConfig) (*MessagePactV3, error) {
	provider := &MessagePactV3{
		config: config,
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
		p.config.PactDir = filepath.Join(dir, "pacts")
	}

	p.messageserver = mockserver.NewMessageServer(p.config.Consumer, p.config.Provider)

	return nil
}

// AddMessage creates a new asynchronous consumer expectation
func (p *MessagePactV3) AddMessage() *Message {
	log.Println("[DEBUG] add message")

	message := p.messageserver.NewMessage()

	m := &Message{
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
func (p *MessagePactV3) verifyMessageConsumerRaw(message *Message, handler MessageConsumer) error {
	log.Printf("[DEBUG] verify message")

	// 1. Strip out the matchers
	// Reify the message back to its "example/generated" form
	body := message.messageHandle.ReifyMessage()

	log.Println("[DEBUG] reified message raw", body)

	var m AsynchronousMessage
	err := json.Unmarshal([]byte(body), &m)
	if err != nil {
		return fmt.Errorf("unexpected response from message server, this is a bug in the framework")
	}
	log.Println("[DEBUG] unmarshalled into an AsynchronousMessage", m)

	// 2. Convert to an actual type (to avoid wrapping if needed/requested)
	// 3. Invoke the message handler
	// 4. write the pact file
	t := reflect.TypeOf(message.Type)
	if t != nil && t.Name() != "interface" {
		s, err := json.Marshal(m.Content)
		if err != nil {
			return fmt.Errorf("unable to generate message for type: %+v", message.Type)
		}
		err = json.Unmarshal(s, &message.Type)

		if err != nil {
			return fmt.Errorf("unable to narrow type to %v: %v", t.Name(), err)
		}

		m.Content = message.Type
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
func (p *MessagePactV3) Verify(t *testing.T, message *Message, handler MessageConsumer) error {
	err := p.verifyMessageConsumerRaw(message, handler)

	if err != nil {
		t.Errorf("VerifyMessageConsumer failed: %v", err)
	}

	return err
}
