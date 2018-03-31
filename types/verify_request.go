package types

import (
	"fmt"
	"log"
)

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

	// URL to retrieve valid Provider States.
	// Deprecation notice: no longer valid/required
	ProviderStatesURL string

	// URL to post currentp provider state to on the Provider API.
	ProviderStatesSetupURL string

	// Username when authenticating to a Pact Broker.
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	BrokerPassword string

	// PublishVerificationResults to the Pact Broker.
	PublishVerificationResults bool

	// ProviderVersion is the semantical version of the Provider API.
	ProviderVersion string

	// Verbose increases verbosity of output
	// Deprecated
	Verbose bool

	// CustomProviderHeaders are header to add to provider state set up
	// and pact verification `requests`. eg 'Authorization: Basic cGFjdDpwYWN0'.
	// NOTE: Use this feature very carefully, as anything in here is not captured
	// in the contract (e.g. time-bound tokens)
	CustomProviderHeaders []string

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
	} else {
		return fmt.Errorf("Pact URLs is mandatory")
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

	// Field is deprecated, leave here to see deprecation notice
	if v.ProviderStatesURL != "" {
		v.Args = append(v.Args, "--provider-states-url", v.ProviderStatesURL)
	}

	if v.BrokerUsername != "" {
		v.Args = append(v.Args, "--broker-username", v.BrokerUsername)
	}

	if v.BrokerPassword != "" {
		v.Args = append(v.Args, "--broker-password", v.BrokerPassword)
	}

	if v.ProviderVersion != "" {
		v.Args = append(v.Args, "--provider_app_version", v.ProviderVersion)
	}

	if v.PublishVerificationResults {
		v.Args = append(v.Args, "--publish_verification_results", "true")
	}

	if v.Verbose {
		log.Println("[DEBUG] verifier: ignoring deprecated Verbose flag")
	}
	return nil
}
