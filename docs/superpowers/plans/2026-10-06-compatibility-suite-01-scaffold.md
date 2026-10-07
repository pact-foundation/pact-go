# Compatibility Suite — PR 1: Scaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land a runnable, CI-gated compatibility suite harness with the upstream features vendored and every scenario recorded in a ratchet baseline, so later PRs only ever move scenarios *out* of the baseline.

**Architecture:** A nested Go module at `compatibility-suite/` (own `go.mod`, `replace` to the root) runs the upstream Gherkin with godog. A `runner` package executes the features and classifies each scenario as passed / failed / unsupported; a `baseline` package loads `baseline.yaml` and reports regressions, fixes, stale and misclassified entries. No step definitions exist yet, so every scenario is undefined → failed → listed in the baseline.

**Tech Stack:** Go 1.25+, `github.com/cucumber/godog` v0.16.0, `gopkg.in/yaml.v3`, `git subtree`, mise, GitHub Actions, `gh stack`.

**Spec:** `docs/superpowers/specs/2026-10-06-compatibility-suite-design.md`

This is the first of five stacked PRs (see the spec's "Delivery" section). PRs 2–5 get their own plans, written once this scaffold has landed.

## Global Constraints

- Upstream: `https://github.com/pact-foundation/pact-compatibility-suite.git`, pinned at `dd08122ce7343730fd373574a91d046e8814be43` (190 scenarios, 20 feature files).
- Nested module path: `github.com/pact-foundation/pact-go/v2/compatibility-suite`, with `replace github.com/pact-foundation/pact-go/v2 => ../`.
- Nothing is added to the root `go.mod`.
- Upstream files under `compatibility-suite/pact-compatibility-suite/` are never edited by hand.
- Scenario key = feature path relative to `features/` (slash-separated) + scenario name.
- Baseline statuses are exactly `failing` and `unsupported`; every entry has a non-empty reason.
- No Docker, no network at test time.
- CI job: `ubuntu-latest`, Go `1.27`, required via `complete.needs`.
- All code passes the repo's `.golangci.yml` (`default: all`, gofumpt).
- Commit messages end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. **Wrong or empty features path**: if the runner finds zero scenarios (path typo, subtree missing), the run must fail loudly, never pass vacuously against an empty result set. → Task 3, `TestRunRejectsEmptyFeatureDir` / `TestRunRejectsMissingFeatureDir`.
2. **Upstream rename of a scenario**: the old name must surface as `stale` and the new one as `regression`, both in the same run, so `-update-baseline` is an obvious fix. → Task 2, `TestCompareReportsStaleEntries`.
3. **Hand-edited baseline mistakes** (typo'd field, unknown status, blank reason, duplicate entry, malformed YAML): must produce a clear load error naming the file, not "190 regressions". → Task 2, `TestLoadRejects*`.
4. **`-update-baseline` churn**: rewriting must keep hand-written reasons, sort deterministically and be idempotent, so re-running produces no diff. → Task 2, `TestUpdateKeepsReasons`, `TestWriteIsStable`.
5. **A panicking step** must be recorded as a failed scenario, not crash the suite and lose all other results. → Task 3, `TestRunClassifiesEveryScenario` (`panics` case).

---

## File Structure

```
compatibility-suite/
  go.mod, go.sum                      # nested module                              (Task 1)
  pact-compatibility-suite/           # git subtree, upstream @ dd08122            (Task 1)
  adapter/adapter.go                  # ErrUnsupported sentinel                    (Task 3)
  baseline/baseline.go                # Key, Entry, Load, Write, Compare, Update   (Task 2)
  baseline/baseline_test.go                                                        (Task 2)
  runner/runner.go                    # Run: godog execution + classification      (Task 3)
  runner/runner_test.go                                                            (Task 3)
  runner/testdata/features/sample.feature                                          (Task 3)
  runner/testdata/features/V1/nested.feature                                       (Task 3)
  runner/testdata/dup/dup.feature                                                  (Task 3)
  runner/testdata/empty/.gitkeep                                                   (Task 3)
  suite_test.go                       # TestCompatibility + -update-baseline       (Task 4)
  baseline.yaml                       # seeded ratchet (190 entries)               (Task 4)
mise.toml                             # compat task; lint/verify cover nested mod  (Task 5)
.github/workflows/test.yml            # compatibility job, gated                   (Task 5)
docs/contributing/compatibility-suite.md                                           (Task 5)
```

---

### Task 0: Branch for the stack

- [ ] **Step 1: Rename the working branch to the stack's first branch**

The spec and this plan are committed on `feat/compatibility-suite`. They ship with PR 1.

```bash
git branch -m feat/compatibility-suite compat/01-scaffold
git status
```
Expected: `On branch compat/01-scaffold`, clean tree.

- [ ] **Step 2: Install the stacked-PR extension**

```bash
gh extension install github/gh-stack
gh stack --help
```
Expected: help text listing the stack subcommands. Note the commands for initialising a stack and submitting/pushing branches as PRs; Task 6 uses them.

---

### Task 1: Vendor the upstream suite and create the nested module

**Files:**
- Create: `compatibility-suite/pact-compatibility-suite/**` (via subtree)
- Create: `compatibility-suite/go.mod`, `compatibility-suite/go.sum`

**Interfaces:**
- Produces: features at `compatibility-suite/pact-compatibility-suite/features/V{1..4}/*.feature`; module `github.com/pact-foundation/pact-go/v2/compatibility-suite`.

- [ ] **Step 1: Import the upstream suite as a squashed subtree at the pinned SHA**

```bash
git fetch https://github.com/pact-foundation/pact-compatibility-suite.git main
git subtree add --prefix=compatibility-suite/pact-compatibility-suite \
  dd08122ce7343730fd373574a91d046e8814be43 --squash \
  -m "chore: vendor pact-compatibility-suite at dd08122

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

- [ ] **Step 2: Verify the import**

```bash
find compatibility-suite/pact-compatibility-suite/features -name '*.feature' | wc -l
cat compatibility-suite/pact-compatibility-suite/features/*/*.feature | grep -cE '^\s*Scenario'
```
Expected: `20`, then `190`.

- [ ] **Step 3: Create the nested module**

`compatibility-suite/go.mod`:
```
module github.com/pact-foundation/pact-go/v2/compatibility-suite

go 1.25.0

replace github.com/pact-foundation/pact-go/v2 => ../
```

```bash
cd compatibility-suite
go get github.com/cucumber/godog@v0.16.0 gopkg.in/yaml.v3@v3.0.1
cd ..
```
(`go mod tidy` will drop these again until Tasks 2–3 import them. That's expected; tidy runs at the end of Task 3.)

- [ ] **Step 4: Confirm the root module is unaffected**

```bash
go list ./... | grep -c compatibility-suite
git diff --stat go.mod go.sum
```
Expected: `0` packages listed, no diff to the root `go.mod`/`go.sum`.

- [ ] **Step 5: Commit**

```bash
git add compatibility-suite/go.mod compatibility-suite/go.sum
git commit -m "chore(compat): add nested module for the compatibility suite

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Baseline ratchet

**Files:**
- Create: `compatibility-suite/baseline/baseline.go`
- Test: `compatibility-suite/baseline/baseline_test.go`

**Interfaces:**
- Produces (used by Tasks 3 and 4):
  - `type Key struct{ Feature, Scenario string }`, `func (Key) String() string`, `func (Key) Compare(Key) int`
  - `type Status string`; `StatusFailing`, `StatusUnsupported`; `const DefaultReason = "not yet triaged"`
  - `type Entry struct{ Status Status; Reason string }`; `type Baseline map[Key]Entry`
  - `type Outcome int`; `Passed`, `Failed`, `Unsupported`
  - `type Result struct{ Outcome Outcome; Message string }`
  - `func Load(path string) (Baseline, error)`: a missing file wraps `fs.ErrNotExist`
  - `func (Baseline) Write(path string) error`
  - `type Kind string`; `Regression`, `Fixed`, `Stale`, `Misclassified`
  - `type Violation struct{ Key Key; Kind Kind; Detail string }`, `func (Violation) String() string`
  - `func Compare(b Baseline, results map[Key]Result) []Violation`
  - `func Update(b Baseline, results map[Key]Result) Baseline`

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/baseline/baseline_test.go`:
```go
package baseline_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/baseline"
)

var (
	keyA = baseline.Key{Feature: "V1/a.feature", Scenario: "alpha"}
	keyB = baseline.Key{Feature: "V1/a.feature", Scenario: "beta"}
	keyC = baseline.Key{Feature: "V2/c.feature", Scenario: "gamma"}
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "baseline.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadValidFile(t *testing.T) {
	path := writeFile(t, `
- feature: V1/a.feature
  scenario: alpha
  status: failing
  reason: mismatch text differs
- feature: V2/c.feature
  scenario: gamma
  status: unsupported
  reason: no public matcher API
`)
	got, err := baseline.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := baseline.Baseline{
		keyA: {Status: baseline.StatusFailing, Reason: "mismatch text differs"},
		keyC: {Status: baseline.StatusUnsupported, Reason: "no public matcher API"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestLoadEmptyFileIsEmptyBaseline(t *testing.T) {
	for name, content := range map[string]string{
		"empty":        "",
		"comment only": "# nothing here\n",
		"empty list":   "[]\n",
	} {
		t.Run(name, func(t *testing.T) {
			got, err := baseline.Load(writeFile(t, content))
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 0 {
				t.Fatalf("want empty baseline, got %v", got)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := baseline.Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want fs.ErrNotExist, got %v", err)
	}
}

func TestLoadRejectsBadEntries(t *testing.T) {
	cases := map[string]struct{ content, wantErr string }{
		"unknown status": {
			"- {feature: V1/a.feature, scenario: alpha, status: flaky, reason: r}\n",
			`unknown status "flaky"`,
		},
		"blank reason": {
			"- {feature: V1/a.feature, scenario: alpha, status: failing, reason: '  '}\n",
			"no reason",
		},
		"missing scenario": {
			"- {feature: V1/a.feature, status: failing, reason: r}\n",
			"missing feature or scenario",
		},
		"typo'd field": {
			"- {feature: V1/a.feature, scenario: alpha, staus: failing, reason: r}\n",
			"staus",
		},
		"duplicate": {
			"- {feature: V1/a.feature, scenario: alpha, status: failing, reason: r}\n" +
				"- {feature: V1/a.feature, scenario: alpha, status: failing, reason: r}\n",
			"listed twice",
		},
		"malformed": {"- [unclosed\n", "parsing baseline"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := writeFile(t, tc.content)
			_, err := baseline.Load(path)
			if err == nil {
				t.Fatal("want error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) || !strings.Contains(err.Error(), path) {
				t.Fatalf("error %q should mention %q and the file path", err, tc.wantErr)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	failing := baseline.Entry{Status: baseline.StatusFailing, Reason: "r"}
	unsupported := baseline.Entry{Status: baseline.StatusUnsupported, Reason: "r"}
	pass := baseline.Result{Outcome: baseline.Passed}
	fail := baseline.Result{Outcome: baseline.Failed, Message: "boom"}
	unsup := baseline.Result{Outcome: baseline.Unsupported, Message: "no API"}

	cases := map[string]struct {
		entry    *baseline.Entry
		result   baseline.Result
		wantKind baseline.Kind // "" means no violation
	}{
		"unlisted pass":            {nil, pass, ""},
		"unlisted fail":            {nil, fail, baseline.Regression},
		"unlisted unsupported":     {nil, unsup, baseline.Regression},
		"failing still fails":      {&failing, fail, ""},
		"failing now passes":       {&failing, pass, baseline.Fixed},
		"failing now unsupported":  {&failing, unsup, baseline.Misclassified},
		"unsupported still is":     {&unsupported, unsup, ""},
		"unsupported now passes":   {&unsupported, pass, baseline.Fixed},
		"unsupported now fails":    {&unsupported, fail, baseline.Misclassified},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			b := baseline.Baseline{}
			if tc.entry != nil {
				b[keyA] = *tc.entry
			}
			got := baseline.Compare(b, map[baseline.Key]baseline.Result{keyA: tc.result})
			switch {
			case tc.wantKind == "" && len(got) != 0:
				t.Fatalf("want no violations, got %v", got)
			case tc.wantKind != "" && (len(got) != 1 || got[0].Kind != tc.wantKind || got[0].Key != keyA):
				t.Fatalf("want one %s violation for %v, got %v", tc.wantKind, keyA, got)
			}
		})
	}
}

func TestCompareRegressionCarriesMessage(t *testing.T) {
	got := baseline.Compare(baseline.Baseline{}, map[baseline.Key]baseline.Result{
		keyA: {Outcome: baseline.Failed, Message: "expected 200 got 500"},
	})
	if len(got) != 1 || !strings.Contains(got[0].String(), "expected 200 got 500") {
		t.Fatalf("regression should carry the failure message, got %v", got)
	}
}

func TestCompareReportsStaleEntries(t *testing.T) {
	// Upstream renamed "alpha" to "beta": the old entry is stale and the new
	// scenario is an unlisted failure, both reported in one run.
	b := baseline.Baseline{keyA: {Status: baseline.StatusFailing, Reason: "r"}}
	got := baseline.Compare(b, map[baseline.Key]baseline.Result{
		keyB: {Outcome: baseline.Failed, Message: "boom"},
	})
	if len(got) != 2 {
		t.Fatalf("want 2 violations, got %v", got)
	}
	if got[0].Key != keyA || got[0].Kind != baseline.Stale {
		t.Errorf("want stale %v first, got %v", keyA, got[0])
	}
	if got[1].Key != keyB || got[1].Kind != baseline.Regression {
		t.Errorf("want regression %v second, got %v", keyB, got[1])
	}
}

func TestUpdateKeepsReasons(t *testing.T) {
	b := baseline.Baseline{
		keyA: {Status: baseline.StatusFailing, Reason: "hand-written reason"},
		keyB: {Status: baseline.StatusFailing, Reason: "fixed since"},
		keyC: {Status: baseline.StatusUnsupported, Reason: "no API"},
	}
	results := map[baseline.Key]baseline.Result{
		keyA: {Outcome: baseline.Failed, Message: "boom"},
		keyB: {Outcome: baseline.Passed},
		keyC: {Outcome: baseline.Failed, Message: "boom"},
	}
	got := baseline.Update(b, results)
	want := baseline.Baseline{
		keyA: {Status: baseline.StatusFailing, Reason: "hand-written reason"},
		keyC: {Status: baseline.StatusFailing, Reason: baseline.DefaultReason},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	if v := baseline.Compare(got, results); len(v) != 0 {
		t.Fatalf("updated baseline should satisfy Compare, got %v", v)
	}
}

func TestUpdateRecordsUnsupportedMessage(t *testing.T) {
	got := baseline.Update(baseline.Baseline{}, map[baseline.Key]baseline.Result{
		keyA: {Outcome: baseline.Unsupported, Message: "matching: unsupported"},
	})
	want := baseline.Entry{Status: baseline.StatusUnsupported, Reason: "matching: unsupported"}
	if got[keyA] != want {
		t.Fatalf("got %#v, want %#v", got[keyA], want)
	}
}

func TestUpdateDropsScenariosNotInResults(t *testing.T) {
	b := baseline.Baseline{keyA: {Status: baseline.StatusFailing, Reason: "r"}}
	if got := baseline.Update(b, map[baseline.Key]baseline.Result{}); len(got) != 0 {
		t.Fatalf("want stale entry dropped, got %v", got)
	}
}

func TestWriteIsStable(t *testing.T) {
	b := baseline.Baseline{
		keyC: {Status: baseline.StatusUnsupported, Reason: "no API"},
		keyA: {Status: baseline.StatusFailing, Reason: "r"},
		keyB: {Status: baseline.StatusFailing, Reason: "r"},
	}
	path := filepath.Join(t.TempDir(), "baseline.yaml")
	if err := b.Write(path); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)

	loaded, err := baseline.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, b) {
		t.Fatalf("round trip: got %#v, want %#v", loaded, b)
	}
	if err := loaded.Write(path); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatalf("rewrite changed the file:\n%s\n---\n%s", first, second)
	}
	if a, c := strings.Index(string(first), "alpha"), strings.Index(string(first), "gamma"); a > c {
		t.Fatalf("entries not sorted by key:\n%s", first)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd compatibility-suite && go test ./baseline/`
Expected: FAIL, build error `no non-test Go files` / undefined `baseline.Key`.

- [ ] **Step 3: Implement**

`compatibility-suite/baseline/baseline.go`:
```go
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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline: %w", err)
	}

	var records []record
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&records); err != nil && !errors.Is(err, io.EOF) {
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
	if err := os.WriteFile(path, append([]byte(header), body...), 0o600); err != nil {
		return fmt.Errorf("writing baseline: %w", err)
	}
	return nil
}

// Kind classifies a Violation.
type Kind string

const (
	// Regression: a scenario not in the baseline did not pass.
	Regression Kind = "regression"
	// Fixed: a listed scenario now passes and should be removed.
	Fixed Kind = "fixed"
	// Stale: a listed scenario no longer exists in the suite.
	Stale Kind = "stale"
	// Misclassified: a listed scenario ended differently from its status.
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd compatibility-suite && go mod tidy && go test ./baseline/ -v`
Expected: PASS for every test above.

- [ ] **Step 5: Lint**

Run: `cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./baseline/...`
Expected: no findings. Fix any (e.g. positional struct literals flagged by a linter → use field names) without changing behaviour; re-run the tests.

- [ ] **Step 6: Commit**

```bash
git add compatibility-suite/baseline compatibility-suite/go.mod compatibility-suite/go.sum
git commit -m "feat(compat): add baseline ratchet for compatibility scenarios

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Runner (godog execution + classification)

**Files:**
- Create: `compatibility-suite/adapter/adapter.go`
- Create: `compatibility-suite/runner/runner.go`
- Test: `compatibility-suite/runner/runner_test.go`
- Create: `compatibility-suite/runner/testdata/features/sample.feature`, `.../features/V1/nested.feature`, `.../dup/dup.feature`, `.../empty/.gitkeep`

**Interfaces:**
- Consumes: `baseline.Key`, `baseline.Result`, `baseline.Passed|Failed|Unsupported` (Task 2).
- Produces (used by Task 4 and PRs 2–5):
  - `adapter.ErrUnsupported error`: steps wrap it (`fmt.Errorf("...: %w", adapter.ErrUnsupported)`) when the API under test can't express a scenario.
  - `type runner.Options struct{ FeaturesDir string; Initializer func(*godog.ScenarioContext); JUnitPath string; Output io.Writer }`
  - `func runner.Run(opts Options) (map[baseline.Key]baseline.Result, error)`

Verified behaviour of godog v0.16.0 (probed while writing this plan): the `After` scenario hook receives `nil` on pass, an error matching `godog.ErrUndefined` for undefined steps, the step's error (wrapping preserved for `errors.Is`) on failure, and the panic value as an error when a step panics. `Scenario.Uri` is the feature path as given in `Paths` plus the file path below it. `Format: "progress,junit:<file>"` writes JUnit XML.

- [ ] **Step 1: Add the test fixtures**

`compatibility-suite/runner/testdata/features/sample.feature`:
```gherkin
Feature: Sample
  Background:
    Given it passes

  Scenario: passes
    Then it passes

  Scenario: fails
    Then it fails

  Scenario: undefined
    Then nobody defined this step

  Scenario: unsupported
    Then it is unsupported

  Scenario: panics
    Then it panics
```

`compatibility-suite/runner/testdata/features/V1/nested.feature`:
```gherkin
Feature: Nested
  Scenario: nested passes
    Then it passes
```

`compatibility-suite/runner/testdata/dup/dup.feature`:
```gherkin
Feature: Duplicates
  Scenario: same name
    Then it passes

  Scenario: same name
    Then it passes
```

```bash
mkdir -p compatibility-suite/runner/testdata/empty && touch compatibility-suite/runner/testdata/empty/.gitkeep
```

- [ ] **Step 2: Write the failing tests**

`compatibility-suite/runner/runner_test.go`:
```go
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
	return runner.Run(runner.Options{FeaturesDir: dir, Initializer: sampleSteps, Output: io.Discard})
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
	if _, err := run(t, "testdata/does-not-exist"); err == nil {
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
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "<testsuites") {
		t.Fatalf("not a JUnit report:\n%s", data)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd compatibility-suite && go test ./runner/`
Expected: FAIL, build errors: packages `adapter` and `runner` do not exist.

- [ ] **Step 4: Implement**

`compatibility-suite/adapter/adapter.go`:
```go
// Package adapter is the seam between the compatibility steps and the Pact Go
// API under test. Step definitions only reach Pact Go through it.
package adapter

import "errors"

// ErrUnsupported is wrapped by a step when the API under test cannot express
// the scenario at all, as opposed to expressing it and getting it wrong. Such
// scenarios are recorded as "unsupported" in the baseline.
var ErrUnsupported = errors.New("not supported by the API under test")
```

`compatibility-suite/runner/runner.go`:
```go
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
			sc.After(func(ctx context.Context, s *godog.Scenario, err error) (context.Context, error) {
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
				return ctx, err
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
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd compatibility-suite && go mod tidy && go test ./runner/ -v`
Expected: PASS for all six tests. If `TestRunRejectsMissingFeatureDir` passes only via the `no scenarios` branch rather than status 2, that's fine: the requirement is an error, not which branch produces it.

- [ ] **Step 6: Lint**

Run: `cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./...`
Expected: no findings; fix any without changing behaviour, then re-run tests.

- [ ] **Step 7: Commit**

```bash
git add compatibility-suite/adapter compatibility-suite/runner compatibility-suite/go.mod compatibility-suite/go.sum
git commit -m "feat(compat): add godog runner that classifies scenario outcomes

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Wire the suite and seed the baseline

**Files:**
- Create: `compatibility-suite/suite_test.go`
- Create: `compatibility-suite/baseline.yaml` (generated)

**Interfaces:**
- Consumes: `runner.Run`, `runner.Options` (Task 3); `baseline.Load/Update/Compare/Write` (Task 2).
- Produces: `TestCompatibility` and the `-update-baseline` test flag; env `COMPAT_JUNIT` (optional JUnit path) used by Task 5. PR 2 replaces `noSteps` with the real step initializer.

- [ ] **Step 1: Write the suite test**

`compatibility-suite/suite_test.go`:
```go
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
		if err := updated.Write(baselinePath); err != nil {
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
```

(`baseline.Update` on a nil `Baseline` from a missing file is safe: reading a nil map returns the zero value.)

- [ ] **Step 2: Run without a baseline to verify it fails**

Run: `cd compatibility-suite && go test -count=1 -run TestCompatibility .`
Expected: FAIL with `reading baseline: open baseline.yaml: no such file or directory`.

- [ ] **Step 3: Seed the baseline**

Run: `cd compatibility-suite && go test -count=1 -run TestCompatibility . -update-baseline -v 2>&1 | tail -3`
Expected: `wrote baseline.yaml: 190 of 190 scenarios listed`, PASS.

```bash
grep -c '^- feature:' baseline.yaml       # 190
grep -c 'status: failing' baseline.yaml   # 190
```

- [ ] **Step 4: Run against the baseline to verify it passes, and is stable**

```bash
cd compatibility-suite
go test -count=1 ./...
cp baseline.yaml "$TMPDIR/baseline.before.yaml"
go test -count=1 -run TestCompatibility . -update-baseline
cmp "$TMPDIR/baseline.before.yaml" baseline.yaml
```
Expected: all packages PASS; `cmp` prints nothing (the rewrite is byte-identical).

- [ ] **Step 5: Prove the ratchet bites**

Temporarily delete the first entry (4 lines) from `baseline.yaml`, run `go test -count=1 -run TestCompatibility .`, and confirm it FAILS with exactly one `regression: V1/http_consumer.feature :: ...` line. Restore with `-update-baseline` (the restored entry gets `reason: not yet triaged`, same as before).

- [ ] **Step 6: Commit**

```bash
git add compatibility-suite/suite_test.go compatibility-suite/baseline.yaml
git commit -m "feat(compat): run the compatibility suite against a seeded baseline

All 190 scenarios are recorded as failing until step definitions land.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Tasks, CI and contributor docs

**Files:**
- Modify: `mise.toml` (`[tasks.lint]`, `[tasks.verify]`, new `[tasks.compat]` after `[tasks.test-examples]`)
- Modify: `.github/workflows/test.yml` (new `compatibility` job; add to `complete.needs`)
- Create: `docs/contributing/compatibility-suite.md`

**Interfaces:**
- Consumes: `TestCompatibility`, `-update-baseline`, `COMPAT_JUNIT` (Task 4).
- Produces: `mise run compat [--update-baseline]` used by CI and later PRs.

- [ ] **Step 1: Add the `compat` task**

Insert after `[tasks.test-examples]` in `mise.toml`:
```toml
[tasks.compat]
description = "Run the Pact compatibility suite against its baseline"
# The FFI is not exercised until step definitions land, but every later PR
# needs it, so the dependency is declared once here.
depends = ["install"]
dir = "compatibility-suite"
usage = 'flag "--update-baseline" help="Rewrite baseline.yaml from this run instead of checking it"'
run = """
#!/usr/bin/env bash
set -euo pipefail
if [ -n "${usage_update_baseline:-}" ]; then
  # Only the root package defines the flag; passing it to ./... would fail
  # the other packages with "flag provided but not defined".
  go test -count=1 -run TestCompatibility . -update-baseline
else
  go test -count=1 ./...
fi
"""
```

- [ ] **Step 2: Cover the nested module in lint and verify**

`golangci-lint run` and `go mod tidy` in the root only see the root module. In `[tasks.lint]` replace the body's last line `golangci-lint run` with:
```bash
golangci-lint run
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml)
```
In `[tasks.verify]` replace `go mod tidy -diff` with:
```bash
go mod tidy -diff
(cd compatibility-suite && go mod tidy -diff)
```

- [ ] **Step 3: Verify the tasks**

```bash
mise run compat
mise run compat --update-baseline && git diff --exit-code compatibility-suite/baseline.yaml
mise run lint
mise run verify
```
Expected: compat PASS; update leaves no diff (this also confirms mise exposes the flag as `usage_update_baseline`. If the diff check passes but the log line `wrote baseline.yaml` is absent, the flag name mapping is wrong; check `mise run compat --help`); lint and verify clean.

Then prove lint now sees the nested module: add `var unused = 1` to `compatibility-suite/adapter/adapter.go`, run `mise run lint`, expect a finding, revert.

- [ ] **Step 4: Add the CI job**

In `.github/workflows/test.yml`, add `- compatibility` to `jobs.complete.needs`, and add this job after `verify`:
```yaml
  compatibility:
    runs-on: ubuntu-latest
    env:
      GO_VERSION: "1.27"
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7
      - uses: jdx/mise-action@2d8d4cafcbd33be2ea37d2b6f5ad595363d1f1ca # v5
      # The runner context is not available in job-level env, so the JUnit
      # path is set on the step from the RUNNER_TEMP shell variable.
      - run: COMPAT_JUNIT="$RUNNER_TEMP/compatibility-junit.xml" mise run compat
      - uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a # v7
        if: ${{ always() }}
        with:
          name: compatibility-junit-${{ github.run_id }}-${{ github.run_attempt }}
          path: ${{ runner.temp }}/compatibility-junit.xml
```

- [ ] **Step 5: Validate the workflow locally**

Run: `mise exec -- go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/test.yml`
Expected: no errors.

- [ ] **Step 6: Write the contributor doc**

`docs/contributing/compatibility-suite.md`:
````markdown
# Compatibility suite

Pact Go runs the shared [Pact compatibility suite](https://github.com/pact-foundation/pact-compatibility-suite)
to pin its behaviour against the Pact specification. It lives in
`compatibility-suite/`, a separate Go module, so its dependencies never reach
Pact Go users.

## Running it

```sh
mise run compat
```

The run is checked against `compatibility-suite/baseline.yaml`, which lists
every scenario that is **not** expected to pass:

- `failing`: Pact Go gets this scenario wrong today.
- `unsupported`: the API under test cannot express the scenario (a step
  returned `adapter.ErrUnsupported`).

The suite fails when:

| What happened | Reported as | What to do |
|---|---|---|
| An unlisted scenario failed | `regression` | Fix the regression. |
| A listed scenario now passes | `fixed` | Remove it from the baseline. |
| A listed scenario no longer exists upstream | `stale` | Remove it from the baseline. |
| A scenario ended differently from its listed status | `misclassified` | Update its status. |

## Updating the baseline

```sh
mise run compat --update-baseline
```

This rewrites `baseline.yaml` from the current run, keeping any reason you
have written for an entry whose status hasn't changed. New failures get
`reason: not yet triaged`; replace that with a short explanation. Review the
diff like any other change: a removed entry is an improvement, an added one
needs a reason.

## Updating the upstream suite

The upstream features are vendored with `git subtree`. Never edit them in
place.

```sh
git subtree pull --prefix=compatibility-suite/pact-compatibility-suite \
  https://github.com/pact-foundation/pact-compatibility-suite.git main --squash
mise run compat --update-baseline
```

Do this in its own PR so the baseline diff shows exactly what upstream changed.
````

Also add a link to it from `docs/contributing/developer.md` (append a line: `See [compatibility-suite.md](compatibility-suite.md) for the Pact compatibility suite.`).

- [ ] **Step 7: Commit**

```bash
git add mise.toml .github/workflows/test.yml docs/contributing/compatibility-suite.md docs/contributing/developer.md
git commit -m "ci(compat): run the compatibility suite as a required check

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Open PR 1 as the base of the stack

- [ ] **Step 1: Final local check**

```bash
mise run lint && mise run verify && mise run compat && mise run test
git log --oneline master..HEAD
```
Expected: all green; commits for spec, plan, subtree, module, baseline, runner, suite, CI.

- [ ] **Step 2: Create the stack and open PR 1**

Using the `gh stack` commands noted in Task 0, initialise a stack with `compat/01-scaffold` on `master` and submit it as a PR. Title: `feat: add the Pact compatibility suite with a ratchet baseline (1/5)`. Body:

```markdown
First of a stack adding the [Pact compatibility suite](https://github.com/pact-foundation/pact-compatibility-suite)
as a behavioural baseline before upcoming architecture changes.
Design: `docs/superpowers/specs/2026-10-06-compatibility-suite-design.md`.

This PR adds the harness only:
- upstream features vendored via `git subtree` at `dd08122`
- nested `compatibility-suite/` module (nothing added to pact-go's `go.mod`)
- godog runner plus `baseline.yaml` ratchet; all 190 scenarios start as `failing`
- `mise run compat` and a required `compatibility` CI job

Following PRs in the stack add step definitions and remove scenarios from the baseline.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

- [ ] **Step 3: Confirm CI**

Run: `gh pr checks --watch`
Expected: `compatibility` job green and listed as required via `Test completion check`.
