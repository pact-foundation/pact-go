package dsl

import (
	"fmt"

	"github.com/mefellows/pact-go/daemon"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *daemon.PactMockServer

	// Port the Pact Daemon is running on.
	Port int

	// Pact RPC Client.
	pactClient *PactClient

	// Consumer is the name of the Consumer/Client.
	Consumer string

	// Provider is the name of the Providing service.
	Provider string

	// Interactions contains all of the Mock Service Interactions to be setup.
	Interactions []*Interaction
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *Pact) AddInteraction() *Interaction {
	if p.Server == nil {
		p.Setup()
	}
	i := &Interaction{}
	p.Interactions = append(p.Interactions, i)
	return i
}

// Setup starts the Pact Mock Server. This is usually called before each test
// suite begins. AddInteraction() will automatically call this if no Mock Server
// has been started.
func (p *Pact) Setup() *Pact {
	client := &PactClient{Port: p.Port}
	p.pactClient = client
	p.Server = client.StartServer()

	return p
}

// Teardown stops the Pact Mock Server. This usually is called on completion
// of each test suite.
func (p *Pact) Teardown() *Pact {
	p.Server = p.pactClient.StopServer(p.Server)

	return p
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *Pact) Verify(integrationTest func() error) error {
	mockServer := &PactMockService{
		BaseURL:  fmt.Sprintf("http://localhost:%d", p.Server.Port),
		Consumer: p.Consumer,
		Provider: p.Provider,
	}

	for _, interaction := range p.Interactions {
		err := mockServer.AddInteraction(interaction)
		if err != nil {
			return err
		}
	}

	// Run the integration test
	integrationTest()

	// Run Verification Process
	err := mockServer.Verify()
	if err != nil {
		return err
	}

	err = mockServer.WritePact()
	if err != nil {
		return err
	}

	return mockServer.DeleteInteractions()
}
