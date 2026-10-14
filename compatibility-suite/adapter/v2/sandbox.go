package v2

import (
	"flag"
	"sync"
	"testing"
)

// neutralFlags are the go test flags that testing.RunTests reads and that
// would otherwise let the caller's command line change a sandboxed run.
var neutralFlags = map[string]string{
	"test.run":      "",
	"test.skip":     "",
	"test.failfast": "false",
	"test.count":    "1",
}

// sandbox runs fn with a *testing.T of its own. pact-go's consumer and
// provider entry points require one, and a failure they report on it must
// not fail the suite's own test: the adapter reads the result pact-go
// returns instead.
//
// testing.RunTests honours the caller's -run, -skip, -failfast, -count and
// -cpu. sandbox neutralises the first four for the duration of the call and
// guards fn so it runs at most once, which also covers -cpu (that list is
// parsed before any test runs, so it cannot be overridden here).
//
// sandbox reports whether fn failed its *testing.T.
//
// sandbox changes process-wide flags, so it is not safe for concurrent use.
func sandbox(name string, fn func(t *testing.T)) bool {
	defer overrideFlags(neutralFlags)()

	var (
		once   sync.Once
		failed bool
	)

	testing.RunTests(func(_, _ string) (bool, error) { return true, nil },
		[]testing.InternalTest{{Name: name, F: func(t *testing.T) {
			t.Helper()

			once.Do(func() {
				defer func() { failed = t.Failed() }()

				fn(t)
			})
		}}})

	return failed
}

// overrideFlags sets flags in flag.CommandLine and returns a function that
// restores their previous values. Flags that are not registered are ignored.
func overrideFlags(values map[string]string) func() {
	saved := map[string]string{}

	for name, value := range values {
		f := flag.Lookup(name)
		if f == nil {
			continue
		}

		saved[name] = f.Value.String()

		_ = f.Value.Set(value)
	}

	return func() {
		for name, value := range saved {
			_ = flag.Lookup(name).Value.Set(value)
		}
	}
}
