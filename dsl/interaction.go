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
	p.Request.Body = toObject(request.Body)

	return p
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
func (p *Interaction) WillRespondWith(response Response) *Interaction {
	p.Response = response

	// Need to fix any weird JSON marshalling issues with the body Here
	// If body is a string, not an object, we need to put it back into an object
	// so that it's not double encoded
	p.Response.Body = toObject(response.Body)

	return p
}

// Takes a string body and converts it to an interface{} representation.
func toObject(stringOrObject interface{}) interface{} {

	switch content := stringOrObject.(type) {
	case []byte:
	case string:
		var obj interface{}
		err := json.Unmarshal([]byte(content), &obj)

		if err != nil {
			log.Printf("[DEBUG] interaction: error unmarshaling string '%v' into an object. Probably not an object: %v\n", stringOrObject, err.Error())
			return content
		}

		return obj
	default:
		// leave alone
	}

	return stringOrObject
}
