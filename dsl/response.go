package dsl

import "net/http"

// Response contains an API a given Consumer expects from a given Provider.
type Response interface {
	// Status specifies the expected HTTP status code for the API.
	Status() int

	// Status specifies the required headers (not necessarily all of them) in
	// a given API request
	Headers() http.Header

	// Content specifies the minmum data required in the API response.
	Content() interface{}
}

// PactResponse is the default implementation of the Response interface.
type PactResponse struct {
	status  int
	headers http.Header
	content interface{}
}

// Status specifies the expected HTTP status code for the API.
func (p *PactResponse) Status() int {
	return 200
}

// Headers specifies the required headers (not necessarily all of them) in
// a given API request
func (p *PactResponse) Headers() http.Header {
	return nil
}

// Content specifies the minmum data required in the API response.
func (p *PactResponse) Content() interface{} {
	return nil
}
