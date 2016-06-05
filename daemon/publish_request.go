package daemon

// PublishRequest contains the details required to Publish Pacts to a broker.
type PublishRequest struct {
	// Array of local Pact files or directories containing them. Required.
	PactUrls []string

	// URL to fetch the provider states for the given provider API. Optional.
	PactBroker string

	// Username for Pact Broker basic authentication. Optional
	PactBrokerUsername string

	// Password for Pact Broker basic authentication. Optional
	PactBrokerPassword string
}
