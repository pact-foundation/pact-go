package types

import (
	"fmt"
	"strings"
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
	if v.ProviderBaseURL != "" {
		v.Args = append(v.Args, "--provider-base-url")
		v.Args = append(v.Args, v.ProviderBaseURL)
	} else {
		return fmt.Errorf("ProviderBaseURL is mandatory.")
	}

	if len(v.PactURLs) != 0 {
		v.Args = append(v.Args, "--pact-urls")
		v.Args = append(v.Args, strings.Join(v.PactURLs[:], ","))
	} else {
		return fmt.Errorf("PactURLs is mandatory.")
	}

	if v.ProviderStatesSetupURL != "" {
		v.Args = append(v.Args, "--provider-states-setup-url")
		v.Args = append(v.Args, v.ProviderStatesSetupURL)
	}

	if v.ProviderStatesURL != "" {
		v.Args = append(v.Args, "--provider-states-url")
		v.Args = append(v.Args, v.ProviderStatesURL)
	}

	if v.BrokerUsername != "" {
		v.Args = append(v.Args, "--broker-username")
		v.Args = append(v.Args, v.BrokerUsername)
	}

	if v.BrokerPassword != "" {
		v.Args = append(v.Args, "--broker-password")
		v.Args = append(v.Args, v.BrokerPassword)
	}

	if v.ProviderVersion != "" {
		v.Args = append(v.Args, "--provider_app_version")
		v.Args = append(v.Args, v.ProviderVersion)
	}

	if v.PublishVerificationResults {
		v.Args = append(v.Args, "--publish_verification_results")
		v.Args = append(v.Args, "true")
	}

	if v.Verbose {
		v.Args = append(v.Args, "--verbose")
		v.Args = append(v.Args, fmt.Sprintf("%v", v.Verbose))
	}
	return nil
}
