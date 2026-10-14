package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// pactJSON renders interactions as a pact file between consumer and
// provider, in the V2 format: the oldest specification pact-go supports.
//
//nolint:unparam // provider is a real parameter; the V1 scenarios happen to share one provider.
func pactJSON(consumer, provider string, interactions []model.Interaction) ([]byte, error) {
	list := make([]map[string]any, 0, len(interactions))
	for _, in := range interactions {
		request := map[string]any{"method": in.Request.Method, "path": in.Request.Path}
		if len(in.Request.Query) > 0 {
			request["query"] = in.Request.Query.Encode()
		}
		err := addPactBody(request, in.Request.Headers, in.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("interaction %d request: %w", in.Number, err)
		}

		response := map[string]any{"status": in.Response.Status}
		err = addPactBody(response, in.Response.Headers, in.Response.Body)
		if err != nil {
			return nil, fmt.Errorf("interaction %d response: %w", in.Number, err)
		}

		entry := map[string]any{"description": in.Description, "request": request, "response": response}
		if len(in.ProviderStates) > 0 {
			entry["providerState"] = in.ProviderStates[0]
		}
		list = append(list, entry)
	}

	data, err := json.MarshalIndent(map[string]any{
		"consumer":     map[string]string{"name": consumer},
		"provider":     map[string]string{"name": provider},
		"interactions": list,
		"metadata":     map[string]any{"pactSpecification": map[string]string{"version": "2.0.0"}},
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding the pact: %w", err)
	}

	return data, nil
}

// addPactBody adds headers and a body to a V2 request or response.
func addPactBody(part map[string]any, headers http.Header, body *model.Body) error {
	h := make(map[string]string, len(headers)+1)
	for name, values := range headers {
		h[name] = strings.Join(values, ", ")
	}

	if body != nil {
		if len(headerValues(headers, "Content-Type")) == 0 && body.ContentType != "" {
			h["Content-Type"] = body.ContentType
		}
		value, err := pactBody(*body)
		if err != nil {
			return err
		}
		part["body"] = value
	}

	if len(h) > 0 {
		part["headers"] = h
	}

	return nil
}

// pactBody is a body as V2 holds it: JSON documents inline, anything else
// as a string. V2 has no encoding for binary bodies, which is a limit of this
// suite's V2 pact writer, not of pact-go, so the error is not ErrUnsupported.
func pactBody(b model.Body) (any, error) {
	if isJSON(b.ContentType) && json.Valid(b.Contents) {
		return json.RawMessage(b.Contents), nil
	}

	if !utf8.Valid(b.Contents) || bytes.IndexByte(b.Contents, 0) >= 0 {
		return nil, fmt.Errorf("the suite's V2 pact writer cannot hold a binary %s body", b.ContentType)
	}

	return string(b.Contents), nil
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)

	return err == nil && (mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"))
}

// pactFile is the part of a written pact file the steps check.
type pactFile struct {
	Interactions []pactInteraction `json:"interactions"`
}

type pactInteraction struct {
	Request  pactPart `json:"request"`
	Response pactPart `json:"response"`
}

type pactPart struct {
	Method  string                     `json:"method"`
	Query   json.RawMessage            `json:"query"`
	Headers map[string]json.RawMessage `json:"headers"`
	Body    json.RawMessage            `json:"body"`
}

// header returns a header's value, matched case-insensitively. V2 pact
// files hold header values as strings, later versions as lists.
func (p pactPart) header(name string) (string, bool) {
	for key, raw := range p.Headers {
		if !strings.EqualFold(key, name) {
			continue
		}

		var text string
		if json.Unmarshal(raw, &text) == nil {
			return text, true
		}

		var list []string
		if json.Unmarshal(raw, &list) == nil {
			return strings.Join(list, ", "), true
		}
	}

	return "", false
}

// query returns the part's query parameters.
func (p pactPart) query() (url.Values, error) {
	query, err := model.ParseQuery(p.Query)
	if err != nil {
		return nil, fmt.Errorf("pact file query: %w", err)
	}

	return query, nil
}
