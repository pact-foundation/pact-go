// Package model describes HTTP interactions and their results independently
// of any Pact implementation, so the steps and the adapters share one
// vocabulary.
package model

import (
	"bytes"
	"net/http"
	"net/url"
	"slices"
)

// Body is the contents of a request or response, with its content type.
type Body struct {
	ContentType string
	Contents    []byte
}

// Request is the HTTP request an interaction expects.
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers http.Header
	Body    *Body
}

// Response is the HTTP response an interaction returns.
type Response struct {
	Status  int
	Headers http.Header
	Body    *Body
}

// Interaction is one HTTP interaction from a feature's table.
type Interaction struct {
	// Number is the interaction's 1-based row in its table, as steps refer
	// to it.
	Number         int
	Description    string
	ProviderStates []string
	Request        Request
	Response       Response
}

// Clone returns a deep copy, so a step can change an interaction without
// changing the table it came from.
func (i Interaction) Clone() Interaction {
	c := i
	c.ProviderStates = slices.Clone(i.ProviderStates)
	c.Request.Query = url.Values(http.Header(i.Request.Query).Clone())
	c.Request.Headers = i.Request.Headers.Clone()
	c.Request.Body = i.Request.Body.clone()
	c.Response.Headers = i.Response.Headers.Clone()
	c.Response.Body = i.Response.Body.clone()

	return c
}

func (b *Body) clone() *Body {
	if b == nil {
		return nil
	}

	return &Body{ContentType: b.ContentType, Contents: bytes.Clone(b.Contents)}
}

// Mismatch is one problem a mock server recorded.
type Mismatch struct {
	// Type is missing-request, request-not-found or request-mismatch.
	Type   string
	Method string
	Path   string
	// Query is the request's query, when the mock server reports it.
	Query   url.Values
	Details []MismatchDetail
}

// MismatchDetail is one difference within a mismatched request.
type MismatchDetail struct {
	// Type is the Pact mismatch type, such as QueryMismatch or BodyMismatch.
	Type string
	// Location is the query parameter, header name or body path concerned.
	Location string
	Message  string
}

// VerificationError is one failure a provider verification reported.
type VerificationError struct {
	Interaction string
	// MismatchTypes are the Pact mismatch types found, such as
	// StatusMismatch.
	MismatchTypes []string
	// Message is set when the interaction failed with an error rather than
	// with mismatches.
	Message string
}
