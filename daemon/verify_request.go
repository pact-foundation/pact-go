package daemon

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

	args []string
}

// Validate checks that the minimum fields are provided.
func (v *VerifyRequest) Validate() error {
	v.args = []string{}
	if v.ProviderBaseURL != "" {
		v.args = append(v.args, fmt.Sprintf("--provider-base-url %s", v.ProviderBaseURL))
	} else {
		return fmt.Errorf("ProviderBaseURL is mandatory.")
	}

	if len(v.PactURLs) != 0 {
		v.args = append(v.args, fmt.Sprintf("--pact-urls %s", strings.Join(v.PactURLs[:], ",")))
	} else {
		return fmt.Errorf("PactURLs is mandatory.")
	}

	if v.ProviderStatesSetupURL != "" {
		v.args = append(v.args, fmt.Sprintf("--provider-states-setup-url %s", v.ProviderStatesSetupURL))
	}

	if v.ProviderStatesURL != "" {
		v.args = append(v.args, fmt.Sprintf("--provider-states-url %s", v.ProviderStatesURL))
	}

	if v.BrokerUsername != "" {
		v.args = append(v.args, fmt.Sprintf("--broker-username %s", v.BrokerUsername))
	}

	if v.BrokerPassword != "" {
		v.args = append(v.args, fmt.Sprintf("--broker-password %s", v.BrokerPassword))
	}
	return nil
}
