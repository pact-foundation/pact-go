// Package runner executes the compatibility features with godog and reports
// how each scenario ended.
package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/baseline"
)

// errSkipped records a scenario that a step skipped without failing.
var errSkipped = errors.New("scenario was skipped, so nothing was checked")

// godog.TestSuite.Run returns 2 when its options are invalid.
const invalidOptionsStatus = 2

// Options configures a run.
type Options struct {
	// FeaturesDir holds the .feature files, searched recursively.
	FeaturesDir string
	// Initializer registers the step definitions for each scenario.
	Initializer func(*godog.ScenarioContext)
	// JUnitPath, when set, receives a JUnit XML report.
	JUnitPath string
	// Output receives godog's progress output.
	Output io.Writer
}

// Run executes every scenario under opts.FeaturesDir. Failing scenarios are
// results, not errors: an error means the run itself could not be trusted.
func Run(opts Options) (map[baseline.Key]baseline.Result, error) {
	var (
		mu      sync.Mutex
		results = map[baseline.Key]baseline.Result{}
		keyErr  error
	)

	format := "progress"
	if opts.JUnitPath != "" {
		format += ",junit:" + opts.JUnitPath
	}

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			opts.Initializer(sc)

			// godog reports a scenario whose step returned godog.ErrSkip as
			// passing. A skipped scenario checked nothing, so it must not count
			// as a pass; steps that cannot express a scenario use
			// adapter.ErrUnsupported instead.
			skipped := false
			sc.StepContext().After(
				func(ctx context.Context, _ *godog.Step, status godog.StepResultStatus, _ error) (context.Context, error) {
					if status == godog.StepSkipped {
						skipped = true
					}
					return ctx, nil
				})

			sc.After(func(ctx context.Context, s *godog.Scenario, err error) (context.Context, error) {
				if err == nil && skipped {
					err = errSkipped
				}
				key, kerr := scenarioKey(opts.FeaturesDir, s)
				mu.Lock()
				defer mu.Unlock()
				switch _, dup := results[key]; {
				case kerr != nil:
					keyErr = errors.Join(keyErr, kerr)
				case dup:
					keyErr = errors.Join(keyErr, fmt.Errorf("two scenarios share the key %s", key))
				default:
					results[key] = classify(err)
				}
				// godog keeps the step error as the scenario's result; returning
				// it here would be reported a second time as a hook failure.
				return ctx, nil
			})
		},
		Options: &godog.Options{
			Format: format,
			Paths:  []string{opts.FeaturesDir},
			Strict: true,
			Output: opts.Output,
		},
	}

	if status := suite.Run(); status == invalidOptionsStatus {
		return nil, fmt.Errorf("godog rejected the options for %s", opts.FeaturesDir)
	}
	if keyErr != nil {
		return nil, keyErr
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no scenarios found under %s", opts.FeaturesDir)
	}
	return results, nil
}

func scenarioKey(featuresDir string, s *godog.Scenario) (baseline.Key, error) {
	rel, err := filepath.Rel(filepath.Clean(featuresDir), filepath.Clean(s.Uri))
	if err != nil || strings.HasPrefix(rel, "..") {
		return baseline.Key{}, fmt.Errorf("scenario %q is outside %s: %s", s.Name, featuresDir, s.Uri)
	}
	return baseline.Key{Feature: filepath.ToSlash(rel), Scenario: s.Name}, nil
}

func classify(err error) baseline.Result {
	switch {
	case err == nil:
		return baseline.Result{Outcome: baseline.Passed}
	case errors.Is(err, adapter.ErrUnsupported):
		return baseline.Result{Outcome: baseline.Unsupported, Message: err.Error()}
	default:
		return baseline.Result{Outcome: baseline.Failed, Message: err.Error()}
	}
}
