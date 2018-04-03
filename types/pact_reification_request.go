package types

import "encoding/json"

// PactReificationRequest contains the response from the Pact Message
// CLI execution.
type PactReificationRequest struct {

	// Message is the object to be marshalled to JSON
	Message interface{}

	// Args are the arguments sent to to the message service
	Args []string
}

// Validate checks all things are well and constructs
// the CLI args to the message service
func (m *PactReificationRequest) Validate() error {
	m.Args = []string{}

	body, err := json.Marshal(m.Message)
	if err != nil {
		return err
	}

	m.Args = append(m.Args, []string{
		"reify",
		string(body),
	}...)

	return nil
}
