package consumer

import "github.com/pact-foundation/pact-go/v2/matchers"

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string              `json:"method"`
	Path    matchers.Matcher    `json:"path"`
	Query   matchers.MapMatcher `json:"query,omitempty"`
	Headers matchers.MapMatcher `json:"headers,omitempty"`
	Body    interface{}         `json:"body,omitempty"`
}
type Method string
