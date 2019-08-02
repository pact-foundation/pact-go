package types

import (
	"errors"
	"fmt"
	"log"

	"github.com/pact-foundation/pact-go/proxy"
)

// Hook functions are used to tap into the lifecycle of a Consumer or Provider test
type Hook func() error

// VerifyRequest contains the verification params.
type VerifyRequest struct {
	// URL to hit during provider verification.
	ProviderBaseURL string

	// Local/HTTP paths to Pact files.
	PactURLs []string

	// Pact Broker URL for broker-based verification
	BrokerURL string

	// Tags to find in Broker for matrix-based testing
	Tags []string

	// ProviderStatesSetupURL is the endpoint to post current provider state
	// to on the Provider API.
	// Deprecated: For backward compatibility ProviderStatesSetupURL is
	// still supported. Use StateHandlers instead.
	ProviderStatesSetupURL string

	// Provider is the name of the Providing service.
	Provider string

	// Username when authenticating to a Pact Broker.
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	BrokerPassword string

	// BrokerToken is required when authenticating using the Bearer token mechanism
	BrokerToken string

	// PublishVerificationResults to the Pact Broker.
	PublishVerificationResults bool

	// ProviderVersion is the semantical version of the Provider API.
	ProviderVersion string

	// CustomProviderHeaders are headers to add during pact verification `requests`.
	// eg 'Authorization: Basic cGFjdDpwYWN0'.
	//
	// NOTE: Use this feature very carefully, as anything in here is not captured
	// in the contract (e.g. time-bound tokens)
	//
	// NOTE: This should be used very carefully and deliberately, as anything you do here
	// runs the risk of changing the contract and breaking the real system.
	CustomProviderHeaders []string

	// StateHandlers contain a mapped list of message states to functions
	// that are used to setup a given provider state prior to the message
	// verification step.
	StateHandlers StateHandlers

	// BeforeEach allows you to configure your provider prior to the individual test execution
	// e.g. setup temporary tokens, prepare data
	BeforeEach Hook

	// AfterEach allows you to configure your provider prior to the test execution
	// e.g. reset the database state
	AfterEach Hook

	// RequestFilter is a piece of middleware that will intercept requests/responses
	// from the provider in order to modify it. This is useful in situations where
	// you need to override a value due to time sensitivity - such as a OAuth Bearer
	// token.
	// NOTE: This should be used very carefully and deliberately, as anything you do here
	// runs the risk of changing the contract and breaking the real system.
	RequestFilter proxy.Middleware

	// Verbose increases verbosity of output
	// Deprecated
	Verbose bool

	// Arguments to the VerificationProvider
	// Deprecated: This will be deleted after the native library replaces Ruby deps.
	Args []string
}

// Validate checks that the minimum fields are provided.
// Deprecated: This map be deleted after the native library replaces Ruby deps,
// and should not be used outside of this library.
func (v *VerifyRequest) Validate() error {
	v.Args = []string{}

	if len(v.PactURLs) != 0 {
		v.Args = append(v.Args, v.PactURLs...)
	}

	if len(v.PactURLs) == 0 && v.BrokerURL == "" {
		return fmt.Errorf("One of 'PactURLs' or 'BrokerURL' must be specified")
	}

	if len(v.CustomProviderHeaders) != 0 {
		for _, header := range v.CustomProviderHeaders {
			v.Args = append(v.Args, "--custom-provider-header", header)
		}
	}

	v.Args = append(v.Args, "--format", "json")

	if v.ProviderBaseURL != "" {
		v.Args = append(v.Args, "--provider-base-url", v.ProviderBaseURL)
	} else {
		return fmt.Errorf("Provider base URL is mandatory")
	}

	if v.ProviderStatesSetupURL != "" {
		v.Args = append(v.Args, "--provider-states-setup-url", v.ProviderStatesSetupURL)
	}

	if v.BrokerUsername != "" {
		v.Args = append(v.Args, "--broker-username", v.BrokerUsername)
	}

	if v.BrokerPassword != "" {
		v.Args = append(v.Args, "--broker-password", v.BrokerPassword)
	}

	if v.BrokerURL != "" && ((v.BrokerUsername == "" && v.BrokerPassword != "") || (v.BrokerUsername != "" && v.BrokerPassword == "")) {
		return errors.New("both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given")
	}

	if v.BrokerURL != "" {
		v.Args = append(v.Args, "--pact-broker-base-url", v.BrokerURL)
	}

	if v.BrokerToken != "" {
		v.Args = append(v.Args, "--broker-token", v.BrokerToken)
	}

	if v.ProviderVersion != "" {
		v.Args = append(v.Args, "--provider_app_version", v.ProviderVersion)
	}

	if v.Provider != "" {
		v.Args = append(v.Args, "--provider", v.Provider)
	}

	if v.PublishVerificationResults {
		v.Args = append(v.Args, "--publish_verification_results", "true")
	}

	if v.Verbose {
		log.Println("[DEBUG] verifier: ignoring deprecated Verbose flag")
	}

	for _, tag := range v.Tags {
		v.Args = append(v.Args, "--consumer-version-tag", tag)
	}

	return nil
}
