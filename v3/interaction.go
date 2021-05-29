package v3

import (
	"log"

	"github.com/pact-foundation/pact-go/v3/internal/native/mockserver"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Reference to the native rust handle
	interaction *mockserver.Interaction
}

type InteractionRequest struct {
	// Reference to the native rust handle
	interaction *mockserver.Interaction
}

type InteractionResponse struct {
	// Reference to the native rust handle
	interaction *mockserver.Interaction
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *Interaction) UponReceiving(description string) *Interaction {
	i.interaction.UponReceiving(description)

	return i
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithRequest(method Method, path Matcher) *InteractionRequest {
	i.interaction.WithRequest(string(method), path)

	return &InteractionRequest{
		interaction: i.interaction,
	}
}

func (i *InteractionRequest) Method(method Method) *InteractionRequest {

	return i
}

func (i *InteractionRequest) Query(query QueryMatcher) *InteractionRequest {
	q := make(map[string][]interface{})
	for k, values := range query {
		for _, v := range values {
			q[k] = append(q[k], v)
		}
	}
	i.interaction.WithQuery(q)

	return i
}

func (i *InteractionRequest) Headers(headers HeadersMatcher) *InteractionRequest {
	h := make(map[string]interface{})
	for k, v := range headers {
		h[k] = v
	}
	i.interaction.WithRequestHeaders(h)

	return i
}

func (i *InteractionRequest) JSON(body interface{}) *InteractionRequest {
	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if isJSONFormattedObject(string(s)) {
			log.Println("[WARN] request body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}

	i.interaction.WithJSONRequestBody(body)

	return i
}

func (i *InteractionRequest) Binary(body []byte) *InteractionRequest {
	i.interaction.WithBinaryRequestBody(body)

	return i
}

func (i *InteractionRequest) Body(contentType string, body []byte) *InteractionRequest {
	// Check if someone tried to add an object as a string representation
	// as per original allowed implementation, e.g.
	// { "foo": "bar", "baz": like("bat") }
	if isJSONFormattedObject(string(body)) {
		log.Println("[WARN] request body appears to be a JSON formatted object, " +
			"no matching will occur. Support for structured strings has been" +
			"deprecated as of 0.13.0. Please use the JSON() method instead")
	}

	i.interaction.WithRequestBody(contentType, body)

	return i
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
// Defaults to application/json.
// Use WillResponseWithContent to define custom type
func (i *InteractionRequest) WillRespondWith(status int) *InteractionResponse {
	i.interaction.WithStatus(status)

	return &InteractionResponse{
		interaction: i.interaction,
	}
}

func (i *InteractionResponse) Headers(headers MapMatcher) *InteractionResponse {
	h := make(map[string]interface{})
	for k, v := range headers {
		h[k] = v.(stringLike).string()
	}
	i.interaction.WithResponseHeaders(h)

	return i
}

// func (i *InteractionResponse) Status(status int) *InteractionResponse {
// 	i.interaction.WithStatus(status)

// 	return i
// }

func (i *InteractionResponse) JSON(body interface{}) *InteractionResponse {
	i.interaction.WithJSONResponseBody(body)

	return i
}

func (i *InteractionResponse) Binary(body []byte) *InteractionResponse {
	i.interaction.WithBinaryResponseBody(body)

	return i
}

func (i *InteractionResponse) Body(contentType string, body []byte) *InteractionResponse {
	i.interaction.WithResponseBody(contentType, body)

	return i
}

// TODO: allow these old interfaces?
//
// func (i *InteractionResponse) WillRespondWithContent(contentType string, response Response) *InteractionResponse {
// return i.WillRespondWithContent("application/json", response)
// i.Response = response
// headers := make(map[string]interface{})
// for k, v := range response.Headers {
// 	headers[k] = v.(stringLike).string()
// }
// i.interaction.WithResponseHeaders(headers)
// i.interaction.WithStatus(response.Status)

// if contentType == "application/json" {
// 	i.interaction.WithJSONResponseBody(response.Body)
// } else {
// 	bodyStr, ok := response.Body.(string)
// 	if !ok {
// 		panic("response body must be a string")
// 	}
// 	i.interaction.WithResponseBody(contentType)
// }

// 	return i
// }
