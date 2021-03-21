// Package v3 contains the main Pact DSL used in the Consumer
// collaboration test cases, and Provider contract test verification.
package v3

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

	"github.com/pact-foundation/pact-go/utils"
	"github.com/pact-foundation/pact-go/v3/internal/native/mockserver"
	native "github.com/pact-foundation/pact-go/v3/internal/native/mockserver"
)

func init() {
	initLogging()
}

// QueryStringStyle allows a user to specific the v2 query string serialisation format
// Different frameworks have different ways to serialise this, which is why
// v3 moved to storing this as a map
type QueryStringStyle int

const (
	// Default uses the param=value1&param=value2 style
	Default QueryStringStyle = iota

	// AlwaysArray uses the [] style even if a parameter only has a single value
	// e.g. param[]=value1&param[]=value2
	AlwaysArray

	// Array uses the [] style only if a parameter only has multiple values
	// e.g.  param[]=value1&param[]=value2&param2=value
	Array
)

// PactSerialisationOptionsV2 allows a user to override specific pact serialisation options
type PactSerialisationOptionsV2 struct {
	QueryStringStyle QueryStringStyle
}

type mockHTTPProviderConfig struct {
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

	matchingConfig PactSerialisationOptionsV2

	// Check if CLI tools are up to date
	toolValidityCheck bool

	// TLS enables a mock service behind a self-signed certificate
	// TODO: document and test this
	TLS bool
}

// MockHTTPProviderConfigV2 configures a V2 Pact HTTP Mock Provider
type MockHTTPProviderConfigV2 = mockHTTPProviderConfig

// MockHTTPProviderConfigV3 configures a V3 Pact HTTP Mock Provider
type MockHTTPProviderConfigV3 = mockHTTPProviderConfig

// httpMockProvider is the entrypoint for http consumer tests and provides the base capability for the
// exported types HTTPMockProviderV2 and HTTPMockProviderV3
type httpMockProvider struct {
	specificationVersion SpecificationVersion
	config               mockHTTPProviderConfig

	v2Interactions []*InteractionV2
	v3Interactions []*InteractionV3

	// fsm state of the interaction
	state      string
	mockserver *mockserver.MockServer
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

	// if p.config.Network == "" {
	// 	p.config.Network = "tcp"
	// }

	if p.config.Host == "" {
		p.config.Host = "127.0.0.1"
	}

	if p.config.LogDir == "" {
		p.config.LogDir = fmt.Sprintf(filepath.Join(dir, "logs"))
	}

	if p.config.PactDir == "" {
		p.config.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
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

	p.mockserver = &mockserver.MockServer{}
	p.mockserver.Init()

	return nil
}

func (p *httpMockProvider) cleanInteractions() {
	p.v2Interactions = make([]*InteractionV2, 0)
	p.v3Interactions = make([]*InteractionV3, 0)
}

// ExecuteTest runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite
// and write the pact file if successful
func (p *httpMockProvider) ExecuteTest(integrationTest func(MockServerConfig) error) error {
	log.Println("[DEBUG] pact verify")

	// Generate interactions for Pact file
	var serialisedPact interface{}
	if p.specificationVersion == V2 {
		serialisedPact = newPactFileV2(p.config.Consumer, p.config.Provider, p.v2Interactions, p.config.matchingConfig)
	} else {
		serialisedPact = newPactFileV3(p.config.Consumer, p.config.Provider, p.v3Interactions, nil)
	}

	log.Println("[DEBUG] Sending pact file:", formatJSONObject(serialisedPact))

	// Clean interactions
	p.cleanInteractions()

	port, err := p.mockserver.CreateMockServer(formatJSONObject(serialisedPact), fmt.Sprintf("%s:%d", p.config.Host, p.config.Port), p.config.TLS)
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

	if err != nil {
		return err
	}

	// Run Verification Process
	// res, mismatches := p.mockserver.Verify(p.config.Port, p.config.PactDir)
	_, mismatches := p.mockserver.Verify(p.config.Port, p.config.PactDir)
	p.displayMismatches(mismatches)

	// if !res {
	// 	return fmt.Errorf("pact validation failed: %+v %+v", res, mismatches)
	// }
	if len(mismatches) > 0 {
		return fmt.Errorf("pact validation failed: %+v", mismatches)
	}

	return p.WritePact()
}

// TODO: pretty print this to make it really easy to understand the problems
// See existing Pact/Ruby code examples
func (p *httpMockProvider) displayMismatches(mismatches []mockserver.MismatchedRequest) {
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

// WritePact should be called when all tests have been performed for a
// given Consumer <-> Provider pair. It will write out the Pact to the
// configured file. This is safe to call multiple times as the service is smart
// enough to merge pacts and avoid duplicates.
func (p *httpMockProvider) WritePact() error {
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
