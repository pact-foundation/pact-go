//package v3 contains the main Pact DSL used in the Consumer
// collaboration test cases, and Provider contract test verification.
package v3

// TODO: setup a proper state machine to prevent actions
// Current issues
// 1. Setup needs to be initialised to get a port

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/utils"
	"github.com/pact-foundation/pact-go/v3/install"
	"github.com/pact-foundation/pact-go/v3/native"
)

// MockProvider is the container structure to run the Consumer MockProvider test cases.
type MockProvider struct {
	// Consumer is the name of the Consumer/Client.
	Consumer string

	// Provider is the name of the Providing service.
	Provider string

	// Interactions contains all of the Mock Service Interactions to be setup.
	Interactions []*Interaction

	// Log levels.
	LogLevel logutils.LogLevel

	// Location of Pact external service invocation output logging.
	// Defaults to `<cwd>/logs`.
	LogDir string

	// Pact files will be saved in this folder.
	// Defaults to `<cwd>/pacts`.
	PactDir string

	// Host is the address of the Mock and Verification Service runs on
	// Examples include 'localhost', '127.0.0.1', '[::1]'
	// Defaults to 'localhost'
	Host string

	// Network is the network of the Mock and Verification Service
	// Examples include 'tcp', 'tcp4', 'tcp6'
	// Defaults to 'tcp'
	Network string

	// Ports MockServer can be deployed to, can be CSV or Range with a dash
	// Example "1234", "12324,5667", "1234-5667"
	AllowedMockServerPorts string

	// Port the mock service should run on. Leave blank to have one assigned
	// automatically by the OS.
	// Use AllowedMockServerPorts to constrain the assigned range.
	// TODO: visit port at this level. If we want to allow multiple interaction tests for
	// the same consumer provider, this is probably best done at the Verify() step
	Port int

	// DisableToolValidityCheck prevents CLI version checking - use this carefully!
	// The ideal situation is to check the tool installation with  before running
	// the tests, which should speed up large test suites significantly
	DisableToolValidityCheck bool

	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration

	// Check if CLI tools are up to date
	toolValidityCheck bool

	SpecificationVersion SpecificationVersion

	// fsm state of the interaction
	state string

	// TLS enables a mock service behind a self-signed certificate
	// TODO: document and test this
	TLS bool

	// TODO: clean pact dir on run?
	// CleanPactFile bool
}

// TODO: pass this into verification test func
type MockServerConfig struct{}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *MockProvider) AddInteraction() *Interaction {
	p.Setup()
	log.Println("[DEBUG] pact add interaction")
	i := &Interaction{}
	p.Interactions = append(p.Interactions, i)
	return i
}

// Setup starts the Pact Mock Server. This is usually called before each test
// suite begins. AddInteraction() will automatically call this if no Mock Server
// has been started.
func (p *MockProvider) Setup() (*MockProvider, error) {
	setLogLevel(p.LogLevel)
	log.Println("[DEBUG] pact setup")
	dir, _ := os.Getwd()

	if p.Network == "" {
		p.Network = "tcp"
	}

	// TODO: use installer to download runtime dependencies
	// if !p.toolValidityCheck && !(p.DisableToolValidityCheck || os.Getenv("PACT_DISABLE_TOOL_VALIDITY_CHECK") != "") {
	// 	checkCliCompatibility()
	// 	p.toolValidityCheck = true
	// }

	if p.Host == "" {
		p.Host = "127.0.0.1"
	}

	if p.LogDir == "" {
		p.LogDir = fmt.Sprintf(filepath.Join(dir, "logs"))
	}

	if p.PactDir == "" {
		p.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	}

	if p.ClientTimeout == 0 {
		p.ClientTimeout = 10 * time.Second
	}

	var pErr error
	if p.AllowedMockServerPorts != "" && p.Port <= 0 {
		p.Port, pErr = utils.FindPortInRange(p.AllowedMockServerPorts)
	} else if p.Port <= 0 {
		p.Port, pErr = utils.GetFreePort()
	}

	if pErr != nil {
		return nil, fmt.Errorf("error: unable to find free port, mock server will fail to start")
	}

	native.Init()

	return p, nil
}

// 	// TODO: do we still want this lifecycle method?
// Teardown stops the Pact Mock Server. This usually is called on completion
// // of each test suite.
// func (p *MockProvider) Teardown() error {
// 	log.Println("[DEBUG] teardown")

// 	if p.Port != 0 {
// 		err := native.WritePactFile(p.Port, p.PactDir)
// 		if err != nil {
// 			return err
// 		}

// 		if native.CleanupMockServer(p.Port) {
// 			p.Port = 0
// 		} else {
// 			log.Println("[DEBUG] unable to teardown server")
// 		}
// 	}
// 	return nil
// }

func (p *MockProvider) cleanInteractions() {
	p.Interactions = make([]*Interaction, 0)
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (p *MockProvider) ExecuteTest(integrationTest func(MockServerConfig) error) error {
	log.Println("[DEBUG] pact verify")
	p.Setup()

	// Generate interactions for Pact file
	serialisedPact := NewPactFile(p.Consumer, p.Provider, p.Interactions, p.SpecificationVersion)
	fmt.Println("[DEBUG] Sending pact file:", formatJSONObject(serialisedPact))

	// Clean interactions
	p.cleanInteractions()

	// TODO: wire this better
	native.CreateMockServer(formatJSONObject(serialisedPact), fmt.Sprintf("%s:%d", p.Host, p.Port), p.TLS)

	// TODO: use cases for having server running post integration test?
	// Probably not...
	defer native.CleanupMockServer(p.Port)

	// Run the integration test
	err := integrationTest(MockServerConfig{})

	if err != nil {
		return err
	}

	// Run Verification Process
	res, mismatches := native.Verify(p.Port, p.PactDir)
	p.displayMismatches(mismatches)

	if !res {
		return fmt.Errorf("pact validation failed: %+v %+v", res, mismatches)
	}
	p.WritePact()

	return nil
}

// TODO: pretty print this to make it really easy to understand the problems
// See existing Pact/Ruby code examples
// What about the Rust/Elm compiler feedback, they are pretty great too.
func (p *MockProvider) displayMismatches(mismatches []native.MismatchedRequest) {
	if len(mismatches) > 0 {
		log.Println("[INFO] pact validation failed, errors: ")
		for _, m := range mismatches {
			formattedRequest := fmt.Sprintf("%s %s", m.Request.Method, m.Request.Path)
			switch m.Type {
			case "missing-request":
				fmt.Printf("Expected request to: %s, but did not receive one\n", formattedRequest)
			case "request-not-found":
				fmt.Printf("Unexpected request was received: %s\n", formattedRequest)
			default:
				// TODO:
			}

			for _, detail := range m.Mismatches {
				switch detail.Type {
				case "HeaderMismatch":
					fmt.Printf("Comparing Header: '%s'\n", detail.Key)
					fmt.Println(detail.Mismatch)
					fmt.Println("Expected:", detail.Expected)
					fmt.Println("Actual:", detail.Actual)
				}
			}
		}
	}
}

// WritePact should be called writes when all tests have been performed for a
// given Consumer <-> Provider pair. It will write out the Pact to the
// configured file. This is safe to call multiple times as the service is smart
// enough to merge pacts and avoid duplicates.
func (p *MockProvider) WritePact() error {
	log.Println("[DEBUG] write pact file")
	if p.Port != 0 {
		return native.WritePactFile(p.Port, p.PactDir)
	}
	return errors.New("pact server not yet started")
}

var installer = install.NewInstaller()

var checkCliCompatibility = func() {
	log.Println("[DEBUG] checking CLI compatability")
	err := installer.CheckInstallation()

	if err != nil {
		log.Fatal("[ERROR] CLI tools are out of date, please upgrade before continuing")
	}
}

// Format a JSON document to make comparison easier.
func formatJSONString(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}

// Format a JSON document for creating Pact files.
func formatJSONObject(object interface{}) string {
	out, _ := json.Marshal(object)
	return formatJSONString(string(out))
}

// GetGetTLSConfigForTLSMockServer gets an http transport with
// the certificates already trusted. Alternatively, simply set
// trust level to insecure
func GetTLSConfigForTLSMockServer() *http.Transport {
	return &http.Transport{
		TLSClientConfig: native.GetTLSConfig(),
	}
}
