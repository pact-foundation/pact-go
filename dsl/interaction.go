package dsl

import (
	"encoding/json"
	"log"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Request
	Request Request `json:"request"`

	// Response
	Response Response `json:"response"`

	// Description to be written into the Pact file
	Description string `json:"description"`

	// Provider state to be written into the Pact file
	State string `json:"providerState,omitempty"`
}

// Given specifies a provider state. Optional.
func (p *Interaction) Given(state string) *Interaction {
	p.State = state
	return p
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (p *Interaction) UponReceiving(description string) *Interaction {
	p.Description = description
	return p
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (p *Interaction) WithRequest(request Request) *Interaction {
	p.Request = request

	// Need to fix any weird JSON marshalling issues with the body Here
	// If body is a string, not an object, we need to put it back into an object
	// so that it's not double encoded
	switch content := request.Body.(type) {
	case string:
		p.Request.Body = toObject([]byte(content))
	default:
		// leave alone
	}

	return p
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
func (p *Interaction) WillRespondWith(response Response) *Interaction {
	p.Response = response

	// Need to fix any weird JSON marshalling issues with the body Here
	// If body is a string, not an object, we need to put it back into an object
	// so that it's not double encoded
	switch content := response.Body.(type) {
	case string:
		p.Response.Body = toObject([]byte(content))
	default:
		// leave alone
	}

	return p
}

// Takes a string body and converts it to an interface{} representation.
func toObject(content []byte) interface{} {
	var obj interface{}
	err := json.Unmarshal(content, &obj)
	if err != nil {
		log.Println("[DEBUG] interaction: error unmarshaling object into string:", err.Error())
		return string(content)
	}

	return obj
}
