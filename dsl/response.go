package dsl

import "net/http"

// Response contains the expectations from an API call from a given Consumer.
type Response interface {
	Method() string
	Path() string
	Query() string
	Headers() http.Header
	Content() interface{}
}

// PactResponse is the default implementation of the Response interface.
type PactResponse struct {
	Method  string
	Path    string
	Query   string
	Headers http.Header
	content interface{}
}
