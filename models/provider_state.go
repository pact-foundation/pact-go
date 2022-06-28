package models

// ProviderState allows parameters and a description to be passed to the verification process
type ProviderState struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"params,omitempty"`
}

// ProviderStateResponse may return values in the state setup
// for the "value from provider state" feature
type ProviderStateResponse map[string]interface{}

// StateHandler is a provider function that sets up a given state before
// the provider interaction is validated
// It can optionally return a map of key => value (JSON) that may be used
// as values in the verification process
// See https://github.com/pact-foundation/pact-reference/tree/master/rust/pact_verifier_cli#state-change-requests
// https://github.com/pact-foundation/pact-js/tree/feat/v3.0.0#provider-state-injected-values for more
type StateHandler func(setup bool, state ProviderState) (ProviderStateResponse, error)

// StateHandlers is a list of StateHandler's
type StateHandlers map[string]StateHandler
