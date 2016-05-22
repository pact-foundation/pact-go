package dsl

import "fmt"

// PactMock is a Mock of the Pact interface.
type PactMock struct {
	PactGiven      string
	PactReceived   string
	PactRequest    *Request
	PactResponse   *Response
	VerifyCalled   bool
	VerifyResponse error
}

// Given specifies a provider state. Optional.
func (p *PactMock) Given(state string) *PactMock {
	fmt.Println("Given()")
	p.PactGiven = state
	return p
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (p *PactMock) UponReceiving(test string) *PactMock {
	fmt.Println("UponReceiving()")
	return p
}

// WithRequest specifies the details of the HTTP request that will be used to
// confirm that the Provider provides an API listening on the given interface.
// Mandatory.
func (p *PactMock) WithRequest(request Request) *PactMock {
	fmt.Println("WithRequest()")
	return p
}

// WillRespondWith specifies the details of the HTTP response that will be used to
// confirm that the Provider must satisfy. Mandatory.
func (p *PactMock) WillRespondWith(response Response) *PactMock {
	fmt.Println("RespondWith()")
	return p
}

// Verify runs the current test case against a Mock Service.
func (p *PactMock) Verify() error {
	fmt.Println("Verify()")
	return p.VerifyResponse
}
