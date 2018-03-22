package dsl

import (
	"encoding/json"
)

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string        `json:"method"`
	Path    MatcherString `json:"path"`
	Query   MatcherMap    `json:"query,omitempty"`
	Headers MatcherMap    `json:"headers,omitempty"`
	Body    interface{}   `json:"body,omitempty"`
}

// MatcherMap allows a map[string]string-like object
// to also contain complex matchers
type MatcherMap map[string]MatcherString

// MarshalJSON is a custom encoder for Header type
func (h MatcherMap) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{}

	for header, value := range h {
		obj[header] = toObject([]byte(value))
	}

	return json.Marshal(obj)
}

// UnmarshalJSON is a custom decoder for Header type
func (h *MatcherMap) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil
}

// MatcherString allows the use of Matchers in string types
// It convert any matchers in interaction type to abstract JSON objects
// See https://github.com/pact-foundation/pact-go/issues/71 for background
type MatcherString string

// MarshalJSON is a custom encoder for Header type
func (m MatcherString) MarshalJSON() ([]byte, error) {
	var obj interface{}

	obj = toObject([]byte(m))

	return json.Marshal(obj)
}

// UnmarshalJSON is a custom decoder for Header type
func (m *MatcherString) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}
