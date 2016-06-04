package dsl

import "net/http"

// Response is the default implementation of the Response interface.
type Response struct {
	Status  int         `json:"status"`
	Headers http.Header `json:"headers,omitempty"`
	Body    interface{} `json:"body,omitempty"`
}
