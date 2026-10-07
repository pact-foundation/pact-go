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
	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	//nolint:gosec // G304: path is a file this test wrote under t.TempDir().
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
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
		"unlisted pass":           {nil, pass, ""},
		"unlisted fail":           {nil, fail, baseline.Regression},
		"unlisted unsupported":    {nil, unsup, baseline.Regression},
		"failing still fails":     {&failing, fail, ""},
		"failing now passes":      {&failing, pass, baseline.Fixed},
		"failing now unsupported": {&failing, unsup, baseline.Misclassified},
		"unsupported still is":    {&unsupported, unsup, ""},
		"unsupported now passes":  {&unsupported, pass, baseline.Fixed},
		"unsupported now fails":   {&unsupported, fail, baseline.Misclassified},
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
	err := b.Write(path)
	if err != nil {
		t.Fatal(err)
	}
	first := readFile(t, path)

	loaded, err := baseline.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded, b) {
		t.Fatalf("round trip: got %#v, want %#v", loaded, b)
	}
	err = loaded.Write(path)
	if err != nil {
		t.Fatal(err)
	}
	second := readFile(t, path)
	if first != second {
		t.Fatalf("rewrite changed the file:\n%s\n---\n%s", first, second)
	}
	if a, c := strings.Index(first, "alpha"), strings.Index(first, "gamma"); a > c {
		t.Fatalf("entries not sorted by key:\n%s", first)
	}
}
