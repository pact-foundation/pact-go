package provider

// ConsumerVersionSelector are the way we specify which pacticipants and
// versions we want to use when configuring verifications
// See https://docs.pact.io/selectors for more
//
// Where a new selector is available in the broker but not yet supported here,
// you may use the UntypedConsumerVersionSelector to pass in arbitrary key/values
//
// Definitive list: https://github.com/pact-foundation/pact_broker/blob/master/spec/lib/pact_broker/api/contracts/verifiable_pacts_json_query_schema_combinations_spec.rb
type ConsumerVersionSelector struct {
	Tag                string `json:"tag,omitempty"`
	FallbackTag        string `json:"fallbackTag,omitempty"`
	Latest             bool   `json:"latest,omitempty"`
	Consumer           string `json:"consumer,omitempty"`
	DeployedOrReleased bool   `json:"deployedOrReleased,omitempty"`
	Deployed           bool   `json:"deployed,omitempty"`
	Released           bool   `json:"released,omitempty"`
	Environment        string `json:"environment,omitempty"`
	MainBranch         bool   `json:"mainBranch,omitempty"`
	MatchingBranch     bool   `json:"matchingBranch,omitempty"`
	Branch             string `json:"branch,omitempty"`
}

// Type marker
func (c *ConsumerVersionSelector) IsSelector() {
}

type UntypedConsumerVersionSelector map[string]interface{}

// Type marker
func (c *UntypedConsumerVersionSelector) IsSelector() {
}

type Selector interface {
	IsSelector()
}
