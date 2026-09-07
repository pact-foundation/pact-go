package consumer

import "github.com/pact-foundation/pact-go/v2/matchers"

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string              `json:"method"`
	Path    matchers.Matcher    `json:"path"`
	Query   matchers.MapMatcher `json:"query,omitempty"`
	Headers matchers.MapMatcher `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
}

// Method is an HTTP request method (e.g. "GET", "POST") used when
// building the expected request of an interaction.
type Method string
