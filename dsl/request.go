package dsl

import (
	"net/http"
)

// Request contains the expectations from an API call from a given Consumer.
type Request interface {
	// Method specifies the HTTP verb in this API Request.
	Method() string

	// Path specifies request path in this request.
	Path() string

	// Query specifies query string (if any) in this request.
	Query() string

	// Headers specifies the required headers (not necessarily all of them) in
	// a given API request.
	Headers() http.Header

	// Content specifies the minmum data required in the API response.
	Content() interface{}
}

// PactRequest is the default implementation of the Request interface.
type PactRequest struct {
	method  string
	path    string
	query   string
	headers http.Header
	content interface{}
}

// Method specifies the HTTP verb in this API Request.
func (p *PactRequest) Method() string {
	return ""
}

// Path specifies request path in this request.
func (p *PactRequest) Path() string {
	return "/"
}

// Query specifies query string (if any) in this request.
func (p *PactRequest) Query() string {
	return ""
}

// Headers specifies the required headers (not necessarily all of them) in
// a given API request.
func (p *PactRequest) Headers() http.Header {
	return nil
}

// Content specifies the minmum data required in the API response.
func (p *PactRequest) Content() interface{} {
	return nil
}
