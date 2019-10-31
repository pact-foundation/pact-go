package dsl

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/types"
)

// PactFile is a simple representation of a Pact file to be able to
// parse Consumer/Provider from the file.
type PactFile struct {
	// The API Consumer name
	Consumer PactName `json:"consumer"`

	// The API Provider name
	Provider PactName `json:"provider"`
}

// PactName represents the name fields in the PactFile.
type PactName struct {
	Name string `json:"name"`
}

// Publisher is the API to send Pact files to a Pact Broker.
type Publisher struct {
	pactClient Client

	// Log levels.
	LogLevel string

	// Used to detect if logging has been configured.
	logFilter *logutils.LevelFilter
}

// Publish sends the Pacts to a broker, optionally tagging them
func (p *Publisher) Publish(request types.PublishRequest) error {
	p.setupLogging()
	log.Println("[DEBUG] pact publisher: publish pact")

	if p.pactClient == nil {
		c := NewClient()
		p.pactClient = c
	}

	err := request.Validate()

	if err != nil {
		return err
	}

	return p.pactClient.PublishPacts(request)
}

// Configure logging
func (p *Publisher) setupLogging() {
	if p.logFilter == nil {
		if p.LogLevel == "" {
			p.LogLevel = "INFO"
		}
		p.logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
			MinLevel: logutils.LogLevel(p.LogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(p.logFilter)
	}
	log.Println("[DEBUG] pact setup logging")
}
