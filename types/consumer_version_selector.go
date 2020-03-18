package types

import "fmt"

// ConsumerVersionSelector are the way we specify which pacticipants and
// versions we want to use when configuring verifications
// See https://docs.pact.io/selectors for more
type ConsumerVersionSelector struct {
	Pacticipant string `json:"pacticipant"`
	Tag         string `json:"tag"`
	Version     string `json:"version"`
	Latest      bool   `json:"latest"`
	All         bool   `json:"all"`
}

// Validate the selector configuration
func (c *ConsumerVersionSelector) Validate() error {
	if c.All && c.Pacticipant == "" {
		return fmt.Errorf("must provide a Pacticpant")
	}

	if c.Pacticipant != "" && c.Tag == "" {
		return fmt.Errorf("must provide at least a Tag if Pacticpant specified")
	}

	if c.All && c.Latest {
		return fmt.Errorf("cannot select both All and Latest")
	}

	return nil
}
