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
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/v3/install"
	"github.com/pact-foundation/pact-go/v3/native"
)

// MockProvider is the container structure to run the Consumer MockProvider test cases.
type MockProvider struct {
	// Current server for the consumer.
	ServerPort int `json:"-"`

	// Pact RPC Client.
	// pactClient *PactClient

	// Consumer is the name of the Consumer/Client.
	Consumer string `json:"consumer"`

	// Provider is the name of the Providing service.
	Provider string `json:"provider"`

	// Interactions contains all of the Mock Service Interactions to be setup.
	Interactions []*Interaction `json:"interactions"`

	// Log levels.
	LogLevel string `json:"-"`

	// Used to detect if logging has been configured.
	logFilter *logutils.LevelFilter

	// Location of Pact external service invocation output logging.
	// Defaults to `<cwd>/logs`.
	LogDir string `json:"-"`

	// Pact files will be saved in this folder.
	// Defaults to `<cwd>/pacts`.
	PactDir string `json:"-"`

	// Host is the address of the Mock and Verification Service runs on
	// Examples include 'localhost', '127.0.0.1', '[::1]'
	// Defaults to 'localhost'
	Host string `json:"-"`

	// Network is the network of the Mock and Verification Service
	// Examples include 'tcp', 'tcp4', 'tcp6'
	// Defaults to 'tcp'
	Network string `json:"-"`

	// Ports MockServer can be deployed to, can be CSV or Range with a dash
	// Example "1234", "12324,5667", "1234-5667"
	AllowedMockServerPorts string `json:"-"`

	// DisableToolValidityCheck prevents CLI version checking - use this carefully!
	// The ideal situation is to check the tool installation with  before running
	// the tests, which should speed up large test suites significantly
	DisableToolValidityCheck bool `json:"-"`

	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration `json:"-"`

	// Check if CLI tools are up to date
	toolValidityCheck bool `json:"-"`

	SpecificationVersion SpecificationVersion
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
func (p *MockProvider) Setup() *MockProvider {
	p.setupLogging()
	log.Println("[DEBUG] pact setup")
	dir, _ := os.Getwd()

	if p.Network == "" {
		p.Network = "tcp"
	}

	// if !p.toolValidityCheck && !(p.DisableToolValidityCheck || os.Getenv("PACT_DISABLE_TOOL_VALIDITY_CHECK") != "") {
	// 	checkCliCompatibility()
	// 	p.toolValidityCheck = true
	// }

	if p.Host == "" {
		p.Host = "localhost"
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

	// if p.pactClient == nil {
	// 	c := NewClient()
	// 	c.TimeoutDuration = p.ClientTimeout
	// 	p.pactClient = c
	// }

	// Need to predefine due to scoping
	// var port int
	// var perr error
	// if p.AllowedMockServerPorts != "" {
	// 	port, perr = utils.FindPortInRange(p.AllowedMockServerPorts)
	// } else {
	// 	port, perr = utils.GetFreePort()
	// }
	// if perr != nil {
	// 	log.Println("[ERROR] unable to find free port, mockserver will fail to start")
	// }

	// if p.Server == nil && startMockServer {
	// 	log.Println("[DEBUG] starting mock service on port:", port)
	// 	args := []string{
	// 		"--pact-specification-version",
	// 		fmt.Sprintf("%d", p.SpecificationVersion),
	// 		"--pact-dir",
	// 		filepath.FromSlash(p.PactDir),
	// 		"--log",
	// 		filepath.FromSlash(p.LogDir + "/" + "pact.log"),
	// 		"--consumer",
	// 		p.Consumer,
	// 		"--provider",
	// 		p.Provider,
	// 		"--pact-file-write-mode",
	// 		p.PactFileWriteMode,
	// 	}

	// 	p.Server = p.pactClient.StartServer(args, port)
	// }

	native.Init()

	return p
}

// Configure logging
func (p *MockProvider) setupLogging() {
	if p.logFilter == nil {
		if p.LogLevel == "" {
			p.LogLevel = "INFO"
		}
		p.logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
			MinLevel: logutils.LogLevel(p.LogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(p.logFilter)
	}
	log.Println("[DEBUG] pact setup logging")
}

// Teardown stops the Pact Mock Server. This usually is called on completion
// of each test suite.
func (p *MockProvider) Teardown() error {
	log.Println("[DEBUG] teardown")
	if p.ServerPort != 0 {
		err := native.WritePactFile(p.ServerPort, p.PactDir)
		if err != nil {
			return err
		}

		if native.CleanupMockServer(p.ServerPort) {
			p.ServerPort = 0
		} else {
			log.Println("[DEBUG] unable to teardown server")
		}
	}
	return nil
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *MockProvider) Verify(integrationTest func(MockServerConfig) error) error {
	log.Println("[DEBUG] pact verify")
	p.Setup()

	// Start server
	serialisedPact := NewPactFile(p.Consumer, p.Provider, p.Interactions, p.SpecificationVersion)
	fmt.Println("[DEBUG] Sending pact file:", formatJSONObject(serialisedPact))

	// TODO: wire this better
	port := native.CreateMockServer(formatJSONObject(serialisedPact), "0.0.0.0:0", false)

	// TODO: not sure we want this?
	p.ServerPort = port

	// TODO: use cases for having server running post integration test?
	// Probably not...
	defer native.CleanupMockServer(port)

	// Run the integration test
	err := integrationTest(MockServerConfig{})

	if err != nil {
		return err
	}

	// Run Verification Process
	res, mismatches := native.Verify(p.ServerPort, p.PactDir)
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
	log.Println("[WARN] write pact file")
	if p.ServerPort != 0 {
		return native.WritePactFile(p.ServerPort, p.PactDir)
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
