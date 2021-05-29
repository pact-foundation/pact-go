package v3

// Request is the default implementation of the Request interface.
type Request struct {
	Method  Method         `json:"method"`
	Path    Matcher        `json:"path"`
	Query   QueryMatcher   `json:"query,omitempty"`
	Headers HeadersMatcher `json:"headers,omitempty"`
	Body    interface{}    `json:"body,omitempty"`
}

type Method string
