package models

// Request is the default implementation of the Request interface.
// type Request struct {
// 	Method  Method                  `json:"method"`
// 	Path    matchers.Matcher        `json:"path"`
// 	Query   matchers.QueryMatcher   `json:"query,omitempty"`
// 	Headers matchers.HeadersMatcher `json:"headers,omitempty"`
// 	Body    interface{}             `json:"body,omitempty"`
// }

type Method string
