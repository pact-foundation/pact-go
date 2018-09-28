package types

import "encoding/json"

// PactMessageRequest contains the response from the Pact Message
// CLI execution.
type PactMessageRequest struct {

	// Message is the object to be marshalled to JSON
	Message interface{}

	// Consumer is the name of the message consumer
	Consumer string

	// Provider is the name of the message provider
	Provider string

	// PactDir is the location of where pacts should be stored
	PactDir string

	// Args are the arguments sent to to the message service
	Args []string
}

// Validate checks all things are well and constructs
// the CLI args to the message service
func (m *PactMessageRequest) Validate() error {
	m.Args = []string{}

	body, err := json.Marshal(m.Message)
	if err != nil {
		return err
	}

	m.Args = append(m.Args, []string{
		"update",
		string(body),
		"--consumer",
		m.Consumer,
		"--provider",
		m.Provider,
		"--pact-dir",
		m.PactDir,
		"--pact-specification-version",
		"3",
	}...)

	return nil
}
