/*
Package dsl contains the main Pact DSL used in the Consumer
collaboration test cases, and Provider contract test verification.
*/
package dsl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/types"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *types.MockServer

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

	// Log levels.
	LogLevel string

	// Used to detect if logging has been configured.
	logFilter *logutils.LevelFilter

	// Location of Pact external service invocation output logging.
	// Defaults to `<cwd>/logs`.
	LogDir string

	// Pact files will be saved in this folder.
	// Defaults to `<cwd>/pacts`.
	PactDir string

	// Specify which version of the Pact Specification should be used (1 or 2).
	// Defaults to 2.
	SpecificationVersion int
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
	dir, _ := os.Getwd()

	if p.LogDir == "" {
		p.LogDir = fmt.Sprintf(filepath.Join(dir, "logs"))
	}

	if p.PactDir == "" {
		p.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	}

	if p.SpecificationVersion == 0 {
		p.SpecificationVersion = 2
	}

	if p.Server == nil {
		args := []string{
			fmt.Sprintf("--pact-specification-version %d", p.SpecificationVersion),
			fmt.Sprintf("--pact-dir %s", filepath.FromSlash(p.PactDir)),
			fmt.Sprintf("--log %s", filepath.FromSlash(p.LogDir+"/"+"pact.log")),
			fmt.Sprintf("--consumer %s", p.Consumer),
			fmt.Sprintf("--provider %s", p.Provider),
		}
		client := &PactClient{Port: p.Port}
		p.pactClient = client
		p.Server = client.StartServer(args)
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
	if p.Server != nil {
		p.Server = p.pactClient.StopServer(p.Server)
	}
	return p
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *Pact) Verify(integrationTest func() error) error {
	p.Setup()
	log.Printf("[DEBUG] pact verify")
	mockServer := &MockService{
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

	// Clear out interations
	p.Interactions = make([]*Interaction, 0)

	return mockServer.DeleteInteractions()
}

// WritePact should be called writes when all tests have been performed for a
// given Consumer <-> Provider pair. It will write out the Pact to the
// configured file.
func (p *Pact) WritePact() error {
	p.Setup()
	log.Printf("[DEBUG] pact write Pact file")
	mockServer := MockService{
		BaseURL:  fmt.Sprintf("http://localhost:%d", p.Server.Port),
		Consumer: p.Consumer,
		Provider: p.Provider,
	}
	err := mockServer.WritePact()
	if err != nil {
		return err
	}

	return nil
}

// VerifyProvider reads the provided pact files and runs verification against
// a running Provider API.
func (p *Pact) VerifyProvider(request types.VerifyRequest) error {
	p.Setup()

	// If we provide a Broker, we go to it to find consumers
	if request.BrokerURL != "" {
		log.Printf("[DEBUG] pact provider verification - finding all consumers from broker: %s", request.BrokerURL)
		err := findConsumers(p.Provider, &request)
		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] pact provider verification")

	content, err := p.pactClient.VerifyProvider(request)

	// Output test result to screen
	log.Println(content)

	return err
}
