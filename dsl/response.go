package dsl

// Response is the default implementation of the Response interface.
type Response struct {
	Status  int         `json:"status"`
	Headers MatcherMap  `json:"headers,omitempty"`
	Body    interface{} `json:"body,omitempty"`
}
