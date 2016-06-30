package types

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

	// ConsumerVersion is the semantical version of the consumer API.
	ConsumerVersion string

	// Tags help you organise your Pacts for different testing purposes.
	// e.g. "production", "latest" and "development" are some common examples.
	Tags []string
}
