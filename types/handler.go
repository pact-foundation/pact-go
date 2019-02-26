package types

// StateHandler is a provider function that sets up a given state before
// the provider interaction is validated
type StateHandler func() error

// StateHandlers is a list of StateHandler's
type StateHandlers map[string]StateHandler

// State specifies how the system should be configured when
// verified. e.g. "user A exists"
type State struct {
	Name string `json:"name"`
}
