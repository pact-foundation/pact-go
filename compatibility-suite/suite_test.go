package compatibility_test

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"testing"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/baseline"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/runner"
)

const (
	featuresDir  = "pact-compatibility-suite/features"
	baselinePath = "baseline.yaml"
)

var updateBaseline = flag.Bool("update-baseline", false,
	"rewrite baseline.yaml from this run's results instead of checking against it")

// noSteps registers no step definitions yet: every scenario is undefined, and
// so failing, until the step packages land.
func noSteps(*godog.ScenarioContext) {}

func TestCompatibility(t *testing.T) {
	results, err := runner.Run(runner.Options{
		FeaturesDir: featuresDir,
		Initializer: noSteps,
		JUnitPath:   os.Getenv("COMPAT_JUNIT"),
		Output:      os.Stdout,
	})
	if err != nil {
		t.Fatal(err)
	}

	current, err := baseline.Load(baselinePath)
	if *updateBaseline {
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			t.Fatal(err)
		}
		updated := baseline.Update(current, results)
		err = updated.Write(baselinePath)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s: %d of %d scenarios listed", baselinePath, len(updated), len(results))
		return
	}
	if err != nil {
		t.Fatal(err)
	}

	violations := baseline.Compare(current, results)
	for _, v := range violations {
		t.Error(v)
	}
	if len(violations) > 0 {
		t.Logf("%d violation(s). If intended, run: mise run compat --update-baseline", len(violations))
	}
}
