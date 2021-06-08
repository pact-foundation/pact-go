package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pact-foundation/pact-go/v2/internal/native/mockserver"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Reference to the native rust handle
	interaction          *mockserver.Interaction
	specificationVersion models.SpecificationVersion
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

type AllInteractionRequest struct {
	// Reference to the native rust handle
	interactionHandle *mockserver.Interaction
	interaction       *Interaction
}

type AllInteractionResponse struct {
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

// WithCompleteRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithCompleteRequest(request Request) *AllInteractionResponse {
	i.interaction.WithRequest(string(request.Method), request.Path)

	if request.Body != nil {
		i.interaction.WithJSONRequestBody(request.Body)
	}

	if request.Headers != nil {
		i.interaction.WithRequestHeaders(headersMapMatcherToNativeHeaders(request.Headers))
	}

	if request.Query != nil {
		i.interaction.WithQuery(headersMapMatcherToNativeHeaders(request.Query))
	}

	return &AllInteractionResponse{
		interactionHandle: i.interaction,
		interaction:       i,
	}
}

// WithCompleteResponse specifies the details of the HTTP response required by the consumer
func (i *AllInteractionResponse) WithCompleteResponse(response Response) *Interaction {
	if response.Body != nil {
		i.interactionHandle.WithJSONResponseBody(response.Body)
	}

	if response.Headers != nil {
		i.interactionHandle.WithResponseHeaders(headersMapMatcherToNativeHeaders(response.Headers))
	}

	i.interactionHandle.WithStatus(response.Status)

	return i.interaction
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithRequest(method Method, path matchers.Matcher) *InteractionRequest {
	i.interaction.WithRequest(string(method), path)

	return &InteractionRequest{
		interactionHandle: i.interaction,
		interaction:       i,
	}
}

func (i *InteractionRequest) WithQuery(key string, values ...matchers.Matcher) *InteractionRequest {
	i.interactionHandle.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *InteractionRequest) WithHeader(key string, values ...matchers.Matcher) *InteractionRequest {
	i.interactionHandle.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *InteractionRequest) WithHeaders(headers matchers.HeadersMatcher) *InteractionRequest {
	i.interactionHandle.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *InteractionRequest) WithJSONBody(body interface{}) *InteractionRequest {
	// TODO: Don't like panic, but not sure if there is a better builder experience?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] request body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}

	i.interactionHandle.WithJSONRequestBody(body)

	return i
}

func (i *InteractionRequest) WithBinaryBody(body []byte) *InteractionRequest {
	i.interactionHandle.WithBinaryRequestBody(body)

	return i
}

func (i *InteractionRequest) WithMultipartBody(contentType string, filename string, mimePartName string) *InteractionRequest {
	i.interactionHandle.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

func (i *InteractionRequest) WithBody(contentType string, body []byte) *InteractionRequest {
	// Check if someone tried to add an object as a string representation
	// as per original allowed implementation, e.g.
	// { "foo": "bar", "baz": like("bat") }
	if utils.IsJSONFormattedObject(string(body)) {
		log.Println("[WARN] request body appears to be a JSON formatted object, " +
			"no matching will occur. Support for structured strings has been" +
			"deprecated as of 0.13.0. Please use the JSON() method instead")
	}

	i.interactionHandle.WithRequestBody(contentType, body)

	return i
}

func (i *InteractionRequest) WithBodyMatch(body interface{}) *InteractionRequest {
	i.interactionHandle.WithJSONRequestBody(matchers.MatchV2(body))

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

func (i *InteractionResponse) WithHeader(key string, values ...matchers.Matcher) *InteractionResponse {
	i.interactionHandle.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

func (i *InteractionResponse) WithHeaders(headers matchers.HeadersMatcher) *InteractionResponse {
	i.interactionHandle.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

func (i *InteractionResponse) WithJSONBody(body interface{}) *InteractionResponse {
	// TODO: Don't like panic, how to build a better builder here - nil return + log?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] response body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}
	i.interactionHandle.WithJSONResponseBody(body)

	return i
}

func (i *InteractionResponse) WithBinaryBody(body []byte) *InteractionResponse {
	i.interactionHandle.WithBinaryResponseBody(body)

	return i
}

func (i *InteractionResponse) WithMultipartBody(contentType string, filename string, mimePartName string) *InteractionResponse {
	i.interactionHandle.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

func (i *InteractionResponse) WithBody(contentType string, body []byte) *InteractionResponse {
	i.interactionHandle.WithResponseBody(contentType, body)

	return i
}

func (i *InteractionResponse) WithBodyMatch(body interface{}) *InteractionResponse {
	i.interactionHandle.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

func validateMatchers(version models.SpecificationVersion, obj interface{}) error {
	if obj == nil {
		return nil
	}

	str, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var maybeMatchers map[string]interface{}
	err = json.Unmarshal(str, &maybeMatchers)
	if err != nil {
		// This means the object is not really an object, it's probably a primitive
		return nil
	}

	invalidMatchers := hasMatcherGreaterThanSpec(version, maybeMatchers)

	if len(invalidMatchers) > 0 {
		return fmt.Errorf("the current pact file with specification version %s has attempted to use matchers from a higher spec version: %s", version, strings.Join(invalidMatchers, ", "))
	}

	return nil
}

func hasMatcherGreaterThanSpec(version models.SpecificationVersion, obj map[string]interface{}) []string {
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

func keyValuesToMapStringArrayInterface(key string, values ...matchers.Matcher) map[string][]interface{} {
	q := make(map[string][]interface{})
	for _, v := range values {
		q[key] = append(q[key], v)
	}

	return q
}

func headersMatcherToNativeHeaders(headers matchers.HeadersMatcher) map[string][]interface{} {
	h := make(map[string][]interface{})

	for k, v := range headers {
		h[k] = make([]interface{}, len(v))
		for i, vv := range v {
			h[k][i] = vv
		}
	}

	return h
}

func headersMapMatcherToNativeHeaders(headers matchers.MapMatcher) map[string][]interface{} {
	h := make(map[string][]interface{})

	for k, v := range headers {
		h[k] = []interface{}{
			v,
		}
	}

	return h
}
