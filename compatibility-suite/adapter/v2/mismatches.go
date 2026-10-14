package v2

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/pact-foundation/pact-go/v2/internal/native"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// readMismatches reads the mock server's mismatches through internal/native:
// pact-go returns them only as text, and its own parse drops the body path.
func readMismatches(port int) ([]model.Mismatch, error) {
	raw := native.MockServerMismatchesJSON(port)
	if raw == "" {
		return nil, nil
	}

	return parseMismatches([]byte(raw))
}

type ffiMismatch struct {
	Type    string `json:"type"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Request struct {
		Query json.RawMessage `json:"query"`
	} `json:"request"`
	Mismatches []struct {
		Type      string `json:"type"`
		Parameter string `json:"parameter"`
		Key       string `json:"key"`
		Path      string `json:"path"`
		Mismatch  string `json:"mismatch"`
	} `json:"mismatches"`
}

// parseMismatches converts pact_ffi's mismatch report to the suite's model.
func parseMismatches(data []byte) ([]model.Mismatch, error) {
	var raw []ffiMismatch

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("parsing mock server mismatches %s: %w", data, err)
	}

	out := make([]model.Mismatch, 0, len(raw))
	for _, r := range raw {
		query, err := model.ParseQuery(r.Request.Query)
		if err != nil {
			return nil, fmt.Errorf("mismatch for %s %s: %w", r.Method, r.Path, err)
		}

		m := model.Mismatch{Type: r.Type, Method: r.Method, Path: r.Path, Query: query}
		for _, d := range r.Mismatches {
			m.Details = append(m.Details, model.MismatchDetail{
				Type:     d.Type,
				Location: cmp.Or(d.Parameter, d.Key, d.Path),
				Message:  d.Mismatch,
			})
		}
		out = append(out, m)
	}

	return out, nil
}
