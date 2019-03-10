package dsl

import (
	"log"

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
}

// Publish sends the Pacts to a broker, optionally tagging them
func (p *Publisher) Publish(request types.PublishRequest) error {
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
