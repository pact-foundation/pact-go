package dsl

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/types"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *types.PactMockServer

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

	// Log levels
	LogLevel string

	// Used to detect if logging has been configured
	logFilter *logutils.LevelFilter
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *Pact) AddInteraction() *Interaction {
	p.Setup()
	log.Printf("[DEBUG] pact add interaction")
	i := &Interaction{}
	p.Interactions = append(p.Interactions, i)
	return i
}

// Setup starts the Pact Mock Server. This is usually called before each test
// suite begins. AddInteraction() will automatically call this if no Mock Server
// has been started.
func (p *Pact) Setup() *Pact {
	p.setupLogging()
	log.Printf("[DEBUG] pact setup")
	if p.Server == nil {
		client := &PactClient{Port: p.Port}
		p.pactClient = client
		p.Server = client.StartServer()
	}

	return p
}

// Configure logging
func (p *Pact) setupLogging() {
	if p.logFilter == nil {
		if p.LogLevel == "" {
			p.LogLevel = "INFO"
		}
		p.logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
			MinLevel: logutils.LogLevel(p.LogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(p.logFilter)
	}
	log.Printf("[DEBUG] pact setup logging")
}

// Teardown stops the Pact Mock Server. This usually is called on completion
// of each test suite.
func (p *Pact) Teardown() *Pact {
	log.Printf("[DEBUG] teardown")
	p.Server = p.pactClient.StopServer(p.Server)

	return p
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *Pact) Verify(integrationTest func() error) error {
	log.Printf("[DEBUG] pact verify")
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

// VerifyProvider reads the provided pact files and runs verification against
// a running Provider API.
func (p *Pact) VerifyProvider(request *types.VerifyRequest) *types.CommandResponse {
	log.Printf("[DEBUG] pact provider verification")
	return p.pactClient.VerifyProvider(request)
}
