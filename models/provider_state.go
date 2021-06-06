package models

// ProviderState Models a provider state coming over the Wire.
// This is generally provided as a request to an HTTP endpoint (e.g. PUT /state)
// to configure a state on a Provider.
type ProviderState struct {
	Consumer string   `json:"consumer"`
	State    string   `json:"state"`
	States   []string `json:"states"`
}

// ProviderStates is a mapping of consumers to all known states. This is usually
// a response from an HTTP endpoint (e.g. GET /states) to find all states a
// provider has.
type ProviderStates map[string][]string

// ProviderStateV3 allows parameters and a description to be passed to the verification process
type ProviderStateV3 struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"params,omitempty"`
}

// ProviderStateV3Response may return values in the state setup
// for the "value from provider state" feature
type ProviderStateV3Response map[string]interface{}

// StateHandler is a provider function that sets up a given state before
// the provider interaction is validated
// It can optionally return a map of key => value (JSON) that may be used
// as values in the verification process
// See https://github.com/pact-foundation/pact-reference/tree/master/rust/pact_verifier_cli#state-change-requests
// https://github.com/pact-foundation/pact-js/tree/feat/v3.0.0#provider-state-injected-values for more
type StateHandler func(setup bool, state ProviderStateV3) (ProviderStateV3Response, error)

// StateHandlers is a list of StateHandler's
type StateHandlers map[string]StateHandler
