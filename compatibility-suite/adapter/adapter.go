// Package adapter is the seam between the compatibility steps and the Pact Go
// API under test. Step definitions only reach Pact Go through it.
package adapter

import "errors"

// ErrUnsupported is wrapped by a step when the API under test cannot express
// the scenario at all, as opposed to expressing it and getting it wrong. Such
// scenarios are recorded as "unsupported" in the baseline.
var ErrUnsupported = errors.New("not supported by the API under test")
