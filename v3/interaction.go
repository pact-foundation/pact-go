package v3

import (
	"log"

	"github.com/pact-foundation/pact-go/v3/internal/native/mockserver"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Request
	Request Request `json:"request"`

	// Response
	Response Response `json:"response"`

	// Description to be written into the Pact file
	Description string `json:"description"`

	// Reference to the native rust handle
	interaction *mockserver.Interaction
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *Interaction) UponReceiving(description string) *Interaction {
	i.Description = description
	i.interaction.UponReceiving(description)

	return i
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithRequest(request Request) *Interaction {
	i.Request = request

	i.interaction.WithRequest(request.Method, request.Path)
	query := make(map[string][]interface{})
	for k, values := range request.Query {
		for _, v := range values {
			query[k] = append(query[k], v)
		}
	}
	i.interaction.WithQuery(query)
	headers := make(map[string]interface{})
	for k, v := range request.Headers {
		headers[k] = v
	}
	i.interaction.WithRequestHeaders(headers)
	i.interaction.WithJSONRequestBody(request.Body)

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
// Defaults to application/json.
// Use WillResponseWithContent to define custom type
func (i *Interaction) WillRespondWith(response Response) *Interaction {
	return i.WillRespondWithContent("application/json", response)
}

func (i *Interaction) WillRespondWithContent(contentType string, response Response) *Interaction {
	i.Response = response
	headers := make(map[string]interface{})
	for k, v := range response.Headers {
		headers[k] = v.(stringLike).string()
	}
	i.interaction.WithResponseHeaders(headers)
	i.interaction.WithStatus(response.Status)

	if contentType == "application/json" {
		i.interaction.WithJSONResponseBody(response.Body)
	} else {
		bodyStr, ok := response.Body.(string)
		if !ok {
			panic("response body must be a string")
		}
		i.interaction.WithResponseBody(bodyStr, contentType)
	}

	return i
}
