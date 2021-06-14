// package consumer contains the main Pact DSL used in the Consumer
// collaboration test cases, and Provider contract test verification.
package consumer

// TODO: setup a proper state machine to prevent actions
// Current issues
// 1. Setup needs to be initialised to get a port -> should be resolved by creating the server at the point of verification
// 2. Ensure that interactions are properly cleared
// 3. Need to ensure only v2 or v3 matchers are added

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	native "github.com/pact-foundation/pact-go/v2/internal/native/mockserver"
	logging "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

func init() {
	logging.InitLogging()
}

// MockHTTPProviderConfig provides the configuration options for an HTTP mock server
// consumer test.
type MockHTTPProviderConfig struct {
	// Consumer is the name of the Consumer/Client.
	Consumer string

	// Provider is the name of the Providing service.
	Provider string

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
	// Network string

	// Ports MockServer can be deployed to, can be CSV or Range with a dash
	// Example "1234", "12324,5667", "1234-5667"
	AllowedMockServerPorts string

	// Port the mock service should run on. Leave blank to have one assigned
	// automatically by the OS.
	// Use AllowedMockServerPorts to constrain the assigned range.
	Port int

	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration

	// TLS enables a mock service behind a self-signed certificate
	// TODO: document and test this
	TLS bool
}

// httpMockProvider is the entrypoint for http consumer tests
// This object is not thread safe
type httpMockProvider struct {
	specificationVersion models.SpecificationVersion
	config               MockHTTPProviderConfig
	mockserver           *native.MockServer
}

// MockServerConfig stores the address configuration details of the server for the current executing test
// This is most useful for the use of OS assigned, dynamic ports and parallel tests
type MockServerConfig struct {
	Port      int
	Host      string
	TLSConfig *tls.Config
}

// validateConfig validates the configuration for the consumer test
func (p *httpMockProvider) validateConfig() error {
	log.Println("[DEBUG] pact setup")
	dir, _ := os.Getwd()

	if p.config.Host == "" {
		p.config.Host = "127.0.0.1"
	}

	if p.config.LogDir == "" {
		p.config.LogDir = filepath.Join(dir, "logs")
	}

	if p.config.PactDir == "" {
		p.config.PactDir = filepath.Join(dir, "pacts")
	}

	if p.config.ClientTimeout == 0 {
		p.config.ClientTimeout = 10 * time.Second
	}

	var pErr error
	if p.config.AllowedMockServerPorts != "" && p.config.Port <= 0 {
		p.config.Port, pErr = utils.FindPortInRange(p.config.AllowedMockServerPorts)
	} else if p.config.Port <= 0 {
		p.config.Port, pErr = utils.GetFreePort()
	}

	if pErr != nil {
		return fmt.Errorf("error: unable to find free port, mock server will fail to start")
	}

	p.mockserver = native.NewHTTPMockServer(p.config.Consumer, p.config.Provider)
	switch p.specificationVersion {
	case models.V2:
		p.mockserver.WithSpecificationVersion(native.SPECIFICATION_VERSION_V2)
	case models.V3:
		p.mockserver.WithSpecificationVersion(native.SPECIFICATION_VERSION_V3)
	}
	native.Init()

	return nil
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (p *httpMockProvider) ExecuteTest(integrationTest func(MockServerConfig) error) error {
	log.Println("[DEBUG] pact verify")

	port, err := p.mockserver.Start(fmt.Sprintf("%s:%d", p.config.Host, p.config.Port), p.config.TLS)
	defer p.mockserver.CleanupMockServer(p.config.Port)
	if err != nil {
		return err
	}

	// Run the integration test
	err = integrationTest(MockServerConfig{
		Port:      port,
		Host:      p.config.Host,
		TLSConfig: GetTLSConfigForTLSMockServer(),
	})

	res, mismatches := p.mockserver.Verify(p.config.Port, p.config.PactDir)
	p.displayMismatches(mismatches)

	if err != nil {
		return err
	}

	if !res {
		return fmt.Errorf("pact validation failed: %+v %+v", res, mismatches)
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("pact validation failed: %+v", mismatches)
	}

	return p.writePact()
}

// TODO: pretty print this to make it really easy to understand the problems
// See existing Pact/Ruby code examples
func (p *httpMockProvider) displayMismatches(mismatches []native.MismatchedRequest) {
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

// writePact may be called after each interaction with a mock server is completed
// the shared core is threadsafe and will merge, as long as the requests come from a single process
// (that is, there isn't separate) instances of the FFI running simultaneously
func (p *httpMockProvider) writePact() error {
	log.Println("[DEBUG] write pact file")
	if p.config.Port != 0 {
		return p.mockserver.WritePactFile(p.config.Port, p.config.PactDir)
	}
	return errors.New("pact server not yet started")
}

// GetTLSConfigForTLSMockServer gets an http transport with
// the certificates already trusted. Alternatively, simply set
// trust level to insecure
func GetTLSConfigForTLSMockServer() *tls.Config {
	return native.GetTLSConfig()
}
