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
func (i *Interaction) Given(state string) *Interaction {
	i.State = state

	return i
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *Interaction) UponReceiving(description string) *Interaction {
	i.Description = description

	return i
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithRequest(request Request) *Interaction {
	i.Request = request

	// Check if someone tried to add an object as a string representation
	// as per original allowed implementation, e.g.
	// { "foo": "bar", "baz": like("bat") }
	if isJSONFormattedObject(request.Body) {
		log.Println("[WARN] request body appears to be a JSON formatted object, " +
			"no structural matching will occur. Support for structured strings has been" +
			"deprecated as of 0.13.0")
	}

	return i
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
func (i *Interaction) WillRespondWith(response Response) *Interaction {
	i.Response = response

	return i
}

// Checks to see if someone has tried to submit a JSON string
// for an object, which is no longer supported
func isJSONFormattedObject(stringOrObject interface{}) bool {
	switch content := stringOrObject.(type) {
	case []byte:
	case string:
		var obj interface{}
		err := json.Unmarshal([]byte(content), &obj)

		if err != nil {
			return false
		}

		// Check if a map type
		if _, ok := obj.(map[string]interface{}); ok {
			return true
		}
	}

	return false
}
