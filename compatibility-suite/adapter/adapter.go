// Package adapter is the seam between the compatibility steps and the Pact Go
// API under test. Step definitions only reach Pact Go through it.
package adapter

import (
	"errors"
	"net/http"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// ErrUnsupported is wrapped by a step when the API under test cannot express
// the scenario at all, as opposed to expressing it and getting it wrong. Such
// scenarios are recorded as "unsupported" in the baseline.
var ErrUnsupported = errors.New("not supported by the API under test")

// Adapter is the API under test, as the steps drive it.
type Adapter struct {
	Consumer HTTPConsumer
	Verifier HTTPVerifier
}

// ConsumerPact names the pact a consumer test records.
type ConsumerPact struct {
	Consumer string
	Provider string
	// PactDir is where the pact file is written if the test passes.
	PactDir string
}

// HTTPConsumer runs HTTP consumer tests against a mock server.
type HTTPConsumer interface {
	// Start serves interactions from a mock server until the returned
	// MockServer is finished. It wraps ErrUnsupported when an interaction
	// cannot be expressed through the API under test.
	Start(pact ConsumerPact, interactions []model.Interaction) (MockServer, error)
}

// MockServer is a consumer test in progress.
type MockServer interface {
	// URL is the mock server's base URL.
	URL() string
	// Mismatches reports what the mock server has recorded so far.
	Mismatches() ([]model.Mismatch, error)
	// Finish ends the test as the end of a test function would, and reports
	// how it ended. Calling it again returns the same Outcome.
	Finish() Outcome
}

// Outcome is how a consumer test ended.
type Outcome struct {
	// Err is the failure the API under test reported; nil means the test
	// passed.
	Err error
	// Mismatches are the mock server's mismatches when the test ended.
	Mismatches []model.Mismatch
}

// HTTPVerifier verifies HTTP providers.
type HTTPVerifier interface {
	// Verify runs one verification. The error is for a verification that
	// could not be run or read back; a failing verification is a result.
	Verify(request VerifyRequest) (VerifyResult, error)
}

// VerifyRequest describes one provider verification.
type VerifyRequest struct {
	// ProviderURL is the base URL of the provider under test.
	ProviderURL string
	// PactFiles are local pact files to verify.
	PactFiles []string
	// Broker, when set, is a Pact Broker to fetch pacts from as well.
	Broker *BrokerSource
	// ProviderVersion identifies the provider to the broker.
	ProviderVersion string
	// Publish sends verification results back to the broker.
	Publish bool
	// StateCallback, when set, is called to set up and tear down each of
	// States.
	StateCallback func(setup bool, state string) error
	States        []string
	// Headers are added to every request made to the provider.
	Headers http.Header
}

// BrokerSource is a Pact Broker holding pacts for Provider.
type BrokerSource struct {
	URL      string
	Provider string
}

// VerifyResult is how a verification ended.
type VerifyResult struct {
	// Err is the failure the API under test reported; nil means the
	// verification passed.
	Err error
	// Errors are the failures the verifier reported, per interaction.
	Errors []model.VerificationError
}
