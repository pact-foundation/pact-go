// Package baseline records which compatibility scenarios are not expected to
// pass, and checks a run's results against that record so the suite can only
// ever get stricter.
package baseline

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Status is why a scenario is listed in the baseline.
type Status string

const (
	// StatusFailing marks a scenario Pact Go currently fails.
	StatusFailing Status = "failing"
	// StatusUnsupported marks a scenario the adapter cannot express through
	// the API under test.
	StatusUnsupported Status = "unsupported"
)

// DefaultReason is recorded by Update for a scenario that newly fails.
const DefaultReason = "not yet triaged"

const header = `# Compatibility scenarios that are not expected to pass.
# Edit reasons freely; regenerate entries with: mise run compat --update-baseline
# See docs/contributing/compatibility-suite.md.
`

// Key identifies a scenario by its feature path, relative to the upstream
// features directory and slash-separated, and its name.
type Key struct {
	Feature  string
	Scenario string
}

func (k Key) String() string { return k.Feature + " :: " + k.Scenario }

// Compare orders keys by feature, then scenario.
func (k Key) Compare(other Key) int {
	if c := cmp.Compare(k.Feature, other.Feature); c != 0 {
		return c
	}
	return cmp.Compare(k.Scenario, other.Scenario)
}

// Entry is one listed scenario.
type Entry struct {
	Status Status
	Reason string
}

// Baseline maps each listed scenario to its expected status.
type Baseline map[Key]Entry

// Outcome is how a scenario ended in a run.
type Outcome int

const (
	// Passed means every step succeeded.
	Passed Outcome = iota
	// Failed means a step failed, panicked or was undefined.
	Failed
	// Unsupported means a step reported adapter.ErrUnsupported.
	Unsupported
)

// Result is how one scenario ended, with the error message when it did not
// pass.
type Result struct {
	Outcome Outcome
	Message string
}

// record is the on-disk form of one entry: a list keeps long scenario names
// readable, where YAML would turn them into complex mapping keys.
type record struct {
	Feature  string `yaml:"feature"`
	Scenario string `yaml:"scenario"`
	Status   Status `yaml:"status"`
	Reason   string `yaml:"reason"`
}

// Load reads a baseline file. A missing file wraps fs.ErrNotExist.
func Load(path string) (Baseline, error) {
	//nolint:gosec // G304: path is the suite's own baseline file, chosen by the test, not by input.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline: %w", err)
	}

	var records []record
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	err = dec.Decode(&records)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parsing baseline %s: %w", path, err)
	}

	b := make(Baseline, len(records))
	for _, r := range records {
		key := Key{Feature: r.Feature, Scenario: r.Scenario}
		switch {
		case r.Feature == "" || r.Scenario == "":
			return nil, fmt.Errorf("baseline %s: entry missing feature or scenario: %+v", path, r)
		case r.Status != StatusFailing && r.Status != StatusUnsupported:
			return nil, fmt.Errorf("baseline %s: %s has unknown status %q", path, key, r.Status)
		case strings.TrimSpace(r.Reason) == "":
			return nil, fmt.Errorf("baseline %s: %s has no reason", path, key)
		}
		if _, dup := b[key]; dup {
			return nil, fmt.Errorf("baseline %s: %s is listed twice", path, key)
		}
		b[key] = Entry{Status: r.Status, Reason: r.Reason}
	}
	return b, nil
}

// Write saves the baseline sorted by key, so rewriting it is stable.
func (b Baseline) Write(path string) error {
	records := make([]record, 0, len(b))
	for key, e := range b {
		records = append(records, record{
			Feature: key.Feature, Scenario: key.Scenario, Status: e.Status, Reason: e.Reason,
		})
	}
	slices.SortFunc(records, func(x, y record) int {
		return Key{x.Feature, x.Scenario}.Compare(Key{y.Feature, y.Scenario})
	})

	body, err := yaml.Marshal(records)
	if err != nil {
		return fmt.Errorf("encoding baseline: %w", err)
	}
	err = os.WriteFile(path, append([]byte(header), body...), 0o600)
	if err != nil {
		return fmt.Errorf("writing baseline: %w", err)
	}
	return nil
}

// Kind classifies a Violation.
type Kind string

const (
	// Regression means a scenario not in the baseline did not pass.
	Regression Kind = "regression"
	// Fixed means a listed scenario now passes and should be removed.
	Fixed Kind = "fixed"
	// Stale means a listed scenario no longer exists in the suite.
	Stale Kind = "stale"
	// Misclassified means a listed scenario ended differently from its status.
	Misclassified Kind = "misclassified"
)

// Violation is one disagreement between a run and the baseline.
type Violation struct {
	Key    Key
	Kind   Kind
	Detail string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: %s: %s", v.Kind, v.Key, v.Detail)
}

// Compare reports every way results disagree with the baseline, sorted by
// key. An empty result means the run matches.
func Compare(b Baseline, results map[Key]Result) []Violation {
	var out []Violation
	for key, r := range results {
		entry, listed := b[key]
		switch {
		case r.Outcome == Passed && listed:
			out = append(out, Violation{key, Fixed, "now passes; remove it from the baseline"})
		case r.Outcome == Passed:
		case !listed:
			out = append(out, Violation{key, Regression, r.Message})
		case r.Outcome == Failed && entry.Status == StatusUnsupported:
			out = append(out, Violation{key, Misclassified, "listed as unsupported but failed: " + r.Message})
		case r.Outcome == Unsupported && entry.Status == StatusFailing:
			out = append(out, Violation{key, Misclassified, "listed as failing but is unsupported: " + r.Message})
		}
	}
	for key := range b {
		if _, ran := results[key]; !ran {
			out = append(out, Violation{key, Stale, "not in the suite; remove it from the baseline"})
		}
	}
	slices.SortFunc(out, func(x, y Violation) int { return x.Key.Compare(y.Key) })
	return out
}

// Update returns the baseline that results satisfy, keeping the reason of any
// entry whose status is unchanged.
func Update(b Baseline, results map[Key]Result) Baseline {
	next := Baseline{}
	for key, r := range results {
		prev, listed := b[key]
		switch r.Outcome {
		case Passed:
		case Failed:
			if listed && prev.Status == StatusFailing {
				next[key] = prev
			} else {
				next[key] = Entry{Status: StatusFailing, Reason: DefaultReason}
			}
		case Unsupported:
			if listed && prev.Status == StatusUnsupported {
				next[key] = prev
			} else {
				next[key] = Entry{Status: StatusUnsupported, Reason: r.Message}
			}
		}
	}
	return next
}
