package v3

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string         `json:"method"`
	Path    MatcherV2      `json:"path"`
	Query   QueryMatcher   `json:"query,omitempty"`
	Headers HeadersMatcher `json:"headers,omitempty"`
	Body    interface{}    `json:"body,omitempty"`
}
