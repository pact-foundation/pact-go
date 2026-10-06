package runner_test

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/baseline"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/runner"
)

func sampleSteps(sc *godog.ScenarioContext) {
	sc.Step(`^it passes$`, func() error { return nil })
	sc.Step(`^it fails$`, func() error { return errors.New("boom") })
	sc.Step(`^it is unsupported$`, func() error {
		return fmt.Errorf("matching via public API: %w", adapter.ErrUnsupported)
	})
	sc.Step(`^it panics$`, func() error { panic("kaboom") })
}

func run(t *testing.T, dir string) (map[baseline.Key]baseline.Result, error) {
	t.Helper()
	results, err := runner.Run(runner.Options{FeaturesDir: dir, Initializer: sampleSteps, Output: io.Discard})
	if err != nil {
		return nil, fmt.Errorf("running %s: %w", dir, err)
	}
	return results, nil
}

func TestRunClassifiesEveryScenario(t *testing.T) {
	got, err := run(t, "testdata/features")
	if err != nil {
		t.Fatal(err)
	}
	want := map[baseline.Key]baseline.Outcome{
		{Feature: "sample.feature", Scenario: "passes"}:           baseline.Passed,
		{Feature: "sample.feature", Scenario: "fails"}:            baseline.Failed,
		{Feature: "sample.feature", Scenario: "undefined"}:        baseline.Failed,
		{Feature: "sample.feature", Scenario: "unsupported"}:      baseline.Unsupported,
		{Feature: "sample.feature", Scenario: "panics"}:           baseline.Failed,
		{Feature: "V1/nested.feature", Scenario: "nested passes"}: baseline.Passed,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d results %v, want %d", len(got), got, len(want))
	}
	for key, outcome := range want {
		if r, ok := got[key]; !ok || r.Outcome != outcome {
			t.Errorf("%v: got %+v (present=%v), want outcome %d", key, r, ok, outcome)
		}
	}
}

func TestRunRecordsFailureMessages(t *testing.T) {
	got, err := run(t, "testdata/features")
	if err != nil {
		t.Fatal(err)
	}
	for scenario, want := range map[string]string{
		"fails":       "boom",
		"undefined":   "undefined",
		"unsupported": "matching via public API",
		"panics":      "kaboom",
	} {
		msg := got[baseline.Key{Feature: "sample.feature", Scenario: scenario}].Message
		if !strings.Contains(msg, want) {
			t.Errorf("%s: message %q should contain %q", scenario, msg, want)
		}
	}
}

func TestRunRejectsDuplicateKeys(t *testing.T) {
	_, err := run(t, "testdata/dup")
	if err == nil || !strings.Contains(err.Error(), "same name") {
		t.Fatalf("want duplicate-key error naming the scenario, got %v", err)
	}
}

func TestRunRejectsEmptyFeatureDir(t *testing.T) {
	_, err := run(t, "testdata/empty")
	if err == nil || !strings.Contains(err.Error(), "no scenarios") {
		t.Fatalf("want no-scenarios error, got %v", err)
	}
}

func TestRunRejectsMissingFeatureDir(t *testing.T) {
	_, err := run(t, "testdata/does-not-exist")
	if err == nil {
		t.Fatal("want error for a missing features directory")
	}
}

func TestRunWritesJUnit(t *testing.T) {
	report := filepath.Join(t.TempDir(), "junit.xml")
	_, err := runner.Run(runner.Options{
		FeaturesDir: "testdata/features", Initializer: sampleSteps, JUnitPath: report, Output: io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // G304: report is a path this test chose under t.TempDir().
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<testsuites") {
		t.Fatalf("not a JUnit report:\n%s", data)
	}
}
