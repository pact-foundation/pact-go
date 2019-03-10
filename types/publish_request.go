package types

import (
	"errors"
	"fmt"
)

// PublishRequest contains the details required to Publish Pacts to a broker.
type PublishRequest struct {
	// Array of local Pact files or directories containing them. Required.
	PactURLs []string

	// URL to fetch the provider states for the given provider API. Optional.
	PactBroker string

	// Username for Pact Broker basic authentication. Optional
	BrokerUsername string

	// Password for Pact Broker basic authentication. Optional
	BrokerPassword string

	// BrokerToken is required when authenticating using the Bearer token mechanism
	BrokerToken string

	// ConsumerVersion is the semantical version of the consumer API.
	ConsumerVersion string

	// Tags help you organise your Pacts for different testing purposes.
	// e.g. "production", "master" and "development" are some common examples.
	Tags []string

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
func (p *PublishRequest) Validate() error {
	p.Args = []string{}

	if len(p.PactURLs) != 0 {
		p.Args = append(p.Args, p.PactURLs...)
	} else {
		return fmt.Errorf("'PactURLs' is mandatory")
	}

	if p.BrokerUsername != "" {
		p.Args = append(p.Args, "--broker-username", p.BrokerUsername)
	}

	if p.BrokerPassword != "" {
		p.Args = append(p.Args, "--broker-password", p.BrokerPassword)
	}

	if p.PactBroker != "" && ((p.BrokerUsername == "" && p.BrokerPassword != "") || (p.BrokerUsername != "" && p.BrokerPassword == "")) {
		return errors.New("both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given")
	}

	if p.PactBroker == "" {
		return fmt.Errorf("'PactBroker' is mandatory")
	}
	p.Args = append(p.Args, "--broker-base-url", p.PactBroker)

	if p.BrokerToken != "" {
		p.Args = append(p.Args, "--broker-token", p.BrokerToken)
	}

	if p.ConsumerVersion == "" {
		return fmt.Errorf("'ConsumerVersion' is mandatory")
	}
	p.Args = append(p.Args, "--consumer-app-version", p.ConsumerVersion)

	if len(p.Tags) > 0 {
		for _, t := range p.Tags {
			p.Args = append(p.Args, "--tag", t)
		}
	}

	if p.Verbose {
		p.Args = append(p.Args, "--verbose")
	}

	return nil
}
