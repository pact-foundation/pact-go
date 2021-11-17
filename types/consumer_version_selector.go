package types

import (
	"fmt"
	"log"
)

// ConsumerVersionSelector are the way we specify which pacticipants and
// versions we want to use when configuring verifications
// See https://docs.pact.io/selectors for more
type ConsumerVersionSelector struct {
	Pacticipant        string `json:"-"` // Deprecated
	All                bool   `json:"-"` // Deprecated
	Version            string `json:"-"` // Deprecated
	Tag                string `json:"tag,omitempty"`
	FallbackTag        string `json:"fallbackTag,omitempty"`
	Latest             bool   `json:"latest,omitempty"`
	Consumer           string `json:"consumer,omitempty"`
	DeployedOrReleased bool   `json:"deployedOrReleased,omitempty"`
	Deployed           bool   `json:"deployed,omitempty"`
	Released           bool   `json:"released,omitempty"`
	Environment        string `json:"environment,omitempty"`
	MainBranch         bool   `json:"mainBranch,omitempty"`
	Branch             string `json:"branch,omitempty"`
	MatchingBranch     bool   `json:"matchingBranch,omitempty"`
}

// Validate the selector configuration
func (c *ConsumerVersionSelector) Validate() error {
	if c.All && c.Latest {
		return fmt.Errorf("cannot select both All and Latest")
	}

	if c.All {
		c.Latest = false
	}

	if c.Pacticipant != "" && c.Consumer != "" {
		return fmt.Errorf("cannot select deprecated field 'Pacticipant' along with Consumer, use only 'Consumer'")
	}

	if c.Pacticipant != "" {
		c.Consumer = c.Pacticipant
		log.Println("[WARN] 'Pacticipant' is deprecated, please use 'Consumer'. 'Consumer' has been automatically set to", c.Pacticipant)
	}

	if c.Version != "" {
		log.Println("[WARN] 'Version' is deprecated and has no effect")
	}

	return nil
}
