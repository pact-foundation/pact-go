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

	// URL to retrieve valid Provider States.
	ProviderStatesURL string

	// URL to post currentp provider state to on the Provider API.
	ProviderStatesSetupURL string

	// Username when authenticating to a Pact Broker.
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	BrokerPassword string

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
		v.Args = append(v.Args, fmt.Sprintf("--provider-base-url %s", v.ProviderBaseURL))
	} else {
		return fmt.Errorf("ProviderBaseURL is mandatory.")
	}

	if len(v.PactURLs) != 0 {
		v.Args = append(v.Args, fmt.Sprintf("--pact-urls %s", strings.Join(v.PactURLs[:], ",")))
	} else {
		return fmt.Errorf("PactURLs is mandatory.")
	}

	if v.ProviderStatesSetupURL != "" {
		v.Args = append(v.Args, fmt.Sprintf("--provider-states-setup-url %s", v.ProviderStatesSetupURL))
	}

	if v.ProviderStatesURL != "" {
		v.Args = append(v.Args, fmt.Sprintf("--provider-states-url %s", v.ProviderStatesURL))
	}

	if v.BrokerUsername != "" {
		v.Args = append(v.Args, fmt.Sprintf("--broker-username %s", v.BrokerUsername))
	}

	if v.BrokerPassword != "" {
		v.Args = append(v.Args, fmt.Sprintf("--broker-password %s", v.BrokerPassword))
	}
	return nil
}
