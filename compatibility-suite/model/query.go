package model

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ParseQuery reads a query the way pact files and pact_ffi write one: as a
// map of values (V3 and later) or as a query string (V2). An absent or null
// query is nil.
func ParseQuery(raw json.RawMessage) (url.Values, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var values map[string][]string
	if json.Unmarshal(raw, &values) == nil {
		return values, nil
	}

	var text string
	err := json.Unmarshal(raw, &text)
	if err != nil {
		return nil, fmt.Errorf("unrecognised query %s: %w", raw, err)
	}

	query, err := url.ParseQuery(text)
	if err != nil {
		return nil, fmt.Errorf("parsing query %q: %w", text, err)
	}

	return query, nil
}
