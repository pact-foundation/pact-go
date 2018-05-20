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

	// PactFileWriteMode specifies how to write to the Pact file, for the life
	// of a Mock Service.
	// "overwrite" will always truncate and replace the pact after each run
	// "update" will append to the pact file, which is useful if your tests
	// are split over multiple files and instantiations of a Mock Server
	// See https://github.com/pact-foundation/pact-ruby/blob/master/documentation/configuration.md#pactfile_write_mode
	PactFileWriteMode string

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
		m.PactFileWriteMode,
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
