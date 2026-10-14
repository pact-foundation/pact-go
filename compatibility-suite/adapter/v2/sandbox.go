package v2

import "testing"

// sandbox runs fn with a *testing.T of its own. pact-go's consumer and
// provider entry points require one, and a failure they report on it must
// not fail the suite's own test: the adapter reads the result pact-go
// returns instead.
func sandbox(name string, fn func(t *testing.T)) {
	testing.RunTests(func(_, _ string) (bool, error) { return true, nil },
		[]testing.InternalTest{{Name: name, F: fn}})
}
