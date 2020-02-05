package types

import "fmt"

// ConsumerVersionSelector are the way we specify which pacticipants and
// versions we want to use when configuring verifications
// See https://docs.pact.io/selectors for more
type ConsumerVersionSelector struct {
	Pacticipant string `json:"pacticipant"`
	Version     string `json:"version"`
	Latest      bool   `json:"latest"`
	All         bool   `json:"all"`
}

func (c *ConsumerVersionSelector) Validate() error {
	if c.Pacticipant == "" {
		return fmt.Errorf("must provide a Pacticpant")
	}

	return nil
}
