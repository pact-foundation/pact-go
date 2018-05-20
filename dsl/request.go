package dsl

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string        `json:"method"`
	Path    StringMatcher `json:"path"`
	Query   MapMatcher    `json:"query,omitempty"`
	Headers MapMatcher    `json:"headers,omitempty"`
	Body    interface{}   `json:"body,omitempty"`
}
