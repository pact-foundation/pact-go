// Package adapter is the seam between the compatibility steps and the Pact Go
// API under test. Step definitions only reach Pact Go through it.
package adapter

import (
	"errors"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// ErrUnsupported is wrapped by a step when the API under test cannot express
// the scenario at all, as opposed to expressing it and getting it wrong. Such
// scenarios are recorded as "unsupported" in the baseline.
var ErrUnsupported = errors.New("not supported by the API under test")

// Adapter is the API under test, as the steps drive it.
type Adapter struct {
	Consumer HTTPConsumer
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
