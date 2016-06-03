package dsl

import (
	"fmt"

	"github.com/mefellows/pact-go/daemon"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact interface {
	// Given specifies a provider state. Optional.
	Given(string) Pact

	// UponReceiving specifies the name of the test case. This becomes the name of
	// the consumer/provider pair in the Pact file. Mandatory.
	UponReceiving(string) Pact

	// WillRespondWith specifies the details of the HTTP response that will be used to
	// confirm that the Provider must satisfy. Mandatory.
	WithRequest(Request) Pact

	// WillRespondWith specifies the details of the HTTP response that will be used to
	// confirm that the Provider must satisfy. Mandatory.
	WillRespondWith(Response) Pact

	// Verify runs the current test case against a Mock Service.
	Verify() error
}

// PactConsumer is the main implementation of the Pact interface.
type PactConsumer struct {
	// Current server for the consumer
	server *daemon.PactMockServer

	// Stores the current state
	State []string

	// Port the Pact Daemon is running on
	Port int

	// Pact RPC Client
	pactClient *PactClient
}

// Before starts the Pact Mock Server before each test suite.
func (p *PactConsumer) Before() Pact {
	client := &PactClient{Port: p.Port}
	p.pactClient = client
	p.server = client.StartServer()

	return p
}

// After stops the Pact Mock Server after each test suite.
func (p *PactConsumer) After() Pact {
	p.server = p.pactClient.StopServer(p.server)

	return p
}

// Given specifies a provider state. Optional.
func (p *PactConsumer) Given(state string) Pact {
	fmt.Println("Pact()")
	return p
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (p *PactConsumer) UponReceiving(test string) Pact {
	fmt.Println("UponReceiving()")
	return p
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (p *PactConsumer) WithRequest(request Request) Pact {
	fmt.Println("WithRequest()")
	return p
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
func (p *PactConsumer) WillRespondWith(response Response) Pact {
	fmt.Println("RespondWith()")
	return p
}

// Verify runs the current test case against a Mock Service.
func (p *PactConsumer) Verify() error {
	fmt.Println("Verify()")
	return nil
}
