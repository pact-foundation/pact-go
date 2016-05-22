package dsl

import (
	"net/http"
)

// Request contains an API a given Consumer expects from a given Provider.
type Request interface {
	Status() int
	Headers() http.Header
	Content() interface{}
}

// PactRequest is the default implementation of the Request interface.
type PactRequest interface {
}
