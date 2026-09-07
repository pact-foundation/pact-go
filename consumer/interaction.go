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
	i.interaction.WithRequest(request.Method, request.Path)

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

// WithCompleteResponse specifies the details of the HTTP response required by the consumer.
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

func validateMatchers(version models.SpecificationVersion, obj any) error {
	if obj == nil {
		return nil
	}

	str, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshalling the interaction body to check its matchers: %w", err)
	}

	var raw any
	err = json.Unmarshal(str, &raw)
	if err != nil {
		return fmt.Errorf("unmarshalling the interaction body to check its matchers: %w", err)
	}

	maybeMatchers, ok := raw.(map[string]any)
	if !ok {
		// Not a JSON object (e.g. a string, number, or array) - nothing to validate.
		return nil
	}

	invalidMatchers := hasMatcherGreaterThanSpec(version, maybeMatchers)

	if len(invalidMatchers) > 0 {
		return fmt.Errorf("the current pact file with specification version %s has attempted to use matchers from a higher spec version: %s", version, strings.Join(invalidMatchers, ", "))
	}

	return nil
}

func hasMatcherGreaterThanSpec(version models.SpecificationVersion, obj map[string]any) []string {
	results := make([]string, 0)

	for k, v := range obj {
		specVersion, ok := v.(string)
		if k == "pact:specification" && ok && specVersion > string(version) {
			matcherType, ok := obj["pact:matcher:type"].(string)
			if !ok {
				// Still surface the finding even if the matcher type is
				// missing or malformed, so it isn't silently dropped.
				matcherType = "<unknown matcher type>"
			}
			results = append(results, matcherType)
		}

		m, ok := v.(map[string]any)
		if ok {
			results = append(results, hasMatcherGreaterThanSpec(version, m)...)
		}
	}

	return results
}

func keyValuesToMapStringArrayInterface(key string, values ...matchers.Matcher) map[string][]any {
	q := make(map[string][]any)
	for _, v := range values {
		q[key] = append(q[key], v)
	}

	return q
}

func headersMatcherToNativeHeaders(headers matchers.HeadersMatcher) map[string][]any {
	h := make(map[string][]any)

	for k, v := range headers {
		h[k] = make([]any, len(v))
		for i, vv := range v {
			h[k][i] = vv
		}
	}

	return h
}

func headersMapMatcherToNativeHeaders(headers matchers.MapMatcher) map[string][]any {
	h := make(map[string][]any)

	for k, v := range headers {
		h[k] = []any{
			v,
		}
	}

	return h
}
