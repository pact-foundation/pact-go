package v3

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pact-foundation/pact-go/v3/internal/native/mockserver"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Reference to the native rust handle
	interaction          *mockserver.Interaction
	specificationVersion SpecificationVersion
}

type InteractionRequest struct {
	// Reference to the native rust handle
	interactionHandle *mockserver.Interaction
	interaction       *Interaction
}

type InteractionResponse struct {
	// Reference to the native rust handle
	interactionHandle *mockserver.Interaction
	interaction       *Interaction
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
		interactionHandle: i.interaction,
		interaction:       i,
	}
}

func (i *InteractionRequest) Query(query QueryMatcher) *InteractionRequest {
	q := make(map[string][]interface{})
	for k, values := range query {
		for _, v := range values {
			q[k] = append(q[k], v)
		}
	}
	i.interactionHandle.WithQuery(q)

	return i
}

func (i *InteractionRequest) HeadersArray(headers HeadersMatcher) *InteractionRequest {
	i.interactionHandle.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *InteractionRequest) Headers(headers MapMatcher) *InteractionRequest {
	i.interactionHandle.WithRequestHeaders(mapMatcherToNativeHeaders(headers))

	return i
}

func (i *InteractionRequest) JSON(body interface{}) *InteractionRequest {
	// TODO: Don't like panic, how to build a better builder here - nil return + log?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

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

	i.interactionHandle.WithJSONRequestBody(body)

	return i
}

func (i *InteractionRequest) Binary(body []byte) *InteractionRequest {
	i.interactionHandle.WithBinaryRequestBody(body)

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

	i.interactionHandle.WithRequestBody(contentType, body)

	return i
}

func (i *InteractionRequest) BodyMatch(body interface{}) *InteractionRequest {
	i.interactionHandle.WithJSONRequestBody(MatchV2(body))

	return i
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
// Defaults to application/json.
// Use WillResponseWithContent to define custom type
func (i *InteractionRequest) WillRespondWith(status int) *InteractionResponse {
	i.interactionHandle.WithStatus(status)

	return &InteractionResponse{
		interactionHandle: i.interactionHandle,
		interaction:       i.interaction,
	}
}

func (i *InteractionResponse) HeadersArray(headers HeadersMatcher) *InteractionResponse {
	i.interactionHandle.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *InteractionResponse) Headers(headers MapMatcher) *InteractionResponse {
	i.interactionHandle.WithRequestHeaders(mapMatcherToNativeHeaders(headers))

	return i
}

func headersMatcherToNativeHeaders(headers HeadersMatcher) map[string][]interface{} {
	h := make(map[string][]interface{})

	for k, v := range headers {
		h[k] = make([]interface{}, len(v))
		for i, vv := range v {
			h[k][i] = vv
		}
	}

	return h
}

func mapMatcherToNativeHeaders(headers MapMatcher) map[string][]interface{} {
	h := make(map[string][]interface{})

	for k, v := range headers {
		h[k] = []interface{}{v}
	}

	return h
}

func (i *InteractionResponse) JSON(body interface{}) *InteractionResponse {
	// TODO: Don't like panic, how to build a better builder here - nil return + log?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

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
	i.interactionHandle.WithJSONResponseBody(body)

	return i
}

func (i *InteractionResponse) Binary(body []byte) *InteractionResponse {
	i.interactionHandle.WithBinaryResponseBody(body)

	return i
}

func (i *InteractionResponse) Body(contentType string, body []byte) *InteractionResponse {
	i.interactionHandle.WithResponseBody(contentType, body)

	return i
}

func (i *InteractionResponse) BodyMatch(body interface{}) *InteractionResponse {
	i.interactionHandle.WithJSONRequestBody(MatchV2(body))

	return i
}

func validateMatchers(version SpecificationVersion, obj interface{}) error {
	str, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var maybeMatchers map[string]interface{}
	err = json.Unmarshal(str, &maybeMatchers)
	if err != nil {
		return err
	}

	invalidMatchers := hasMatcherGreaterThanSpec(version, maybeMatchers)

	if len(invalidMatchers) > 0 {
		return fmt.Errorf("the current pact file with specification version %s has attempted to use matchers from a higher spec version: %s", version, strings.Join(invalidMatchers, ", "))
	}

	return nil
}

func hasMatcherGreaterThanSpec(version SpecificationVersion, obj map[string]interface{}) []string {
	results := make([]string, 0)

	for k, v := range obj {
		if k == "pact:specification" && v.(string) > string(version) {
			results = append(results, obj["pact:matcher:type"].(string))
		}

		m, ok := v.(map[string]interface{})
		if ok {
			results = append(results, hasMatcherGreaterThanSpec(version, m)...)
		}
	}

	return results
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
