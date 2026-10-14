package compatibility_test

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	adapterv2 "github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter/v2"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/baseline"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/runner"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/steps"
)

const (
	featuresDir  = "pact-compatibility-suite/features"
	fixturesDir  = "pact-compatibility-suite/fixtures"
	baselinePath = "baseline.yaml"
)

var (
	updateBaseline = flag.Bool("update-baseline", false,
		"rewrite baseline.yaml from this run's results instead of checking against it")
	adapterName = flag.String("adapter", "v2", "the Pact Go adapter to run the suite against")
)

func adapterFor(name string) (adapter.Adapter, error) {
	switch name {
	case "v2":
		return adapterv2.New(), nil
	default:
		return adapter.Adapter{}, fmt.Errorf("unknown adapter %q (known: v2)", name)
	}
}

func TestCompatibility(t *testing.T) {
	api, err := adapterFor(*adapterName)
	if err != nil {
		t.Fatal(err)
	}

	results, err := runner.Run(runner.Options{
		FeaturesDir: featuresDir,
		Initializer: steps.Initialize(api, fixturesDir),
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
