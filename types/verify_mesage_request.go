package types

import (
	"fmt"
)

// VerifyMessageRequest contains the verification params.
// TODO: make this CLI "request" type an Interface (e.g. Validate())
//       also make the core of it embeddable to be re-used
type VerifyMessageRequest struct {
	// Local/HTTP paths to Pact files.
	PactURLs []string

	// Pact Broker URL for broker-based verification
	BrokerURL string

	// Tags to find in Broker for matrix-based testing
	Tags []string

	// Username when authenticating to a Pact Broker.
	BrokerUsername string

	// Password when authenticating to a Pact Broker.
	BrokerPassword string

	// PublishVerificationResults to the Pact Broker.
	PublishVerificationResults bool

	// ProviderVersion is the semantical version of the Provider API.
	ProviderVersion string

	// Arguments to the VerificationProvider
	// Deprecated: This will be deleted after the native library replaces Ruby deps.
	Args []string
}

// Validate checks that the minimum fields are provided.
// Deprecated: This map be deleted after the native library replaces Ruby deps,
// and should not be used outside of this library.
func (v *VerifyMessageRequest) Validate() error {
	v.Args = []string{}

	if len(v.PactURLs) != 0 {
		v.Args = append(v.Args, v.PactURLs...)
	} else {
		return fmt.Errorf("Pact URLs is mandatory")
	}

	v.Args = append(v.Args, "--format", "json")

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

	return nil
}
