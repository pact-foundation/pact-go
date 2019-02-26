package types

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
