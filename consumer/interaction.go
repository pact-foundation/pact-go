package consumer

import (
	"encoding/json"
	"fmt"
	"strings"

	mockserver "github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
)

// Interaction is the main implementation of the Pact interface.
type Interaction struct {
	// Reference to the native rust handle
	interaction          *mockserver.Interaction
	specificationVersion models.SpecificationVersion
}

// WithCompleteRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (i *Interaction) WithCompleteRequest(request Request) *Interaction {
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

	return i
}

// WithCompleteResponse specifies the details of the HTTP response required by the consumer
func (i *Interaction) WithCompleteResponse(response Response) *Interaction {
	if response.Body != nil {
		i.interaction.WithJSONResponseBody(response.Body)
	}

	if response.Headers != nil {
		i.interaction.WithResponseHeaders(headersMapMatcherToNativeHeaders(response.Headers))
	}

	i.interaction.WithStatus(response.Status)

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
