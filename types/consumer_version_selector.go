package types

// ConsumerVersionSelector are the way we specify which pacticipants and
// versions we want to use when configuring verifications
// See https://docs.pact.io/selectors for more
// Definitive list: https://github.com/pact-foundation/pact_broker/blob/master/spec/lib/pact_broker/api/contracts/verifiable_pacts_json_query_schema_combinations_spec.rb
type ConsumerVersionSelector struct {
	Pacticipant string `json:"-"` // Deprecated
	Version     string `json:"-"` // Deprecated
	All         bool   `json:"-"` // Deprecated

	Tag                string `json:"tag"`
	FallbackTag        string `json:"fallbackTag"`
	Latest             bool   `json:"latest"`
	Consumer           string `json:"consumer"`
	DeployedOrReleased bool   `json:"deployedOrReleased"`
	Deployed           bool   `json:"deployed"`
	Released           bool   `json:"released"`
	Environment        string `json:"environment"`
	MainBranch         bool   `json:"mainBranch"`
	Branch             string `json:"branch"`
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
