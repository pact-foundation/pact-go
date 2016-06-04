package dsl

import (
	"net/http"
)

// Request is the default implementation of the Request interface.
type Request struct {
	Method  string      `json:"method"`
	Path    string      `json:"path,omitempty"`
	Query   string      `json:"query,omitempty"`
	Headers http.Header `json:"headers,omitempty"`
	Body    interface{} `json:"body,omitempty"`
}
