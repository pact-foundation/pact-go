# Compatibility Suite — PR 2: V1 HTTP Steps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the 50 V1 HTTP scenarios (`V1/http_consumer.feature`, `V1/http_provider.feature`) run against pact-go, drop the ones that pass from `baseline.yaml`, and record every remaining one with a real reason.

**Architecture:**
- **Steps:** step definitions in `compatibility-suite/steps` reach pact-go only through `adapter.Adapter`.
- **The v2 adapter:**
  - It drives scenarios through pact-go's public API (`consumer.NewV2Pact`/`ExecuteTest`, `provider.NewVerifier().VerifyProvider`), each under a sandboxed `*testing.T` obtained via `testing.RunTests`.
  - It reads back what pact-go does not return as data (mock server mismatches, verifier results) through two small read-only hooks added to `internal/native`.
- **Fakes:** an in-process provider stub and Pact Broker stand in for the external services.

**Tech Stack:** Go 1.25+, godog v0.16.0, pact-go v2 (via `replace => ../`), pact_ffi, `net/http/httptest`, mise, `gh stack`.

**Spec:** `docs/superpowers/specs/2026-10-06-compatibility-suite-design.md` (updated on this branch: "How `adapter/v2` drives and checks", "Upstream sync").

## Global Constraints

- **Driving:** scenarios are **driven only through pact-go's public API**.
- **Checking:** `internal/native` may be used **only to read back** results pact-go does not return as data: `native.MockServerMismatchesJSON`, `native.ResultObserver`.
- **Package boundaries:**
  - `steps/` never imports pact-go.
  - `adapter/v2` is the only package that does.
- **Spec version:** pact-go has no V1 specification, so V1 features run as **V2 pacts**.
- **Binary bodies:** pact-go cannot express a binary body with its own content type, so one is `adapter.ErrUnsupported` (wrapped, with a reason).
- **Environment:** no Docker, no network. The provider and broker are `httptest` fakes.
- **Upstream sync:** upstream is vendored by `mise run compat-sync` and pinned in `compatibility-suite/pact-compatibility-suite.ref`. Upstream files are never edited by hand.
- **Lint:** all code passes `.golangci.yml` (`default: all`, gofumpt). Use plain `err := ...; if err != nil` (the `noinlineerr` linter rejects `if err := ...;`), and `//nolint:gosec // G304: <reason>` for fixture or file reads.
- **Commits:** messages end with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

## Review Focus

1. **A scenario that fails mid-way:** an earlier step errors while `ExecuteTest` is still blocked in its goroutine. The After hook must unblock it, and `Finish` must be safe to call twice, rather than hanging the suite. → Task 4, `TestConsumerFinishIsRepeatable`.
2. **A verification that never runs:** pact-go can reject its own config before calling the FFI, e.g. a broker with no provider name. This must surface as an error carrying pact-go's message, not as a pass or an empty result. → Task 5, `TestVerifierReportsWhenNothingRan`.
3. **The observer hook leaking between verifications:** `native.ResultObserver` must be cleared after every verification, so a later caller doesn't read a stale result. → Task 5, `TestVerifierClearsObserver`.
4. **Fixture files checked out with CRLF line endings:** for example on Windows with `core.autocrlf`. An `eol="CRLF"` body must not end up with `\r\r\n`. → Task 3, `TestBodyXMLFixtureCheckedOutWithCRLF`.
5. **A change table naming a column the steps don't handle:** this must fail with the column's name, not silently leave the request unchanged and let a negative scenario pass for the wrong reason. → Task 7, `TestRequestChangesRejectUnknownColumn`.

---

## File Structure

```
mise.toml                                       # + compat-sync task                         (Task 1)
compatibility-suite/pact-compatibility-suite.ref                                             (Task 1)
docs/contributing/compatibility-suite.md        # sync + how scenarios are driven            (Tasks 1, 9)
internal/native/mock_server.go                  # + MockServerMismatchesJSON                 (Task 2)
internal/native/verifier.go                     # + ResultObserver, resultJSON               (Task 2)
compatibility-suite/
  model/model.go, model/query.go (+ tests)      # interactions, mismatches, ParseQuery       (Task 3)
  fixtures/fixtures.go (+ test)                 # body cells and fixture files               (Task 3)
  adapter/adapter.go                            # Adapter, consumer + verifier interfaces    (Tasks 4, 5)
  adapter/v2/v2.go, sandbox.go                  # New(), sandboxed *testing.T                (Task 4)
  adapter/v2/consumer.go, mismatches.go (+tests)                                             (Task 4)
  adapter/v2/verifier.go, result.go (+ tests)                                                (Task 5)
  fakes/provider.go, fakes/broker.go (+ tests)                                               (Task 6)
  steps/table.go, steps/pactfile.go, steps/http.go (+ tests)                                 (Task 7)
  steps/steps.go, steps/consumer.go, steps/provider.go                                       (Task 8)
  suite_test.go                                 # real steps, -adapter flag                  (Task 8)
  baseline.yaml                                 # V1 entries removed or triaged              (Task 9)
```

The branch `compat/02-v1-http` already exists as a new `gh stack` on `master`, with the spec update uncommitted. Task 1 commits it.

---

### Task 1: Replace subtree sync with a pinned fetch

**Files:**
- Modify: `mise.toml` (new `[tasks.compat-sync]` after `[tasks.compat]`)
- Create: `compatibility-suite/pact-compatibility-suite.ref`
- Modify: `docs/contributing/compatibility-suite.md` (section "Updating the upstream suite")
- Commit: `docs/superpowers/specs/2026-10-06-compatibility-suite-design.md` (already edited)

**Interfaces:**
- Produces: `mise run compat-sync [--ref <ref>]`. The `.ref` file contains one line, the 40-char SHA.

- [ ] **Step 1: Add the task**

Insert after the `[tasks.compat]` block in `mise.toml`:
```toml
[tasks.compat-sync]
description = "Vendor the upstream compatibility suite at a branch, tag or commit"
dir = "compatibility-suite"
usage = 'flag "--ref <ref>" default="main" help="Upstream branch, tag or commit to vendor"'
# Replaces the vendored tree wholesale and records the commit it came from.
# This works whatever merge strategy the PR uses, unlike git subtree, whose
# metadata a squash merge drops.
run = """
#!/usr/bin/env bash
set -euo pipefail
repo=https://github.com/pact-foundation/pact-compatibility-suite.git
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q
git -C "$tmp" fetch -q --depth 1 "$repo" "$usage_ref"
sha=$(git -C "$tmp" rev-parse FETCH_HEAD)
rm -rf pact-compatibility-suite
mkdir pact-compatibility-suite
git -C "$tmp" archive FETCH_HEAD | tar -x -C pact-compatibility-suite
echo "$sha" > pact-compatibility-suite.ref
echo "vendored pact-compatibility-suite at $sha"
"""
```

- [ ] **Step 2: Run it at the commit already vendored**

```bash
mise run compat-sync --ref dd08122ce7343730fd373574a91d046e8814be43
git status --short compatibility-suite/
cat compatibility-suite/pact-compatibility-suite.ref
```
Expected:
- The status shows only `?? compatibility-suite/pact-compatibility-suite.ref`. The re-vendored tree is byte-identical to the subtree import.
- The file contains `dd08122ce7343730fd373574a91d046e8814be43`.

If any upstream file shows as modified or deleted, stop and diff it. The script must reproduce the tree exactly.

- [ ] **Step 3: Update the contributor doc**

In `docs/contributing/compatibility-suite.md`, replace the whole `## Updating the upstream suite` section with:
````markdown
## Updating the upstream suite

The upstream features are vendored into
`compatibility-suite/pact-compatibility-suite/`, and the commit they came
from is recorded in `compatibility-suite/pact-compatibility-suite.ref`. Never
edit the vendored files in place.

```sh
mise run compat-sync                 # latest upstream main
mise run compat-sync --ref <sha>     # or a specific branch, tag or commit
mise run compat --update-baseline
```

Do this in its own PR so the baseline diff shows exactly what upstream changed.
````

- [ ] **Step 4: Commit**

```bash
git add mise.toml compatibility-suite/pact-compatibility-suite.ref docs/contributing/compatibility-suite.md docs/superpowers/specs/2026-10-06-compatibility-suite-design.md docs/superpowers/plans/2026-10-08-compatibility-suite-02-v1-http.md
git commit -m "chore(compat): vendor the upstream suite with a pinned fetch

PR 1's squash merge dropped the git subtree metadata, so subtree pull no
longer works. compat-sync replaces the tree at a given ref and records the
commit in pact-compatibility-suite.ref.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 2: Read-back hooks in internal/native

**Files:**
- Modify: `internal/native/mock_server.go` (add after `MockServerMismatchedRequests`)
- Modify: `internal/native/verifier.go` (`Execute`, plus new declarations)
- Test: `internal/native/mock_server_test.go`, `internal/native/verifier_test.go` (build tag `provider`)

**Interfaces:**
- Produces (used by Tasks 4, 5):
  - `func native.MockServerMismatchesJSON(port int) string`: pact_ffi's raw mismatch report for the mock server on `port`, or `""` when there is none.
  - `var native.ResultObserver func(resultJSON string)`: when non-nil, it is called with `pactffi_verifier_json` right after every `Verifier.Execute`. It is nil in normal use.

- [ ] **Step 1: Write the failing tests**

Append to `internal/native/mock_server_test.go`:
```go
func TestMockServerMismatchesJSONKeepsBodyPath(t *testing.T) {
	m := NewHTTPPact("consumer", "provider")
	m.NewInteraction("json body").
		UponReceiving("json body").
		WithRequest("PUT", "/basic").
		WithJSONRequestBody(map[string]string{"one": "a", "two": "b"}).
		WithStatus(200)
	port, err := m.Start("0.0.0.0:0", false)
	require.NoError(t, err)
	defer m.CleanupMockServer(port)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut,
		fmt.Sprintf("http://localhost:%d/basic", port), strings.NewReader(`{"one":"a","two":"c"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	assert.Contains(t, MockServerMismatchesJSON(port), `"path":"$.two"`)
}
```

Append to `internal/native/verifier_test.go`, and add `"github.com/stretchr/testify/require"` to its imports:
```go
func TestVerifier_ExecuteReportsResultToObserver(t *testing.T) {
	var got string
	ResultObserver = func(resultJSON string) { got = resultJSON }
	defer func() { ResultObserver = nil }()

	v := NewVerifier("pact-go", "test")
	defer v.Shutdown()
	require.NoError(t, v.Execute())

	assert.Contains(t, got, `"result":true`)
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go test -tags=provider -run 'MismatchesJSON|ReportsResultToObserver' ./internal/native/`
Expected: FAIL to build with `undefined: MockServerMismatchesJSON` and `undefined: ResultObserver`.

- [ ] **Step 3: Implement**

In `internal/native/mock_server.go`, after `MockServerMismatchedRequests`:
```go
// MockServerMismatchesJSON returns pact_ffi's mismatch report for the mock
// server on port, unparsed, or "" when there is none. Unlike
// MockServerMismatchedRequests it keeps every field, such as the body path
// of a mismatch. It exists so the compatibility suite can check results that
// pact-go does not return as data.
func MockServerMismatchesJSON(port int) string {
	mismatches := C.pactffi_mock_server_mismatches(C.int(port))
	if mismatches == nil {
		return ""
	}

	return C.GoString(mismatches)
}
```

In `internal/native/verifier.go`, add above `// Shutdown releases`:
```go
// ResultObserver, when set, receives each verification's result as pact_ffi
// reports it (pactffi_verifier_json), right after Execute runs. It exists so
// the compatibility suite can check results that pact-go returns only as a
// sentinel error, and is nil in normal use. It must not be changed while a
// verification runs.
var ResultObserver func(resultJSON string)
```
In `Execute`, directly after `result := C.pactffi_verifier_execute(v.handle)`:
```go
	if ResultObserver != nil {
		ResultObserver(v.resultJSON())
	}
```
And add, after `Execute`:
```go
// resultJSON returns the verification result of v as JSON.
func (v *Verifier) resultJSON() string {
	out := C.pactffi_verifier_json(v.handle)
	if out == nil {
		return ""
	}
	defer C.pactffi_free_string(out)

	return C.GoString(out)
}
```
(While writing the plan, a throwaway build in a separate worktree confirmed that `pactffi_free_string` accepts the `pactffi_verifier_json` result.)

- [ ] **Step 4: Run them to verify they pass, then the package**

Run: `go test -tags=provider -count=1 ./internal/native/`
Expected: PASS, including the two new tests.

- [ ] **Step 5: Lint and commit**

```bash
golangci-lint run ./internal/native/...
git add internal/native/mock_server.go internal/native/verifier.go internal/native/mock_server_test.go internal/native/verifier_test.go
git commit -m "feat(native): expose raw mismatch and verifier results for checking

Read-only hooks for the compatibility suite: the mock server's full
mismatch report, and an observer for the verifier's JSON result. Neither
changes behaviour; ResultObserver is nil unless a caller sets it.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 3: Model and fixtures

**Files:**
- Create: `compatibility-suite/model/model.go`, `compatibility-suite/model/query.go`
- Test: `compatibility-suite/model/model_test.go`
- Create: `compatibility-suite/fixtures/fixtures.go`
- Test: `compatibility-suite/fixtures/fixtures_test.go`

**Interfaces:**
- Produces (every later task uses these):
  - `model.Body{ContentType string; Contents []byte}`
  - `model.Request{Method, Path string; Query url.Values; Headers http.Header; Body *Body}`
  - `model.Response{Status int; Headers http.Header; Body *Body}`
  - `model.Interaction{Number int; Description string; ProviderStates []string; Request Request; Response Response}`, `func (Interaction) Clone() Interaction`
  - `model.Mismatch{Type, Method, Path string; Query url.Values; Details []MismatchDetail}`
  - `model.MismatchDetail{Type, Location, Message string}`
  - `model.VerificationError{Interaction string; MismatchTypes []string; Message string}`
  - `func model.ParseQuery(raw json.RawMessage) (url.Values, error)`
  - `fixtures.Loader{Dir string}` with:
    - `Read(name string) ([]byte, error)`
    - `Body(cell, fallback string) (*model.Body, error)`
    - `File(name, fallback string) (*model.Body, error)`

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/model/model_test.go`:
```go
package model_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestInteractionCloneIsDeep(t *testing.T) {
	original := model.Interaction{
		ProviderStates: []string{"a"},
		Request: model.Request{
			Query:   url.Values{"a": {"1"}},
			Headers: http.Header{"X": {"1"}},
			Body:    &model.Body{ContentType: "text/plain", Contents: []byte("hi")},
		},
		Response: model.Response{
			Headers: http.Header{"Y": {"1"}},
			Body:    &model.Body{ContentType: "text/plain", Contents: []byte("ok")},
		},
	}

	c := original.Clone()
	c.ProviderStates[0] = "b"
	c.Request.Query.Set("a", "2")
	c.Request.Headers.Set("X", "2")
	c.Request.Body.Contents[0] = 'H'
	c.Response.Headers.Set("Y", "2")
	c.Response.Body.Contents[0] = 'O'

	if original.ProviderStates[0] != "a" || original.Request.Query.Get("a") != "1" ||
		original.Request.Headers.Get("X") != "1" || string(original.Request.Body.Contents) != "hi" ||
		original.Response.Headers.Get("Y") != "1" || string(original.Response.Body.Contents) != "ok" {
		t.Fatalf("changing the clone changed the original: %+v", original)
	}
}

func TestCloneKeepsNilBodies(t *testing.T) {
	c := model.Interaction{}.Clone()
	if c.Request.Body != nil || c.Response.Body != nil {
		t.Fatalf("want nil bodies, got %+v", c)
	}
}

func TestParseQuery(t *testing.T) {
	cases := map[string]struct {
		raw  string
		want url.Values
	}{
		"absent":       {"", nil},
		"null":         {"null", nil},
		"V2 string":    {`"a=1&b=2"`, url.Values{"a": {"1"}, "b": {"2"}}},
		"V3 map":       {`{"a":["1"],"b":["2","3"]}`, url.Values{"a": {"1"}, "b": {"2", "3"}}},
		"empty string": {`""`, url.Values{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := model.ParseQuery(json.RawMessage(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseQueryRejectsOtherJSON(t *testing.T) {
	_, err := model.ParseQuery(json.RawMessage(`42`))
	if err == nil {
		t.Fatal("want an error for a numeric query")
	}
}
```

`compatibility-suite/fixtures/fixtures_test.go`:
```go
package fixtures_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
)

var fx = fixtures.Loader{Dir: "../pact-compatibility-suite/fixtures"}

func TestBodyFromJSONCell(t *testing.T) {
	got, err := fx.Body(`JSON: {"one": "a", "two": "c"}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/json" || string(got.Contents) != `{"one": "a", "two": "c"}` {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromXMLCell(t *testing.T) {
	got, err := fx.Body(`XML: <?xml version="1.0"?><values/>`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/xml" || string(got.Contents) != `<?xml version="1.0"?><values/>` {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromJSONFixture(t *testing.T) {
	got, err := fx.Body("file: basic.json", "")
	if err != nil {
		t.Fatal(err)
	}
	want, err := fx.Read("basic.json")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/json" || !bytes.Equal(got.Contents, want) {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromTextBodyFixture(t *testing.T) {
	got, err := fx.Body("file: text-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "text/plain" || string(got.Contents) != "Hello World!" {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}
}

func TestBodyFromXMLBodyFixtureKeepsTheDocument(t *testing.T) {
	got, err := fx.Body("file: xml-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(got.Contents)
	if got.ContentType != "application/xml" || !strings.HasPrefix(contents, "<?xml") ||
		!strings.HasSuffix(contents, "</values>") {
		t.Fatalf("got %q %q", got.ContentType, contents)
	}
}

func TestBodyFromMultipartFixtureUsesCRLF(t *testing.T) {
	got, err := fx.Body("file: multipart-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(got.Contents)
	if !strings.HasPrefix(got.ContentType, "multipart/mixed; boundary=gc0p4Jq0M2Yt08jU534c0p") {
		t.Fatalf("content type %q", got.ContentType)
	}
	if !strings.HasPrefix(contents, "--gc0p4Jq0M2Yt08jU534c0p\r\n") {
		t.Fatalf("contents start %q", contents[:min(40, len(contents))])
	}
	if strings.Count(contents, "\n") != strings.Count(contents, "\r\n") {
		t.Fatalf("contents have bare LF line endings: %q", contents)
	}
}

func TestBodyFromBinaryFixture(t *testing.T) {
	got, err := fx.Body("file: rat.jpg", "")
	if err != nil {
		t.Fatal(err)
	}
	want, err := fx.Read("rat.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "image/jpeg" || !bytes.Equal(got.Contents, want) {
		t.Fatalf("got %q with %d bytes", got.ContentType, len(got.Contents))
	}
}

func TestBodyFromTextUsesFallback(t *testing.T) {
	got, err := fx.Body("a=1&b=2&c=33&d=4", "application/x-www-form-urlencoded")
	if err != nil {
		t.Fatal(err)
	}
	if got.ContentType != "application/x-www-form-urlencoded" || string(got.Contents) != "a=1&b=2&c=33&d=4" {
		t.Fatalf("got %q %q", got.ContentType, got.Contents)
	}

	plain, err := fx.Body("Hello Mars!", "")
	if err != nil {
		t.Fatal(err)
	}
	if plain.ContentType != "text/plain" {
		t.Fatalf("want text/plain without a fallback, got %q", plain.ContentType)
	}
}

func TestBodyEmptyMarker(t *testing.T) {
	got, err := fx.Body("EMPTY", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Contents) != 0 {
		t.Fatalf("want an empty body, got %q", got.Contents)
	}
}

func TestBodyRejectsEmptyCell(t *testing.T) {
	_, err := fx.Body("  ", "")
	if err == nil {
		t.Fatal("want an error for an empty cell")
	}
}

func TestBodyMissingFixtureNamesIt(t *testing.T) {
	_, err := fx.Body("file: nope.json", "")
	if err == nil || !strings.Contains(err.Error(), "nope.json") {
		t.Fatalf("want an error naming nope.json, got %v", err)
	}
}

func TestBodyXMLFixtureCheckedOutWithCRLF(t *testing.T) {
	dir := t.TempDir()
	doc := "<body>\r\n<contentType>text/plain</contentType>\r\n" +
		"<contents eol=\"CRLF\"><![CDATA[line one\r\nline two]]></contents>\r\n</body>\r\n"
	err := os.WriteFile(filepath.Join(dir, "crlf-body.xml"), []byte(doc), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	got, err := fixtures.Loader{Dir: dir}.Body("file: crlf-body.xml", "")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Contents) != "line one\r\nline two" {
		t.Fatalf("got %q", got.Contents)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go -C compatibility-suite test ./model/ ./fixtures/`
Expected: FAIL to build, because packages `model` and `fixtures` have no non-test files.

- [ ] **Step 3: Implement**

`compatibility-suite/model/model.go`:
```go
// Package model describes HTTP interactions and their results independently
// of any Pact implementation, so the steps and the adapters share one
// vocabulary.
package model

import (
	"bytes"
	"net/http"
	"net/url"
	"slices"
)

// Body is the contents of a request or response, with its content type.
type Body struct {
	ContentType string
	Contents    []byte
}

// Request is the HTTP request an interaction expects.
type Request struct {
	Method  string
	Path    string
	Query   url.Values
	Headers http.Header
	Body    *Body
}

// Response is the HTTP response an interaction returns.
type Response struct {
	Status  int
	Headers http.Header
	Body    *Body
}

// Interaction is one HTTP interaction from a feature's table.
type Interaction struct {
	// Number is the interaction's 1-based row in its table, as steps refer
	// to it.
	Number         int
	Description    string
	ProviderStates []string
	Request        Request
	Response       Response
}

// Clone returns a deep copy, so a step can change an interaction without
// changing the table it came from.
func (i Interaction) Clone() Interaction {
	c := i
	c.ProviderStates = slices.Clone(i.ProviderStates)
	c.Request.Query = url.Values(http.Header(i.Request.Query).Clone())
	c.Request.Headers = i.Request.Headers.Clone()
	c.Request.Body = i.Request.Body.clone()
	c.Response.Headers = i.Response.Headers.Clone()
	c.Response.Body = i.Response.Body.clone()

	return c
}

func (b *Body) clone() *Body {
	if b == nil {
		return nil
	}

	return &Body{ContentType: b.ContentType, Contents: bytes.Clone(b.Contents)}
}

// Mismatch is one problem a mock server recorded.
type Mismatch struct {
	// Type is missing-request, request-not-found or request-mismatch.
	Type   string
	Method string
	Path   string
	// Query is the request's query, when the mock server reports it.
	Query   url.Values
	Details []MismatchDetail
}

// MismatchDetail is one difference within a mismatched request.
type MismatchDetail struct {
	// Type is the Pact mismatch type, such as QueryMismatch or BodyMismatch.
	Type string
	// Location is the query parameter, header name or body path concerned.
	Location string
	Message  string
}

// VerificationError is one failure a provider verification reported.
type VerificationError struct {
	Interaction string
	// MismatchTypes are the Pact mismatch types found, such as
	// StatusMismatch.
	MismatchTypes []string
	// Message is set when the interaction failed with an error rather than
	// with mismatches.
	Message string
}
```

`compatibility-suite/model/query.go`:
```go
package model

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// ParseQuery reads a query the way pact files and pact_ffi write one: as a
// map of values (V3 and later) or as a query string (V2). An absent or null
// query is nil.
func ParseQuery(raw json.RawMessage) (url.Values, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var values map[string][]string
	if json.Unmarshal(raw, &values) == nil {
		return values, nil
	}

	var text string
	err := json.Unmarshal(raw, &text)
	if err != nil {
		return nil, fmt.Errorf("unrecognised query %s: %w", raw, err)
	}

	query, err := url.ParseQuery(text)
	if err != nil {
		return nil, fmt.Errorf("parsing query %q: %w", text, err)
	}

	return query, nil
}
```

`compatibility-suite/fixtures/fixtures.go`:
```go
// Package fixtures reads the upstream suite's fixture files and the body
// cells of its feature tables.
package fixtures

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// Loader reads fixtures from the upstream suite's fixtures directory.
type Loader struct {
	Dir string
}

// Read returns a fixture file's raw contents.
func (l Loader) Read(name string) ([]byte, error) {
	//nolint:gosec // G304: name comes from the vendored upstream feature files.
	data, err := os.ReadFile(filepath.Join(l.Dir, name))
	if err != nil {
		return nil, fmt.Errorf("reading fixture %s: %w", name, err)
	}

	return data, nil
}

// Body parses a body cell:
//   - "JSON: <document>" and "XML: <document>" give that document;
//   - "file: <name>" loads a fixture (see File);
//   - anything else is literal text, with "EMPTY" meaning an empty body.
//
// fallback is the content type for text whose type cannot be inferred. A
// step changing a body passes the existing body's content type, as the
// reference implementation does.
func (l Loader) Body(cell, fallback string) (*model.Body, error) {
	cell = strings.TrimSpace(cell)

	switch {
	case cell == "":
		return nil, errors.New("a body cell is empty")
	case strings.HasPrefix(cell, "JSON:"):
		return &model.Body{
			ContentType: "application/json",
			Contents:    []byte(strings.TrimSpace(strings.TrimPrefix(cell, "JSON:"))),
		}, nil
	case strings.HasPrefix(cell, "XML:"):
		return &model.Body{
			ContentType: "application/xml",
			Contents:    []byte(strings.TrimSpace(strings.TrimPrefix(cell, "XML:"))),
		}, nil
	case strings.HasPrefix(cell, "file:"):
		return l.File(strings.TrimSpace(strings.TrimPrefix(cell, "file:")), fallback)
	}

	text := cell
	if cell == "EMPTY" {
		text = ""
	}

	return &model.Body{ContentType: contentType(cell, fallback), Contents: []byte(text)}, nil
}

// File loads a fixture as a body. A "*-body.xml" fixture describes a body
// (its content type and contents); any other file is the body itself, typed
// by its file name.
func (l Loader) File(name, fallback string) (*model.Body, error) {
	data, err := l.Read(name)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(name, "-body.xml") {
		return parseBodyXML(name, data)
	}

	return &model.Body{ContentType: contentType(name, fallback), Contents: data}, nil
}

// contentType infers a content type from a file name or text the way the
// reference implementation does, falling back to fallback, then text/plain.
func contentType(name, fallback string) string {
	switch {
	case strings.HasSuffix(name, ".json"):
		return "application/json"
	case strings.HasSuffix(name, ".xml"):
		return "application/xml"
	case strings.HasSuffix(name, ".jpg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".pdf"):
		return "application/pdf"
	case fallback != "":
		return fallback
	default:
		return "text/plain"
	}
}

type bodyXML struct {
	ContentType string `xml:"contentType"`
	Contents    struct {
		EOL  string `xml:"eol,attr"`
		Text string `xml:",chardata"`
	} `xml:"contents"`
}

// parseBodyXML reads a body fixture: <body><contentType/><contents/></body>.
func parseBodyXML(name string, data []byte) (*model.Body, error) {
	var doc bodyXML

	err := xml.Unmarshal(data, &doc)
	if err != nil {
		return nil, fmt.Errorf("parsing body fixture %s: %w", name, err)
	}

	contents := strings.TrimSpace(doc.Contents.Text)
	// Normalise first, so a checkout with CRLF line endings is not doubled up.
	contents = strings.ReplaceAll(contents, "\r\n", "\n")
	if doc.Contents.EOL == "CRLF" {
		contents = strings.ReplaceAll(contents, "\n", "\r\n")
	}

	ct := strings.TrimSpace(doc.ContentType)
	if ct == "" {
		ct = "text/plain"
	}

	return &model.Body{ContentType: ct, Contents: []byte(contents)}, nil
}
```

- [ ] **Step 4: Run them to verify they pass**

Run: `go -C compatibility-suite test -count=1 ./model/ ./fixtures/`
Expected: PASS.

- [ ] **Step 5: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./model/... ./fixtures/...)
git add compatibility-suite/model compatibility-suite/fixtures
git commit -m "feat(compat): add the interaction model and fixture loader

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 4: Consumer adapter

**Files:**
- Modify: `compatibility-suite/adapter/adapter.go`
- Create: `compatibility-suite/adapter/v2/v2.go`, `sandbox.go`, `consumer.go`, `mismatches.go`
- Test: `compatibility-suite/adapter/v2/mismatches_test.go`, `consumer_test.go`

**Interfaces:**
- Consumes: `model.*`, `model.ParseQuery` (Task 3); `native.MockServerMismatchesJSON` (Task 2).
- Produces:
  - `adapter.Adapter{Consumer HTTPConsumer; Verifier HTTPVerifier}` (Task 5 adds `Verifier`)
  - `adapter.ConsumerPact{Consumer, Provider, PactDir string}`
  - `adapter.HTTPConsumer` with `Start(pact ConsumerPact, interactions []model.Interaction) (MockServer, error)`
  - `adapter.MockServer` with `URL() string`, `Mismatches() ([]model.Mismatch, error)` and `Finish() Outcome`
  - `adapter.Outcome{Err error; Mismatches []model.Mismatch}`
  - `func v2.New() adapter.Adapter`
  - The pact file is written to `<PactDir>/<Consumer>-<Provider>.json`.

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/adapter/v2/mismatches_test.go`:
```go
package v2

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestParseMismatches(t *testing.T) {
	data := []byte(`[
	  {"type":"request-mismatch","method":"PUT","path":"/basic","mismatches":[
	    {"type":"BodyMismatch","path":"$.two","expected":"\"b\"","actual":"\"c\"","mismatch":"Expected 'c' (String) to be equal to 'b' (String)"},
	    {"type":"QueryMismatch","parameter":"b","mismatch":"Expected query parameter 'b' but was missing"},
	    {"type":"HeaderMismatch","key":"X-TEST","mismatch":"Expected a header 'X-TEST' but was missing"}]},
	  {"type":"request-not-found","method":"GET","path":"/with_params","request":{"method":"GET","path":"/with_params","query":{"a":["1"],"b":["2"]}}},
	  {"type":"missing-request","method":"GET","path":"/basic","request":{"method":"GET","path":"/basic","query":"x=1"}}
	]`)

	got, err := parseMismatches(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Mismatch{
		{Type: "request-mismatch", Method: "PUT", Path: "/basic", Details: []model.MismatchDetail{
			{Type: "BodyMismatch", Location: "$.two", Message: "Expected 'c' (String) to be equal to 'b' (String)"},
			{Type: "QueryMismatch", Location: "b", Message: "Expected query parameter 'b' but was missing"},
			{Type: "HeaderMismatch", Location: "X-TEST", Message: "Expected a header 'X-TEST' but was missing"},
		}},
		{Type: "request-not-found", Method: "GET", Path: "/with_params", Query: url.Values{"a": {"1"}, "b": {"2"}}},
		{Type: "missing-request", Method: "GET", Path: "/basic", Query: url.Values{"x": {"1"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %#v\nwant %#v", got, want)
	}
}

func TestParseMismatchesRejectsMalformedJSON(t *testing.T) {
	_, err := parseMismatches([]byte(`{not json`))
	if err == nil {
		t.Fatal("want an error for malformed JSON")
	}
}
```

`compatibility-suite/adapter/v2/consumer_test.go`:
```go
package v2

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

var basicJSON = []byte(`{"one": "a", "two": "b"}`)

func getBasic() model.Interaction {
	return model.Interaction{
		Number: 1, Description: "get basic",
		Request:  model.Request{Method: http.MethodGet, Path: "/basic"},
		Response: model.Response{Status: http.StatusOK, Body: &model.Body{ContentType: "application/json", Contents: basicJSON}},
	}
}

func putBasic() model.Interaction {
	return model.Interaction{
		Number: 2, Description: "put basic",
		Request:  model.Request{Method: http.MethodPut, Path: "/basic", Body: &model.Body{ContentType: "application/json", Contents: basicJSON}},
		Response: model.Response{Status: http.StatusOK},
	}
}

func startConsumer(t *testing.T, interactions ...model.Interaction) (adapter.MockServer, string) {
	t.Helper()
	dir := t.TempDir()
	server, err := consumerAdapter{}.Start(adapter.ConsumerPact{Consumer: "c", Provider: "p", PactDir: dir}, interactions)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { server.Finish() })

	return server, dir
}

func send(t *testing.T, method, target, contentType, body string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	return res.StatusCode
}

func hasDetail(mismatches []model.Mismatch, kind, location string) bool {
	for _, m := range mismatches {
		for _, d := range m.Details {
			if d.Type == kind && d.Location == location {
				return true
			}
		}
	}

	return false
}

func TestConsumerPassesAndWritesPact(t *testing.T) {
	server, dir := startConsumer(t, getBasic())
	status := send(t, http.MethodGet, server.URL()+"/basic", "", "")
	if status != http.StatusOK {
		t.Fatalf("want 200, got %d", status)
	}

	outcome := server.Finish()
	if outcome.Err != nil {
		t.Fatalf("want the test to pass, got %v", outcome.Err)
	}
	_, err := os.Stat(filepath.Join(dir, "c-p.json"))
	if err != nil {
		t.Fatalf("want a pact file: %v", err)
	}
}

func TestConsumerReportsMismatchDetailsWhileRunningAndAfter(t *testing.T) {
	server, dir := startConsumer(t, putBasic())
	status := send(t, http.MethodPut, server.URL()+"/basic", "application/json", `{"one": "a", "two": "c"}`)
	if status != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", status)
	}

	running, err := server.Mismatches()
	if err != nil {
		t.Fatal(err)
	}
	if !hasDetail(running, "BodyMismatch", "$.two") {
		t.Fatalf("want a BodyMismatch at $.two while running, got %+v", running)
	}

	outcome := server.Finish()
	if outcome.Err == nil {
		t.Fatal("want the test to fail")
	}
	if !hasDetail(outcome.Mismatches, "BodyMismatch", "$.two") {
		t.Fatalf("want a BodyMismatch at $.two after finishing, got %+v", outcome.Mismatches)
	}
	_, err = os.Stat(filepath.Join(dir, "c-p.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want no pact file, got %v", err)
	}
}

func TestConsumerReportsMissingRequests(t *testing.T) {
	server, _ := startConsumer(t, getBasic())
	outcome := server.Finish()
	if outcome.Err == nil {
		t.Fatal("want the test to fail")
	}
	if len(outcome.Mismatches) != 1 || outcome.Mismatches[0].Type != "missing-request" ||
		outcome.Mismatches[0].Path != "/basic" {
		t.Fatalf("want one missing-request for /basic, got %+v", outcome.Mismatches)
	}
}

func TestConsumerFinishIsRepeatable(t *testing.T) {
	server, _ := startConsumer(t, getBasic())
	first := server.Finish()
	second := server.Finish()
	if first.Err == nil || !errors.Is(second.Err, first.Err) {
		t.Fatalf("want the same failure twice, got %v and %v", first.Err, second.Err)
	}
}

func TestConsumerRejectsBinaryBodies(t *testing.T) {
	in := putBasic()
	in.Request.Body = &model.Body{ContentType: "image/jpeg", Contents: []byte{0xff, 0xd8, 0x00, 0x10}}

	_, err := consumerAdapter{}.Start(adapter.ConsumerPact{Consumer: "c", Provider: "p", PactDir: t.TempDir()},
		[]model.Interaction{in})
	if !errors.Is(err, adapter.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go -C compatibility-suite test ./adapter/v2/`
Expected: FAIL to build: package `v2` has no non-test files, so `consumerAdapter`, `parseMismatches` and others are undefined.

- [ ] **Step 3: Implement**

Replace `compatibility-suite/adapter/adapter.go` with:
```go
// Package adapter is the seam between the compatibility steps and the Pact Go
// API under test. Step definitions only reach Pact Go through it.
package adapter

import (
	"errors"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// ErrUnsupported is wrapped by a step when the API under test cannot express
// the scenario at all, as opposed to expressing it and getting it wrong. Such
// scenarios are recorded as "unsupported" in the baseline.
var ErrUnsupported = errors.New("not supported by the API under test")

// Adapter is the API under test, as the steps drive it.
type Adapter struct {
	Consumer HTTPConsumer
}

// ConsumerPact names the pact a consumer test records.
type ConsumerPact struct {
	Consumer string
	Provider string
	// PactDir is where the pact file is written if the test passes.
	PactDir string
}

// HTTPConsumer runs HTTP consumer tests against a mock server.
type HTTPConsumer interface {
	// Start serves interactions from a mock server until the returned
	// MockServer is finished. It wraps ErrUnsupported when an interaction
	// cannot be expressed through the API under test.
	Start(pact ConsumerPact, interactions []model.Interaction) (MockServer, error)
}

// MockServer is a consumer test in progress.
type MockServer interface {
	// URL is the mock server's base URL.
	URL() string
	// Mismatches reports what the mock server has recorded so far.
	Mismatches() ([]model.Mismatch, error)
	// Finish ends the test as the end of a test function would, and reports
	// how it ended. Calling it again returns the same Outcome.
	Finish() Outcome
}

// Outcome is how a consumer test ended.
type Outcome struct {
	// Err is the failure the API under test reported; nil means the test
	// passed.
	Err error
	// Mismatches are the mock server's mismatches when the test ended.
	Mismatches []model.Mismatch
}
```

`compatibility-suite/adapter/v2/v2.go`:
```go
// Package v2 runs the compatibility suite against pact-go v2. It drives
// scenarios only through pact-go's public API, and uses internal/native
// solely to read back results that the public API does not return as data.
package v2

import "github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"

// New returns the adapter for pact-go v2.
func New() adapter.Adapter {
	return adapter.Adapter{Consumer: consumerAdapter{}}
}
```

`compatibility-suite/adapter/v2/sandbox.go`:
```go
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
```

`compatibility-suite/adapter/v2/consumer.go`:
```go
package v2

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

type consumerAdapter struct{}

// Start builds a V2 pact, the oldest specification pact-go supports, and
// runs ExecuteTest in the background until the mock server is finished.
func (consumerAdapter) Start(pact adapter.ConsumerPact, interactions []model.Interaction) (adapter.MockServer, error) {
	for _, in := range interactions {
		err := expressible(in)
		if err != nil {
			return nil, err
		}
	}

	mock, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: pact.Consumer,
		Provider: pact.Provider,
		PactDir:  pact.PactDir,
		LogDir:   pact.PactDir,
	})
	if err != nil {
		return nil, fmt.Errorf("creating the pact: %w", err)
	}
	for _, in := range interactions {
		addInteraction(mock, in)
	}

	server := &mockServer{
		ready:    make(chan consumer.MockServerConfig, 1),
		done:     make(chan struct{}),
		finished: make(chan struct{}),
	}
	go server.run(mock)

	select {
	case server.config = <-server.ready:
		return server, nil
	case <-server.finished:
		return nil, fmt.Errorf("the mock server did not start: %w", server.outcome.Err)
	}
}

// expressible reports an interaction pact-go's public API cannot express.
func expressible(in model.Interaction) error {
	for _, body := range []*model.Body{in.Request.Body, in.Response.Body} {
		if body != nil && bytes.IndexByte(body.Contents, 0) >= 0 {
			return fmt.Errorf("interaction %d has a binary %s body, which pact-go can set only as "+
				"application/octet-stream (BinaryBody) or cut short at its first NUL byte (Body): %w",
				in.Number, body.ContentType, adapter.ErrUnsupported)
		}
	}

	return nil
}

func addInteraction(mock *consumer.V2HTTPMockProvider, in model.Interaction) {
	interaction := mock.AddInteraction()
	for _, state := range in.ProviderStates {
		interaction.Given(state)
	}

	interaction.UponReceiving(in.Description).
		WithRequest(consumer.Method(in.Request.Method), in.Request.Path, func(b *consumer.V2RequestBuilder) {
			for name, values := range in.Request.Query {
				b.Query(name, exactly(values)...)
			}
			for name, values := range in.Request.Headers {
				b.Header(name, exactly(values)...)
			}
			if in.Request.Body != nil {
				b.Body(in.Request.Body.ContentType, in.Request.Body.Contents)
			}
		}).
		WillRespondWith(in.Response.Status, func(b *consumer.V2ResponseBuilder) {
			for name, values := range in.Response.Headers {
				b.Header(name, exactly(values)...)
			}
			if in.Response.Body != nil {
				b.Body(in.Response.Body.ContentType, in.Response.Body.Contents)
			}
		})
}

// exactly wraps values as exact-match values for pact-go's builders.
func exactly(values []string) []matchers.Matcher {
	out := make([]matchers.Matcher, len(values))
	for i, v := range values {
		out[i] = matchers.String(v)
	}

	return out
}

// mockServer runs pact-go's ExecuteTest in a goroutine and holds its test
// function open until Finish, so steps can make requests in between.
type mockServer struct {
	config   consumer.MockServerConfig
	ready    chan consumer.MockServerConfig
	done     chan struct{}
	finished chan struct{}
	stop     sync.Once
	outcome  adapter.Outcome
}

func (s *mockServer) run(mock *consumer.V2HTTPMockProvider) {
	defer close(s.finished)

	var (
		mismatches []model.Mismatch
		readErr    error
		testErr    error
	)
	sandbox("consumer", func(t *testing.T) {
		testErr = mock.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			s.ready <- config
			<-s.done
			// Read before the test function returns: pact-go shuts the mock
			// server down afterwards.
			mismatches, readErr = readMismatches(config.Port)

			return nil
		})
	})

	s.outcome = adapter.Outcome{Err: testErr, Mismatches: mismatches}
	if readErr != nil {
		s.outcome.Err = errors.Join(testErr, readErr)
	}
}

func (s *mockServer) URL() string {
	return fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
}

func (s *mockServer) Mismatches() ([]model.Mismatch, error) {
	select {
	case <-s.finished:
		return s.outcome.Mismatches, nil
	default:
		return readMismatches(s.config.Port)
	}
}

func (s *mockServer) Finish() adapter.Outcome {
	s.stop.Do(func() { close(s.done) })
	<-s.finished

	return s.outcome
}
```

`compatibility-suite/adapter/v2/mismatches.go`:
```go
package v2

import (
	"cmp"
	"encoding/json"
	"fmt"

	"github.com/pact-foundation/pact-go/v2/internal/native"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// readMismatches reads the mock server's mismatches through internal/native:
// pact-go returns them only as text, and its own parse drops the body path.
func readMismatches(port int) ([]model.Mismatch, error) {
	raw := native.MockServerMismatchesJSON(port)
	if raw == "" {
		return nil, nil
	}

	return parseMismatches([]byte(raw))
}

type ffiMismatch struct {
	Type    string `json:"type"`
	Method  string `json:"method"`
	Path    string `json:"path"`
	Request struct {
		Query json.RawMessage `json:"query"`
	} `json:"request"`
	Mismatches []struct {
		Type      string `json:"type"`
		Parameter string `json:"parameter"`
		Key       string `json:"key"`
		Path      string `json:"path"`
		Mismatch  string `json:"mismatch"`
	} `json:"mismatches"`
}

// parseMismatches converts pact_ffi's mismatch report to the suite's model.
func parseMismatches(data []byte) ([]model.Mismatch, error) {
	var raw []ffiMismatch

	err := json.Unmarshal(data, &raw)
	if err != nil {
		return nil, fmt.Errorf("parsing mock server mismatches %s: %w", data, err)
	}

	out := make([]model.Mismatch, 0, len(raw))
	for _, r := range raw {
		query, err := model.ParseQuery(r.Request.Query)
		if err != nil {
			return nil, fmt.Errorf("mismatch for %s %s: %w", r.Method, r.Path, err)
		}

		m := model.Mismatch{Type: r.Type, Method: r.Method, Path: r.Path, Query: query}
		for _, d := range r.Mismatches {
			m.Details = append(m.Details, model.MismatchDetail{
				Type:     d.Type,
				Location: cmp.Or(d.Parameter, d.Key, d.Path),
				Message:  d.Mismatch,
			})
		}
		out = append(out, m)
	}

	return out, nil
}
```

- [ ] **Step 4: Run them to verify they pass**

Run: `go -C compatibility-suite mod tidy && go -C compatibility-suite test -count=1 ./adapter/...`
Expected: PASS, all seven tests. (`go mod tidy` adds the `github.com/pact-foundation/pact-go/v2` requirement, which the `replace` resolves to `../`.)

- [ ] **Step 5: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./adapter/...)
git add compatibility-suite/adapter compatibility-suite/go.mod compatibility-suite/go.sum
git commit -m "feat(compat): drive pact-go consumer tests through the v2 adapter

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 5: Verifier adapter

**Files:**
- Modify: `compatibility-suite/adapter/adapter.go`, `compatibility-suite/adapter/v2/v2.go`
- Create: `compatibility-suite/adapter/v2/verifier.go`, `result.go`
- Test: `compatibility-suite/adapter/v2/result_test.go`, `verifier_test.go`

**Interfaces:**
- Consumes: `model.VerificationError` (Task 3); `native.ResultObserver` (Task 2); `sandbox` (Task 4).
- Produces:
  - The `Adapter.Verifier HTTPVerifier` field
  - `adapter.HTTPVerifier` with `Verify(request VerifyRequest) (VerifyResult, error)`
  - `adapter.VerifyRequest{ProviderURL string; PactFiles []string; Broker *BrokerSource; ProviderVersion string; Publish bool; StateCallback func(setup bool, state string) error; States []string; Headers http.Header}`
  - `adapter.BrokerSource{URL, Provider string}`
  - `adapter.VerifyResult{Err error; Errors []model.VerificationError}`

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/adapter/v2/result_test.go`:
```go
package v2

import (
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestParseVerifierResult(t *testing.T) {
	data := []byte(`{"result":false,"errors":[
	  {"interaction":"c_1 - Interaction 0","mismatch":{"type":"mismatches","mismatches":[
	    {"type":"StatusMismatch","expected":200,"actual":400,"mismatch":"expected 200 but was 400"},
	    {"type":"BodyMismatch","path":"$"}]}},
	  {"interaction":"c_2 - Interaction 1","mismatch":{"type":"error","message":"One or more of the setup state change handlers has failed"}}]}`)

	got, err := parseVerifierResult(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []model.VerificationError{
		{Interaction: "c_1 - Interaction 0", MismatchTypes: []string{"StatusMismatch", "BodyMismatch"}},
		{Interaction: "c_2 - Interaction 1", Message: "One or more of the setup state change handlers has failed"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %#v\nwant %#v", got, want)
	}
}

func TestParseVerifierResultWithNoErrors(t *testing.T) {
	got, err := parseVerifierResult([]byte(`{"result":true,"errors":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want no errors, got %+v", got)
	}
}

func TestParseVerifierResultRejectsMalformedJSON(t *testing.T) {
	_, err := parseVerifierResult([]byte(`{`))
	if err == nil {
		t.Fatal("want an error for malformed JSON")
	}
}
```

`compatibility-suite/adapter/v2/verifier_test.go`:
```go
package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
)

// writePact writes a V2 pact expecting GET /basic to return 200 and the basic
// JSON document, optionally in a provider state.
func writePact(t *testing.T, state string) string {
	t.Helper()
	interaction := map[string]any{
		"description": "get basic",
		"request":     map[string]any{"method": "GET", "path": "/basic"},
		"response": map[string]any{
			"status":  200,
			"headers": map[string]string{"Content-Type": "application/json"},
			"body":    map[string]string{"one": "a", "two": "b"},
		},
	}
	if state != "" {
		interaction["providerState"] = state
	}
	pact := map[string]any{
		"consumer":     map[string]string{"name": "c"},
		"provider":     map[string]string{"name": "p"},
		"interactions": []any{interaction},
		"metadata":     map[string]any{"pactSpecification": map[string]string{"version": "2.0.0"}},
	}

	data, err := json.Marshal(pact)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "c-p.json")
	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

// headerLog records the headers of each request a stub receives.
type headerLog struct {
	mu  sync.Mutex
	all []http.Header
}

func (l *headerLog) add(h http.Header) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.all = append(l.all, h.Clone())
}

func (l *headerLog) list() []http.Header {
	l.mu.Lock()
	defer l.mu.Unlock()

	return slices.Clone(l.all)
}

// stubProvider answers every request with status and the basic JSON document.
func stubProvider(t *testing.T, status int) (*httptest.Server, *headerLog) {
	t.Helper()
	log := &headerLog{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.add(r.Header)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(basicJSON)
	}))
	t.Cleanup(server.Close)

	return server, log
}

func TestVerifierReportsSuccess(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err != nil || len(result.Errors) != 0 {
		t.Fatalf("want a pass, got %v %+v", result.Err, result.Errors)
	}
}

func TestVerifierReportsMismatchTypes(t *testing.T) {
	server, _ := stubProvider(t, http.StatusBadRequest)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err == nil {
		t.Fatal("want the verification to fail")
	}
	if len(result.Errors) == 0 || !slices.Contains(result.Errors[0].MismatchTypes, "StatusMismatch") {
		t.Fatalf("want a StatusMismatch, got %+v", result.Errors)
	}
}

func TestVerifierCallsStateCallback(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	var (
		mu    sync.Mutex
		calls []string
	)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: server.URL,
		PactFiles:   []string{writePact(t, "state one")},
		States:      []string{"", "state one"},
		StateCallback: func(setup bool, state string) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, fmt.Sprintf("%t:%s", setup, state))

			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err != nil {
		t.Fatalf("want a pass, got %v %+v", result.Err, result.Errors)
	}
	if !slices.Contains(calls, "true:state one") || !slices.Contains(calls, "false:state one") {
		t.Fatalf("want setup and teardown of 'state one', got %v", calls)
	}
}

func TestVerifierAddsHeaders(t *testing.T) {
	server, log := stubProvider(t, http.StatusOK)
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: server.URL,
		PactFiles:   []string{writePact(t, "")},
		Headers:     http.Header{"A": {"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	received := log.list()
	if len(received) == 0 {
		t.Fatal("the provider received no requests")
	}
	for _, h := range received {
		if h.Get("A") != "1" {
			t.Fatalf("want header A: 1 on every request, got %v", h)
		}
	}
}

func TestVerifierReportsWhenNothingRan(t *testing.T) {
	// pact-go rejects a broker source without a provider name before it calls
	// the verifier, so no result is ever reported.
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: "http://127.0.0.1:1",
		Broker:      &adapter.BrokerSource{URL: "http://127.0.0.1:1"},
	})
	if err == nil {
		t.Fatal("want an error when the verifier never ran")
	}
}

func TestVerifierClearsObserver(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if native.ResultObserver != nil {
		t.Fatal("want ResultObserver cleared after a verification")
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go -C compatibility-suite test ./adapter/v2/`
Expected: FAIL to build: `undefined: verifierAdapter`, `undefined: parseVerifierResult`, `undefined: adapter.VerifyRequest`.

- [ ] **Step 3: Implement**

In `compatibility-suite/adapter/adapter.go`:
- Add `"net/http"` to the imports.
- Add `Verifier HTTPVerifier` to `Adapter`, which becomes:
```go
// Adapter is the API under test, as the steps drive it.
type Adapter struct {
	Consumer HTTPConsumer
	Verifier HTTPVerifier
}
```
- Append:
```go
// HTTPVerifier verifies HTTP providers.
type HTTPVerifier interface {
	// Verify runs one verification. The error is for a verification that
	// could not be run or read back; a failing verification is a result.
	Verify(request VerifyRequest) (VerifyResult, error)
}

// VerifyRequest describes one provider verification.
type VerifyRequest struct {
	// ProviderURL is the base URL of the provider under test.
	ProviderURL string
	// PactFiles are local pact files to verify.
	PactFiles []string
	// Broker, when set, is a Pact Broker to fetch pacts from as well.
	Broker *BrokerSource
	// ProviderVersion identifies the provider to the broker.
	ProviderVersion string
	// Publish sends verification results back to the broker.
	Publish bool
	// StateCallback, when set, is called to set up and tear down each of
	// States.
	StateCallback func(setup bool, state string) error
	States        []string
	// Headers are added to every request made to the provider.
	Headers http.Header
}

// BrokerSource is a Pact Broker holding pacts for Provider.
type BrokerSource struct {
	URL      string
	Provider string
}

// VerifyResult is how a verification ended.
type VerifyResult struct {
	// Err is the failure the API under test reported; nil means the
	// verification passed.
	Err error
	// Errors are the failures the verifier reported, per interaction.
	Errors []model.VerificationError
}
```

In `compatibility-suite/adapter/v2/v2.go` the body of `New` becomes:
```go
	return adapter.Adapter{Consumer: consumerAdapter{}, Verifier: verifierAdapter{}}
```

`compatibility-suite/adapter/v2/verifier.go`:
```go
package v2

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/proxy"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
)

type verifierAdapter struct{}

// Verify runs pact-go's provider verification, and reads the verifier's own
// result through internal/native because pact-go returns only a sentinel
// error.
func (verifierAdapter) Verify(req adapter.VerifyRequest) (adapter.VerifyResult, error) {
	request := provider.VerifyRequest{
		ProviderBaseURL:            req.ProviderURL,
		PactFiles:                  req.PactFiles,
		ProviderVersion:            req.ProviderVersion,
		PublishVerificationResults: req.Publish,
	}
	if req.Broker != nil {
		request.BrokerURL = req.Broker.URL
		request.Provider = req.Broker.Provider
	}
	if req.StateCallback != nil {
		request.StateHandlers = stateHandlers(req.States, req.StateCallback)
	}
	if len(req.Headers) > 0 {
		request.RequestFilter = addHeaders(req.Headers)
	}

	var resultJSON string
	native.ResultObserver = func(result string) { resultJSON = result }
	defer func() { native.ResultObserver = nil }()

	var verifyErr error
	sandbox("provider", func(t *testing.T) {
		verifyErr = provider.NewVerifier().VerifyProvider(t, request)
	})

	if resultJSON == "" {
		return adapter.VerifyResult{}, fmt.Errorf("the verifier reported no result (pact-go returned: %v)", verifyErr)
	}

	errs, err := parseVerifierResult([]byte(resultJSON))
	if err != nil {
		return adapter.VerifyResult{}, err
	}

	return adapter.VerifyResult{Err: verifyErr, Errors: errs}, nil
}

// stateHandlers registers callback for each state. pact-go looks handlers up
// by exact state name, so an interaction without a state needs "" here.
func stateHandlers(states []string, callback func(setup bool, state string) error) models.StateHandlers {
	handlers := make(models.StateHandlers, len(states))
	for _, state := range states {
		handlers[state] = func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return nil, callback(setup, s.Name)
		}
	}

	return handlers
}

// addHeaders is a pact-go request filter adding headers to every request.
func addHeaders(headers http.Header) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for name, values := range headers {
				for _, v := range values {
					r.Header.Add(name, v)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

`compatibility-suite/adapter/v2/result.go`:
```go
package v2

import (
	"encoding/json"
	"fmt"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

type ffiVerifierResult struct {
	Errors []struct {
		Interaction string `json:"interaction"`
		Mismatch    struct {
			Type       string `json:"type"`
			Message    string `json:"message"`
			Mismatches []struct {
				Type string `json:"type"`
			} `json:"mismatches"`
		} `json:"mismatch"`
	} `json:"errors"`
}

// parseVerifierResult reads the failures out of pact_ffi's verification
// result (pactffi_verifier_json).
func parseVerifierResult(data []byte) ([]model.VerificationError, error) {
	var result ffiVerifierResult

	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing the verifier result: %w", err)
	}

	out := make([]model.VerificationError, 0, len(result.Errors))
	for _, e := range result.Errors {
		failure := model.VerificationError{Interaction: e.Interaction, Message: e.Mismatch.Message}
		for _, m := range e.Mismatch.Mismatches {
			failure.MismatchTypes = append(failure.MismatchTypes, m.Type)
		}
		out = append(out, failure)
	}

	return out, nil
}
```

- [ ] **Step 4: Run them to verify they pass**

Run: `go -C compatibility-suite test -count=1 ./adapter/...`
Expected: PASS, all tests from Tasks 4 and 5.

- [ ] **Step 5: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./adapter/...)
git add compatibility-suite/adapter
git commit -m "feat(compat): drive pact-go provider verification through the v2 adapter

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 6: Fake provider and broker

**Files:**
- Create: `compatibility-suite/fakes/provider.go`, `compatibility-suite/fakes/broker.go`
- Test: `compatibility-suite/fakes/provider_test.go`, `compatibility-suite/fakes/broker_test.go`

**Interfaces:**
- Consumes: `model.Interaction`, `model.Request`, `model.Response` (Task 3).
- Produces (used by Task 8):
  - `fakes.NewProvider(interactions []model.Interaction) *Provider`
    - `(*Provider).URL() string`
    - `(*Provider).Requests() []RecordedRequest`
    - `(*Provider).Close()`
  - `fakes.RecordedRequest{Method, Path string; Header http.Header}`
  - `fakes.NewBroker(provider string) *Broker`
    - `(*Broker).URL() string`
    - `(*Broker).AddPact(consumer string, pact []byte)`
    - `(*Broker).Publications() []Publication`
    - `(*Broker).Close()`
  - `fakes.Publication{Consumer string; Body map[string]any}`

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/fakes/provider_test.go`:
```go
package fakes_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func get(t *testing.T, target string, header http.Header) (int, http.Header, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header = header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, res.Header, string(body)
}

func basicProvider(t *testing.T) *fakes.Provider {
	t.Helper()
	p := fakes.NewProvider([]model.Interaction{
		{
			Request: model.Request{Method: http.MethodGet, Path: "/basic"},
			Response: model.Response{
				Status:  http.StatusOK,
				Headers: http.Header{"X-Test": {"Something"}},
				Body:    &model.Body{ContentType: "application/json", Contents: []byte(`{"one":"a"}`)},
			},
		},
		{
			Request:  model.Request{Method: http.MethodGet, Path: "/with_params", Query: url.Values{"a": {"1"}}},
			Response: model.Response{Status: http.StatusNoContent},
		},
	})
	t.Cleanup(p.Close)

	return p
}

func TestProviderAnswersMatchingRequests(t *testing.T) {
	p := basicProvider(t)
	status, header, body := get(t, p.URL()+"/basic", http.Header{})
	if status != http.StatusOK || header.Get("Content-Type") != "application/json" ||
		header.Get("X-Test") != "Something" || body != `{"one":"a"}` {
		t.Fatalf("got %d %v %q", status, header, body)
	}

	status, _, _ = get(t, p.URL()+"/with_params?a=1", http.Header{})
	if status != http.StatusNoContent {
		t.Fatalf("want 204 for a matching query, got %d", status)
	}
}

func TestProviderRejectsUnmatchedRequests(t *testing.T) {
	p := basicProvider(t)
	for _, target := range []string{"/other", "/with_params?a=2", "/with_params", "/basic?x=1"} {
		status, _, _ := get(t, p.URL()+target, http.Header{})
		if status != http.StatusInternalServerError {
			t.Errorf("%s: want 500, got %d", target, status)
		}
	}
}

func TestProviderRecordsRequests(t *testing.T) {
	p := basicProvider(t)
	get(t, p.URL()+"/basic", http.Header{"A": {"1"}})

	got := p.Requests()
	if len(got) != 1 || got[0].Method != http.MethodGet || got[0].Path != "/basic" || got[0].Header.Get("A") != "1" {
		t.Fatalf("got %+v", got)
	}
}
```

`compatibility-suite/fakes/broker_test.go`:
```go
package fakes_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
)

func fetchJSON(t *testing.T, method, target, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var doc map[string]any
	err = json.NewDecoder(res.Body).Decode(&doc)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, doc
}

// link returns doc._links[rel].href.
func link(t *testing.T, doc map[string]any, rel string) string {
	t.Helper()
	links, _ := doc["_links"].(map[string]any)
	l, _ := links[rel].(map[string]any)
	href, _ := l["href"].(string)
	if href == "" {
		t.Fatalf("no %s link in %v", rel, doc)
	}

	return href
}

func TestBrokerServesPactsForVerification(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	b.AddPact("c_1", []byte(`{"consumer":{"name":"c_1"},"provider":{"name":"p"},"interactions":[]}`))

	_, index := fetchJSON(t, http.MethodGet, b.URL()+"/", "")
	forVerification := strings.ReplaceAll(link(t, index, "pb:provider-pacts-for-verification"), "{provider}", "p")

	status, list := fetchJSON(t, http.MethodPost, forVerification, `{"consumerVersionSelectors":[]}`)
	embedded, _ := list["_embedded"].(map[string]any)
	pacts, _ := embedded["pacts"].([]any)
	if status != http.StatusOK || len(pacts) != 1 {
		t.Fatalf("want one pact, got %d %v", status, list)
	}
	first, _ := pacts[0].(map[string]any)
	self := link(t, first, "self")

	_, pact := fetchJSON(t, http.MethodGet, self, "")
	consumer, _ := pact["consumer"].(map[string]any)
	if consumer["name"] != "c_1" {
		t.Fatalf("want the c_1 pact, got %v", pact)
	}
	if !strings.HasSuffix(link(t, pact, "pb:publish-verification-results"), "/consumer/c_1/verification-results") {
		t.Fatalf("want a publish link for c_1, got %v", pact)
	}
}

func TestBrokerRecordsPublications(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	b.AddPact("c_1", []byte(`{"interactions":[]}`))

	status, _ := fetchJSON(t, http.MethodPost, b.URL()+"/pacts/provider/p/consumer/c_1/verification-results",
		`{"success":true,"providerApplicationVersion":"1.2.3"}`)
	if status != http.StatusCreated {
		t.Fatalf("want 201, got %d", status)
	}

	got := b.Publications()
	if len(got) != 1 || got[0].Consumer != "c_1" || got[0].Body["success"] != true {
		t.Fatalf("got %+v", got)
	}
}

func TestBrokerRejectsUnknownPacts(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, b.URL()+"/pacts/provider/p/consumer/nobody", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", res.StatusCode)
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go -C compatibility-suite test ./fakes/`
Expected: FAIL to build: package `fakes` has no non-test files.

- [ ] **Step 3: Implement**

`compatibility-suite/fakes/provider.go`:
```go
// Package fakes provides in-process stand-ins for the services a provider
// verification talks to.
package fakes

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// RecordedRequest is a request a fake received.
type RecordedRequest struct {
	Method string
	Path   string
	Header http.Header
}

// Provider is an HTTP provider that answers each request with the response
// of the first interaction whose method, path and query match it, and with a
// 500 otherwise, as a Pact mock server would. It records every request.
type Provider struct {
	server       *httptest.Server
	interactions []model.Interaction
	mu           sync.Mutex
	requests     []RecordedRequest
}

// NewProvider starts a provider serving interactions.
func NewProvider(interactions []model.Interaction) *Provider {
	p := &Provider{interactions: interactions}
	p.server = httptest.NewServer(http.HandlerFunc(p.serve))

	return p
}

// URL is the provider's base URL.
func (p *Provider) URL() string { return p.server.URL }

// Close stops the provider.
func (p *Provider) Close() { p.server.Close() }

// Requests returns the requests received so far.
func (p *Provider) Requests() []RecordedRequest {
	p.mu.Lock()
	defer p.mu.Unlock()

	return slices.Clone(p.requests)
}

func (p *Provider) serve(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.requests = append(p.requests, RecordedRequest{Method: r.Method, Path: r.URL.Path, Header: r.Header.Clone()})
	p.mu.Unlock()

	for _, in := range p.interactions {
		if matches(in.Request, r) {
			writeResponse(w, in.Response)

			return
		}
	}
	http.Error(w, "no interaction matches "+r.Method+" "+r.URL.RequestURI(), http.StatusInternalServerError)
}

func matches(want model.Request, r *http.Request) bool {
	if want.Method != r.Method || want.Path != r.URL.Path {
		return false
	}
	got := r.URL.Query()
	if len(want.Query) == 0 {
		return len(got) == 0
	}

	return reflect.DeepEqual(map[string][]string(want.Query), map[string][]string(got))
}

func writeResponse(w http.ResponseWriter, res model.Response) {
	for name, values := range res.Headers {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
	if res.Body != nil && res.Body.ContentType != "" && w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", res.Body.ContentType)
	}
	w.WriteHeader(res.Status)
	if res.Body != nil {
		_, _ = w.Write(res.Body.Contents)
	}
}
```

`compatibility-suite/fakes/broker.go`:
```go
package fakes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
)

// Publication is a verification result published to the Broker.
type Publication struct {
	Consumer string
	Body     map[string]any
}

// Broker is the part of a Pact Broker a provider verification uses: it
// serves pacts for one provider, and records the verification results
// published back.
type Broker struct {
	server    *httptest.Server
	provider  string
	mu        sync.Mutex
	pacts     map[string][]byte
	consumers []string
	published []Publication
}

// NewBroker starts a broker for provider.
func NewBroker(provider string) *Broker {
	b := &Broker{provider: provider, pacts: map[string][]byte{}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", b.index)
	mux.HandleFunc("POST /pacts/provider/{provider}/for-verification", b.pactsForVerification)
	mux.HandleFunc("GET /pacts/provider/{provider}/consumer/{consumer}", b.pact)
	mux.HandleFunc("POST /pacts/provider/{provider}/consumer/{consumer}/verification-results", b.publish)
	b.server = httptest.NewServer(mux)

	return b
}

// URL is the broker's base URL.
func (b *Broker) URL() string { return b.server.URL }

// Close stops the broker.
func (b *Broker) Close() { b.server.Close() }

// AddPact makes pact, a pact file between consumer and the broker's
// provider, available for verification.
func (b *Broker) AddPact(consumer string, pact []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.pacts[consumer]; !ok {
		b.consumers = append(b.consumers, consumer)
	}
	b.pacts[consumer] = pact
}

// Publications returns the verification results published so far.
func (b *Broker) Publications() []Publication {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.published)
}

func (b *Broker) pactURL(consumer string) string {
	return b.URL() + "/pacts/provider/" + b.provider + "/consumer/" + consumer
}

func (b *Broker) index(w http.ResponseWriter, _ *http.Request) {
	writeHAL(w, http.StatusOK, map[string]any{"_links": map[string]any{
		"pb:provider-pacts-for-verification": map[string]any{
			"href":      b.URL() + "/pacts/provider/{provider}/for-verification",
			"templated": true,
		},
	}})
}

func (b *Broker) pactsForVerification(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("provider") != b.provider {
		http.NotFound(w, r)

		return
	}

	b.mu.Lock()
	consumers := slices.Clone(b.consumers)
	b.mu.Unlock()

	pacts := make([]map[string]any, 0, len(consumers))
	for _, consumer := range consumers {
		pacts = append(pacts, map[string]any{
			"shortDescription":       "latest",
			"verificationProperties": map[string]any{"pending": false, "notices": []any{}},
			"_links": map[string]any{"self": map[string]any{
				"href": b.pactURL(consumer),
				"name": "Pact between " + consumer + " and " + b.provider,
			}},
		})
	}
	writeHAL(w, http.StatusOK, map[string]any{
		"_embedded": map[string]any{"pacts": pacts},
		"_links": map[string]any{"self": map[string]any{
			"href": b.URL() + "/pacts/provider/" + b.provider + "/for-verification",
		}},
	})
}

func (b *Broker) pact(w http.ResponseWriter, r *http.Request) {
	consumer := r.PathValue("consumer")
	b.mu.Lock()
	raw, ok := b.pacts[consumer]
	b.mu.Unlock()
	if r.PathValue("provider") != b.provider || !ok {
		http.NotFound(w, r)

		return
	}

	var doc map[string]any
	err := json.Unmarshal(raw, &doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	doc["_links"] = map[string]any{"pb:publish-verification-results": map[string]any{
		"title": "Publish verification results",
		"href":  b.pactURL(consumer) + "/verification-results",
	}}
	writeHAL(w, http.StatusOK, doc)
}

func (b *Broker) publish(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	b.mu.Lock()
	b.published = append(b.published, Publication{Consumer: r.PathValue("consumer"), Body: body})
	b.mu.Unlock()
	writeHAL(w, http.StatusCreated, map[string]any{})
}

func writeHAL(w http.ResponseWriter, status int, doc any) {
	data, err := json.Marshal(doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/hal+json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
```

- [ ] **Step 4: Run them to verify they pass**

Run: `go -C compatibility-suite test -count=1 ./fakes/`
Expected: PASS.

- [ ] **Step 5: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./fakes/...)
git add compatibility-suite/fakes
git commit -m "feat(compat): add in-process fake provider and Pact Broker

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 7: Step helpers: tables, pact files, HTTP

**Files:**
- Create: `compatibility-suite/steps/table.go`, `steps/pactfile.go`, `steps/http.go`
- Test: `compatibility-suite/steps/table_test.go`, `steps/pactfile_test.go`, `steps/http_test.go`

**Interfaces:**
- Consumes: `model.*`, `model.ParseQuery`, `fixtures.Loader` (Task 3); `adapter.ErrUnsupported`.
- Produces (package-internal, used by Task 8):
  - `rows(table *godog.Table) [][]string`
  - `changes(rows [][]string) (map[string]string, error)`
  - `interactionsFromTable(rows [][]string, fx fixtures.Loader) ([]model.Interaction, error)`
  - `applyRequestChanges(req *model.Request, cells map[string]string, fx fixtures.Loader) error`
  - `applyResponseChanges(res *model.Response, cells map[string]string, fx fixtures.Loader, mergeHeaders bool) error`
  - `parseHeaders(value string) (http.Header, error)`
  - `parseNumbers(list string) ([]int, error)`
  - `ordinal(word string) (int, error)`
  - `pactJSON(consumer, provider string, interactions []model.Interaction) ([]byte, error)`
  - `pactFile{Interactions []pactInteraction}`, `pactInteraction{Request, Response pactPart}`
  - `pactPart` with `header(name string) (string, bool)` and `query() (url.Values, error)`
  - `sendRequest(baseURL string, req model.Request) (*http.Response, []byte, error)`
  - `jsonEqual(got, want []byte) error`

- [ ] **Step 1: Write the failing tests**

`compatibility-suite/steps/table_test.go`:
```go
package steps

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

var fx = fixtures.Loader{Dir: "../pact-compatibility-suite/fixtures"}

func TestInteractionsFromConsumerTable(t *testing.T) {
	got, err := interactionsFromTable([][]string{
		{"No", "method", "path", "query", "headers", "body", "response", "response content", "response body"},
		{"1", "GET", "/basic", "", "", "", "200", "application/json", "file: basic.json"},
		{"2", "GET", "/with_params", "a=1&b=2", "", "", "200", "", ""},
		{"3", "GET", "/with_headers", "", "'X-TEST: Compatibility'", "", "200", "", ""},
		{"8", "PUT", "/form", "", "", "file: form-post-body.xml", "200", "", ""},
	}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("want 4 interactions, got %d", len(got))
	}
	first := got[0]
	if first.Number != 1 || first.Description != "Interaction 0" || first.Request.Method != "GET" ||
		first.Request.Path != "/basic" || first.Response.Status != 200 ||
		first.Response.Body == nil || first.Response.Body.ContentType != "application/json" || first.Request.Body != nil {
		t.Fatalf("interaction 1: %+v", first)
	}
	if got[1].Request.Query.Get("a") != "1" || got[1].Request.Query.Get("b") != "2" {
		t.Fatalf("interaction 2 query: %v", got[1].Request.Query)
	}
	if got[2].Request.Headers.Get("X-TEST") != "Compatibility" {
		t.Fatalf("interaction 3 headers: %v", got[2].Request.Headers)
	}
	if got[3].Request.Body.ContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("interaction 4 body: %+v", got[3].Request.Body)
	}
}

func TestInteractionsFromProviderTableReadsResponseHeaders(t *testing.T) {
	got, err := interactionsFromTable([][]string{
		{"No", "method", "path", "query", "headers", "body", "response", "response headers", "response content", "response body"},
		{"5", "GET", "/basic", "", "", "", "200", "'X-TEST: Something'", "application/json", "file: basic.json"},
	}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Response.Headers.Get("X-Test") != "Something" {
		t.Fatalf("response headers: %v", got[0].Response.Headers)
	}
}

func TestRequestChangesKeepContentTypeForText(t *testing.T) {
	req := model.Request{Body: &model.Body{ContentType: "application/x-www-form-urlencoded", Contents: []byte("a=1")}}
	err := applyRequestChanges(&req, map[string]string{"body": "a=1&b=2&c=33&d=4"}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if req.Body.ContentType != "application/x-www-form-urlencoded" || string(req.Body.Contents) != "a=1&b=2&c=33&d=4" {
		t.Fatalf("body: %+v", req.Body)
	}
}

func TestRequestChangesReplaceHeadersAndQuery(t *testing.T) {
	req := model.Request{Headers: http.Header{"X-Test": {"Compatibility"}}}
	err := applyRequestChanges(&req, map[string]string{"headers": "'X-OTHER: Something'", "query": "a=1&c=3", "method": "HEAD", "path": "/path"}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if req.Headers.Get("X-Test") != "" || req.Headers.Get("X-Other") != "Something" {
		t.Fatalf("headers: %v", req.Headers)
	}
	if req.Query.Get("c") != "3" || req.Method != "HEAD" || req.Path != "/path" {
		t.Fatalf("request: %+v", req)
	}
}

func TestRequestChangesRejectUnknownColumn(t *testing.T) {
	req := model.Request{}
	err := applyRequestChanges(&req, map[string]string{"colour": "red"}, fx)
	if err == nil || !strings.Contains(err.Error(), "colour") {
		t.Fatalf("want an error naming the column, got %v", err)
	}
}

func TestResponseChangesMergeHeaders(t *testing.T) {
	res := model.Response{Status: 200, Headers: http.Header{"X-Test": {"Something"}}}
	err := applyResponseChanges(&res, map[string]string{"response headers": "'X-TEST: Compatibility'", "response": "400"}, fx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.Headers.Values("X-Test"), []string{"Something", "Compatibility"}) || res.Status != 400 {
		t.Fatalf("response: %+v", res)
	}
}

func TestResponseChangesRejectUnknownColumn(t *testing.T) {
	res := model.Response{}
	err := applyResponseChanges(&res, map[string]string{"colour": "red"}, fx, true)
	if err == nil || !strings.Contains(err.Error(), "colour") {
		t.Fatalf("want an error naming the column, got %v", err)
	}
}

func TestChangesNeedsOneRowOfValues(t *testing.T) {
	_, err := changes([][]string{{"body"}})
	if err == nil {
		t.Fatal("want an error for a table without values")
	}
	got, err := changes([][]string{{" body "}, {" Hello Mars! "}})
	if err != nil || got["body"] != "Hello Mars!" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestParseHeaders(t *testing.T) {
	got, err := parseHeaders("'A: 1', 'B: 2'")
	if err != nil {
		t.Fatal(err)
	}
	if got.Get("A") != "1" || got.Get("B") != "2" {
		t.Fatalf("got %v", got)
	}
	_, err = parseHeaders("'no colon'")
	if err == nil {
		t.Fatal("want an error for a header without ':'")
	}
}

func TestParseNumbers(t *testing.T) {
	got, err := parseNumbers("1, 2")
	if err != nil || !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got %v %v", got, err)
	}
	_, err = parseNumbers("1, x")
	if err == nil {
		t.Fatal("want an error for a non-number")
	}
}

func TestOrdinal(t *testing.T) {
	for word, want := range map[string]int{"first": 0, "second": 1, "third": 2} {
		got, err := ordinal(word)
		if err != nil || got != want {
			t.Errorf("%s: got %d %v", word, got, err)
		}
	}
	_, err := ordinal("fourth")
	if err == nil {
		t.Fatal("want an error for an unknown ordinal")
	}
}
```

`compatibility-suite/steps/pactfile_test.go`:
```go
package steps

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestPactJSONWritesV2Interaction(t *testing.T) {
	in := model.Interaction{
		Number: 2, Description: "Interaction 1", ProviderStates: []string{"state one"},
		Request: model.Request{Method: "GET", Path: "/with_params", Query: url.Values{"a": {"1"}, "b": {"2"}}},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "application/json", Contents: []byte(`{"one": "a"}`)}},
	}
	data, err := pactJSON("c_2", "p", []model.Interaction{in})
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Consumer     struct{ Name string } `json:"consumer"`
		Provider     struct{ Name string } `json:"provider"`
		Interactions []struct {
			Description   string `json:"description"`
			ProviderState string `json:"providerState"`
			Request       struct {
				Method string `json:"method"`
				Path   string `json:"path"`
				Query  string `json:"query"`
			} `json:"request"`
			Response struct {
				Status  int               `json:"status"`
				Headers map[string]string `json:"headers"`
				Body    map[string]string `json:"body"`
			} `json:"response"`
		} `json:"interactions"`
		Metadata struct {
			PactSpecification struct{ Version string } `json:"pactSpecification"`
		} `json:"metadata"`
	}
	err = json.Unmarshal(data, &doc)
	if err != nil {
		t.Fatal(err)
	}
	got := doc.Interactions[0]
	if doc.Consumer.Name != "c_2" || doc.Provider.Name != "p" || doc.Metadata.PactSpecification.Version != "2.0.0" ||
		got.Description != "Interaction 1" || got.ProviderState != "state one" ||
		got.Request.Method != "GET" || got.Request.Path != "/with_params" || got.Request.Query != "a=1&b=2" ||
		got.Response.Status != 200 || got.Response.Headers["Content-Type"] != "application/json" || got.Response.Body["one"] != "a" {
		t.Fatalf("got %s", data)
	}
}

func TestPactJSONWritesTextBodiesAsStrings(t *testing.T) {
	in := model.Interaction{
		Request:  model.Request{Method: "GET", Path: "/form"},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "application/x-www-form-urlencoded", Contents: []byte("a=1&b=2")}},
	}
	data, err := pactJSON("c", "p", []model.Interaction{in})
	if err != nil {
		t.Fatal(err)
	}
	var pact pactFile
	err = json.Unmarshal(data, &pact)
	if err != nil {
		t.Fatal(err)
	}
	var body string
	err = json.Unmarshal(pact.Interactions[0].Response.Body, &body)
	if err != nil || body != "a=1&b=2" {
		t.Fatalf("want a string body, got %s", pact.Interactions[0].Response.Body)
	}
}

func TestPactJSONRejectsBinaryBodies(t *testing.T) {
	in := model.Interaction{
		Request:  model.Request{Method: "GET", Path: "/bin"},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "image/jpeg", Contents: []byte{0xff, 0xd8, 0x00}}},
	}
	_, err := pactJSON("c", "p", []model.Interaction{in})
	if !errors.Is(err, adapter.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

func TestPactPartHeadersAndQuery(t *testing.T) {
	var v2, v3 pactPart
	err := json.Unmarshal([]byte(`{"headers":{"x-test":"Compatibility"},"query":"a=1&b=2"}`), &v2)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(`{"headers":{"X-Test":["Compatibility"]},"query":{"a":["1"],"b":["2"]}}`), &v3)
	if err != nil {
		t.Fatal(err)
	}
	for name, part := range map[string]pactPart{"V2": v2, "V3": v3} {
		value, ok := part.header("X-TEST")
		if !ok || value != "Compatibility" {
			t.Errorf("%s header: %q %v", name, value, ok)
		}
		query, err := part.query()
		if err != nil || query.Get("a") != "1" || query.Get("b") != "2" {
			t.Errorf("%s query: %v %v", name, query, err)
		}
	}
	_, ok := v2.header("Missing")
	if ok {
		t.Fatal("want no value for a missing header")
	}
}
```

`compatibility-suite/steps/http_test.go`:
```go
package steps

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestSendRequest(t *testing.T) {
	var got *http.Request
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(r.Context())
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("X-Reply", "yes")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	}))
	t.Cleanup(server.Close)

	res, body, err := sendRequest(server.URL, model.Request{
		Method:  http.MethodPut,
		Path:    "/basic",
		Query:   url.Values{"b": {"2"}, "a": {"1"}},
		Headers: http.Header{"X-Test": {"Compatibility"}},
		Body:    &model.Body{ContentType: "application/json", Contents: []byte(`{"one":"a"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusTeapot || res.Header.Get("X-Reply") != "yes" || string(body) != "short and stout" {
		t.Fatalf("response: %d %v %q", res.StatusCode, res.Header, body)
	}
	if got.Method != http.MethodPut || got.URL.Path != "/basic" || got.URL.RawQuery != "a=1&b=2" ||
		got.Header.Get("X-Test") != "Compatibility" || got.Header.Get("Content-Type") != "application/json" ||
		string(gotBody) != `{"one":"a"}` {
		t.Fatalf("request: %s %s %v %q", got.Method, got.URL, got.Header, gotBody)
	}
}

func TestJSONEqual(t *testing.T) {
	err := jsonEqual([]byte(`{"one":"a","two":"b"}`), []byte("{\n  \"two\": \"b\",\n  \"one\": \"a\"\n}"))
	if err != nil {
		t.Fatal(err)
	}
	err = jsonEqual([]byte(`{"one":"a"}`), []byte(`{"one":"b"}`))
	if err == nil {
		t.Fatal("want an error for different documents")
	}
	err = jsonEqual([]byte(`not json`), []byte(`{}`))
	if err == nil {
		t.Fatal("want an error for a body that is not JSON")
	}
}
```

- [ ] **Step 2: Run them to verify they fail**

Run: `go -C compatibility-suite test ./steps/`
Expected: FAIL to build: package `steps` has no non-test files.

- [ ] **Step 3: Implement**

`compatibility-suite/steps/table.go`:
```go
package steps

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// rows returns a godog table's cell values.
func rows(table *godog.Table) [][]string {
	out := make([][]string, 0, len(table.Rows))
	for _, row := range table.Rows {
		cells := make([]string, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cells = append(cells, cell.Value)
		}
		out = append(out, cells)
	}

	return out
}

// changes reads a "with the following changes" table: a header row and one
// row of values, keyed by column name.
func changes(rows [][]string) (map[string]string, error) {
	if len(rows) != 2 {
		return nil, fmt.Errorf("a changes table needs a header and one row of values, got %d rows", len(rows))
	}

	out := make(map[string]string, len(rows[0]))
	for i, name := range rows[0] {
		if i < len(rows[1]) {
			out[strings.TrimSpace(name)] = strings.TrimSpace(rows[1][i])
		}
	}

	return out, nil
}

// interactionsFromTable builds the interactions of a "the following HTTP
// interactions have been defined" table, numbered from 1 in row order.
// Columns other than those read here (such as "No") are ignored, as is
// "response content": the reference implementation also takes a body's
// content type from its body cell.
func interactionsFromTable(rows [][]string, fx fixtures.Loader) ([]model.Interaction, error) {
	if len(rows) == 0 {
		return nil, errors.New("the interactions table is empty")
	}

	header := rows[0]
	out := make([]model.Interaction, 0, len(rows)-1)
	for i, row := range rows[1:] {
		cells := make(map[string]string, len(header))
		for c, name := range header {
			if c < len(row) {
				cells[strings.TrimSpace(name)] = strings.TrimSpace(row[c])
			}
		}

		in := model.Interaction{
			Number: i + 1,
			// The reference implementation numbers descriptions from 0.
			Description: fmt.Sprintf("Interaction %d", i),
			Request:     model.Request{Method: cells["method"], Path: cells["path"]},
			Response:    model.Response{Status: http.StatusOK},
		}

		err := applyRequestChanges(&in.Request, map[string]string{
			"query": cells["query"], "headers": cells["headers"], "body": cells["body"],
		}, fx)
		if err != nil {
			return nil, fmt.Errorf("interaction %d: %w", in.Number, err)
		}

		err = applyResponseChanges(&in.Response, map[string]string{
			"response": cells["response"], "response headers": cells["response headers"], "response body": cells["response body"],
		}, fx, false)
		if err != nil {
			return nil, fmt.Errorf("interaction %d: %w", in.Number, err)
		}

		out = append(out, in)
	}

	return out, nil
}

// applyRequestChanges sets the request fields named by cells; an empty
// value clears a field. A text body keeps the content type of the body it
// replaces, as the reference implementation does.
func applyRequestChanges(req *model.Request, cells map[string]string, fx fixtures.Loader) error {
	for name, value := range cells {
		var err error

		switch name {
		case "method":
			req.Method = value
		case "path":
			req.Path = value
		case "query":
			req.Query, err = parseQueryCell(value)
		case "headers":
			req.Headers, err = parseHeaders(value)
		case "body":
			req.Body, err = bodyCell(value, req.Body, fx)
		default:
			return fmt.Errorf("unsupported request change %q", name)
		}

		if err != nil {
			return fmt.Errorf("request %s: %w", name, err)
		}
	}

	return nil
}

// applyResponseChanges sets the response fields named by cells. With
// mergeHeaders, new header values are added to existing ones rather than
// replacing them, as the reference implementation's provider steps do.
func applyResponseChanges(res *model.Response, cells map[string]string, fx fixtures.Loader, mergeHeaders bool) error {
	for name, value := range cells {
		var err error

		switch name {
		case "response":
			if value != "" {
				res.Status, err = strconv.Atoi(value)
			}
		case "response headers":
			err = setResponseHeaders(res, value, mergeHeaders)
		case "response body":
			res.Body, err = bodyCell(value, res.Body, fx)
		default:
			return fmt.Errorf("unsupported response change %q", name)
		}

		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func setResponseHeaders(res *model.Response, value string, merge bool) error {
	headers, err := parseHeaders(value)
	if err != nil {
		return err
	}

	if !merge || res.Headers == nil {
		res.Headers = headers

		return nil
	}

	for name, values := range headers {
		for _, v := range values {
			res.Headers.Add(name, v)
		}
	}

	return nil
}

// bodyCell parses a body cell, with an empty cell meaning no body.
func bodyCell(value string, current *model.Body, fx fixtures.Loader) (*model.Body, error) {
	if value == "" {
		return nil, nil //nolint:nilnil // an empty cell is a valid "no body"
	}

	fallback := ""
	if current != nil {
		fallback = current.ContentType
	}

	body, err := fx.Body(value, fallback)
	if err != nil {
		return nil, fmt.Errorf("parsing body: %w", err)
	}

	return body, nil
}

func parseQueryCell(value string) (url.Values, error) {
	if value == "" {
		return nil, nil
	}

	query, err := url.ParseQuery(value)
	if err != nil {
		return nil, fmt.Errorf("parsing query %q: %w", value, err)
	}

	return query, nil
}

// parseHeaders reads a headers cell such as 'X-TEST: Compatibility', with
// several headers separated by commas.
func parseHeaders(value string) (http.Header, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	headers := http.Header{}
	for _, part := range strings.Split(value, ",") {
		part = strings.Trim(strings.TrimSpace(part), "'")
		name, v, ok := strings.Cut(part, ":")
		if !ok {
			return nil, fmt.Errorf("header %q has no ':'", part)
		}
		headers.Add(strings.TrimSpace(name), strings.TrimSpace(v))
	}

	return headers, nil
}

// parseNumbers reads a list of interaction numbers such as "1, 2".
func parseNumbers(list string) ([]int, error) {
	parts := strings.Split(list, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("interaction list %q: %w", list, err)
		}
		out = append(out, n)
	}

	return out, nil
}

// ordinal turns "first", "second" or "third" into an index.
func ordinal(word string) (int, error) {
	index := map[string]int{"first": 0, "second": 1, "third": 2}
	i, ok := index[word]
	if !ok {
		return 0, fmt.Errorf("unknown ordinal %q", word)
	}

	return i, nil
}
```

`compatibility-suite/steps/pactfile.go`:
```go
package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// pactJSON renders interactions as a pact file between consumer and
// provider, in the V2 format: the oldest specification pact-go supports.
func pactJSON(consumer, provider string, interactions []model.Interaction) ([]byte, error) {
	list := make([]map[string]any, 0, len(interactions))
	for _, in := range interactions {
		request := map[string]any{"method": in.Request.Method, "path": in.Request.Path}
		if len(in.Request.Query) > 0 {
			request["query"] = in.Request.Query.Encode()
		}
		err := addPactBody(request, in.Request.Headers, in.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("interaction %d request: %w", in.Number, err)
		}

		response := map[string]any{"status": in.Response.Status}
		err = addPactBody(response, in.Response.Headers, in.Response.Body)
		if err != nil {
			return nil, fmt.Errorf("interaction %d response: %w", in.Number, err)
		}

		entry := map[string]any{"description": in.Description, "request": request, "response": response}
		if len(in.ProviderStates) > 0 {
			entry["providerState"] = in.ProviderStates[0]
		}
		list = append(list, entry)
	}

	data, err := json.MarshalIndent(map[string]any{
		"consumer":     map[string]string{"name": consumer},
		"provider":     map[string]string{"name": provider},
		"interactions": list,
		"metadata":     map[string]any{"pactSpecification": map[string]string{"version": "2.0.0"}},
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding the pact: %w", err)
	}

	return data, nil
}

// addPactBody adds headers and a body to a V2 request or response.
func addPactBody(part map[string]any, headers http.Header, body *model.Body) error {
	h := make(map[string]string, len(headers)+1)
	for name, values := range headers {
		h[name] = strings.Join(values, ", ")
	}

	if body != nil {
		if _, ok := h["Content-Type"]; !ok && body.ContentType != "" {
			h["Content-Type"] = body.ContentType
		}
		value, err := pactBody(*body)
		if err != nil {
			return err
		}
		part["body"] = value
	}

	if len(h) > 0 {
		part["headers"] = h
	}

	return nil
}

// pactBody is a body as V2 holds it: JSON documents inline, anything else
// as a string. V2 has no encoding for binary bodies.
func pactBody(b model.Body) (any, error) {
	if isJSON(b.ContentType) && json.Valid(b.Contents) {
		return json.RawMessage(b.Contents), nil
	}

	if !utf8.Valid(b.Contents) || bytes.IndexByte(b.Contents, 0) >= 0 {
		return nil, fmt.Errorf("a binary %s body cannot be written to a V2 pact file: %w", b.ContentType, adapter.ErrUnsupported)
	}

	return string(b.Contents), nil
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)

	return err == nil && (mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"))
}

// pactFile is the part of a written pact file the steps check.
type pactFile struct {
	Interactions []pactInteraction `json:"interactions"`
}

type pactInteraction struct {
	Request  pactPart `json:"request"`
	Response pactPart `json:"response"`
}

type pactPart struct {
	Method  string                     `json:"method"`
	Query   json.RawMessage            `json:"query"`
	Headers map[string]json.RawMessage `json:"headers"`
	Body    json.RawMessage            `json:"body"`
}

// header returns a header's value, matched case-insensitively. V2 pact
// files hold header values as strings, later versions as lists.
func (p pactPart) header(name string) (string, bool) {
	for key, raw := range p.Headers {
		if !strings.EqualFold(key, name) {
			continue
		}

		var text string
		if json.Unmarshal(raw, &text) == nil {
			return text, true
		}

		var list []string
		if json.Unmarshal(raw, &list) == nil {
			return strings.Join(list, ", "), true
		}
	}

	return "", false
}

// query returns the part's query parameters.
func (p pactPart) query() (url.Values, error) {
	query, err := model.ParseQuery(p.Query)
	if err != nil {
		return nil, fmt.Errorf("pact file query: %w", err)
	}

	return query, nil
}
```

`compatibility-suite/steps/http.go`:
```go
package steps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// sendRequest makes req against baseURL, returning the response with its
// body already read.
func sendRequest(baseURL string, req model.Request) (*http.Response, []byte, error) {
	target := baseURL + req.Path
	if len(req.Query) > 0 {
		target += "?" + req.Query.Encode()
	}

	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body.Contents)
	}

	httpReq, err := http.NewRequestWithContext(context.Background(), req.Method, target, body)
	if err != nil {
		return nil, nil, fmt.Errorf("building %s %s: %w", req.Method, target, err)
	}
	for name, values := range req.Headers {
		for _, v := range values {
			httpReq.Header.Add(name, v)
		}
	}
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", req.Body.ContentType)
	}

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("%s %s: %w", req.Method, target, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading the response to %s %s: %w", req.Method, target, err)
	}

	return res, data, nil
}

// jsonEqual reports whether two JSON documents are equal, ignoring layout:
// pact-go serves and writes JSON bodies compacted.
func jsonEqual(got, want []byte) error {
	var g, w any

	err := json.Unmarshal(got, &g)
	if err != nil {
		return fmt.Errorf("want JSON, got %q: %w", got, err)
	}
	err = json.Unmarshal(want, &w)
	if err != nil {
		return fmt.Errorf("the expected document is not JSON: %w", err)
	}
	if !reflect.DeepEqual(g, w) {
		return fmt.Errorf("want JSON %s, got %s", want, got)
	}

	return nil
}
```

- [ ] **Step 4: Run them to verify they pass**

Run: `go -C compatibility-suite test -count=1 ./steps/`
Expected: PASS.

- [ ] **Step 5: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./steps/...)
git add compatibility-suite/steps
git commit -m "feat(compat): add table, pact file and HTTP helpers for the steps

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```

---

### Task 8: V1 HTTP step definitions and suite wiring

**Files:**
- Create: `compatibility-suite/steps/steps.go`, `steps/consumer.go`, `steps/provider.go`
- Modify: `compatibility-suite/suite_test.go` (whole file below)

**Interfaces:**
- Consumes: everything in Tasks 3–7. `v2.New()` (Tasks 4–5).
- Produces:
  - `steps.Initialize(a adapter.Adapter, fixturesDir string) func(*godog.ScenarioContext)`
  - Test flag `-adapter` (default `v2`).

The "test" for this task is the compatibility suite itself. Before this task every V1 scenario is undefined, and so recorded as failing in the baseline. After it, scenarios that now pass are reported as `fixed`, and the binary-body ones as `misclassified`. That is the RED-to-GREEN signal.

- [ ] **Step 1: Confirm the starting point**

Run: `go -C compatibility-suite test -count=1 -run TestCompatibility .`
Expected: PASS, with all 190 scenarios undefined and every one listed in the baseline.

- [ ] **Step 2: Write the world and registration**

`compatibility-suite/steps/steps.go`:
```go
// Package steps defines the compatibility suite's step definitions. It
// drives the API under test only through an adapter.Adapter.
package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// world is the state of one scenario.
type world struct {
	adapter  adapter.Adapter
	fixtures fixtures.Loader
	dir      string

	interactions []model.Interaction

	// consumer scenarios
	mock         adapter.MockServer
	pactDir      string
	response     *http.Response
	responseBody []byte
	outcome      *adapter.Outcome
	pact         *pactFile

	// provider scenarios
	provider      *fakes.Provider
	broker        *fakes.Broker
	pactFiles     []string
	states        []string
	publish       bool
	stateCalls    *stateRecorder
	filterHeaders http.Header
	result        *adapter.VerifyResult
}

// Initialize returns the godog scenario initializer registering every step
// definition, driving the API under test through a.
func Initialize(a adapter.Adapter, fixturesDir string) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &world{adapter: a, fixtures: fixtures.Loader{Dir: fixturesDir}}

		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			dir, err := os.MkdirTemp("", "compatibility-suite-")
			if err != nil {
				return ctx, fmt.Errorf("creating the scenario directory: %w", err)
			}
			w.dir = dir

			return ctx, nil
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			w.close()

			return ctx, nil
		})

		sc.Step(`^the following HTTP interactions have been defined:$`, w.defineInteractions)
		registerConsumer(sc, w)
		registerProvider(sc, w)
	}
}

// close releases everything the scenario started, whether or not it passed.
func (w *world) close() {
	if w.mock != nil {
		w.mock.Finish()
	}
	if w.provider != nil {
		w.provider.Close()
	}
	if w.broker != nil {
		w.broker.Close()
	}
	if w.dir != "" {
		_ = os.RemoveAll(w.dir)
	}
}

func (w *world) defineInteractions(table *godog.Table) error {
	interactions, err := interactionsFromTable(rows(table), w.fixtures)
	if err != nil {
		return err
	}
	w.interactions = interactions

	return nil
}

// interaction returns a copy of the table's interaction n (1-based).
func (w *world) interaction(n int) (model.Interaction, error) {
	if n < 1 || n > len(w.interactions) {
		return model.Interaction{}, fmt.Errorf("there is no interaction %d (%d defined)", n, len(w.interactions))
	}

	return w.interactions[n-1].Clone(), nil
}
```

- [ ] **Step 3: Write the consumer steps**

`compatibility-suite/steps/consumer.go`:
```go
package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

const (
	consumerName = "compatibility-suite-c"
	providerName = "p"
)

func registerConsumer(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the mock server is started with interaction (\d+)$`, w.startMockServerWith)
	sc.Step(`^the mock server is started with interactions "([^"]*)"$`, w.startMockServerWithList)
	sc.Step(`^request (\d+) is made to the mock server$`, w.makeRequest)
	sc.Step(`^request (\d+) is made to the mock server with the following changes:$`, w.makeChangedRequest)
	sc.Step(`^a (\d+) (?:success|error) response is returned$`, w.responseStatusIs)
	sc.Step(`^the payload will contain the "([^"]*)" JSON document$`, w.payloadIsJSONDocument)
	sc.Step(`^the content type will be set as "([^"]*)"$`, w.responseContentTypeIs)
	sc.Step(`^the pact test is done$`, w.finishPactTest)
	sc.Step(`^the mock server status will be OK$`, w.mockServerOK)
	sc.Step(`^the mock server status will NOT be OK$`, w.mockServerNotOK)
	sc.Step(`^the mock server status will be mismatches$`, w.mockServerHasMismatches)
	sc.Step(`^the mock server will write out a Pact file for the interactions? when done$`, w.pactFileWritten)
	sc.Step(`^the mock server will NOT write out a Pact file for the interactions? when done$`, w.pactFileNotWritten)
	sc.Step(`^the pact file will contain \{(\d+)\} interactions?$`, w.pactInteractionCount)
	sc.Step(`^the \{(\w+)\} interaction request will be for a "([^"]*)"$`, w.pactRequestMethodIs)
	sc.Step(`^the \{(\w+)\} interaction request query parameters will be "([^"]*)"$`, w.pactRequestQueryIs)
	sc.Step(`^the \{(\w+)\} interaction request will contain the header "([^"]*)" with value "([^"]*)"$`, w.pactRequestHeaderIs)
	sc.Step(`^the \{(\w+)\} interaction request content type will be "([^"]*)"$`, w.pactRequestContentTypeIs)
	sc.Step(`^the \{(\w+)\} interaction request will contain the "([^"]*)" document$`, w.pactRequestDocumentIs)
	sc.Step(`^the \{(\w+)\} interaction response will contain the "([^"]*)" document$`, w.pactResponseDocumentIs)
	sc.Step(`^the mock server status will be an expected but not received error for interaction \{(\d+)\}$`, w.missingRequestFor)
	sc.Step(`^the mock server status will be an unexpected "([^"]*)" request received error for interaction \{(\d+)\}$`, w.unexpectedRequestFor)
	sc.Step(`^the mock server status will be an unexpected "([^"]*)" request received error for path "([^"]*)"$`, w.unexpectedRequestAt)
	sc.Step(`^the mismatches will contain a "([^"]*)" mismatch with error "([^"]*)"$`, w.mismatchWithError)
	sc.Step(`^the mismatches will contain a "([^"]*)" mismatch with path "([^"]*)" with error "([^"]*)"$`, w.mismatchAtPathWithError)
}

func (w *world) startMockServerWith(n int) error {
	return w.startMockServer([]int{n})
}

func (w *world) startMockServerWithList(list string) error {
	numbers, err := parseNumbers(list)
	if err != nil {
		return err
	}

	return w.startMockServer(numbers)
}

func (w *world) startMockServer(numbers []int) error {
	selected := make([]model.Interaction, 0, len(numbers))
	for _, n := range numbers {
		in, err := w.interaction(n)
		if err != nil {
			return err
		}
		selected = append(selected, in)
	}

	w.pactDir = filepath.Join(w.dir, "pacts")
	mock, err := w.adapter.Consumer.Start(adapter.ConsumerPact{
		Consumer: consumerName, Provider: providerName, PactDir: w.pactDir,
	}, selected)
	if err != nil {
		return fmt.Errorf("starting the mock server: %w", err)
	}
	w.mock = mock

	return nil
}

func (w *world) makeRequest(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.send(in.Request)
}

func (w *world) makeChangedRequest(n int, table *godog.Table) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	err = applyRequestChanges(&in.Request, cells, w.fixtures)
	if err != nil {
		return err
	}

	return w.send(in.Request)
}

func (w *world) send(req model.Request) error {
	if w.mock == nil {
		return errors.New("the mock server has not been started")
	}
	res, body, err := sendRequest(w.mock.URL(), req)
	if err != nil {
		return err
	}
	w.response, w.responseBody = res, body

	return nil
}

func (w *world) responseStatusIs(status int) error {
	if w.response == nil {
		return errors.New("no request has been made")
	}
	if w.response.StatusCode != status {
		return fmt.Errorf("want status %d, got %d: %s", status, w.response.StatusCode, w.responseBody)
	}

	return nil
}

func (w *world) payloadIsJSONDocument(name string) error {
	want, err := w.fixtures.Read(name + ".json")
	if err != nil {
		return err
	}

	return jsonEqual(w.responseBody, want)
}

func (w *world) responseContentTypeIs(contentType string) error {
	if w.response == nil {
		return errors.New("no request has been made")
	}
	got := w.response.Header.Get("Content-Type")
	if got != contentType {
		return fmt.Errorf("want content type %q, got %q", contentType, got)
	}

	return nil
}

func (w *world) finishPactTest() error {
	if w.mock == nil {
		return errors.New("the mock server has not been started")
	}
	outcome := w.mock.Finish()
	w.outcome = &outcome

	return nil
}

func (w *world) finished() (adapter.Outcome, error) {
	if w.outcome == nil {
		return adapter.Outcome{}, errors.New("the pact test is not done")
	}

	return *w.outcome, nil
}

func (w *world) mockServerOK() error {
	outcome, err := w.finished()
	if err != nil {
		return err
	}
	if outcome.Err != nil {
		return fmt.Errorf("want the pact test to pass, but it failed: %w", outcome.Err)
	}

	return nil
}

func (w *world) mockServerNotOK() error {
	outcome, err := w.finished()
	if err != nil {
		return err
	}
	if outcome.Err == nil {
		return errors.New("want the pact test to fail, but it passed")
	}

	return nil
}

func (w *world) mockServerHasMismatches() error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	if len(mismatches) == 0 {
		return errors.New("want mismatches, got none")
	}

	return nil
}

// mismatches are the mock server's mismatches: final once the pact test is
// done, and as recorded so far before then.
func (w *world) mismatches() ([]model.Mismatch, error) {
	if w.outcome != nil {
		return w.outcome.Mismatches, nil
	}
	if w.mock == nil {
		return nil, errors.New("the mock server has not been started")
	}

	mismatches, err := w.mock.Mismatches()
	if err != nil {
		return nil, fmt.Errorf("reading mismatches: %w", err)
	}

	return mismatches, nil
}

func (w *world) pactFilePath() string {
	return filepath.Join(w.pactDir, consumerName+"-"+providerName+".json")
}

func (w *world) pactFileWritten() error {
	_, err := w.finished()
	if err != nil {
		return err
	}

	//nolint:gosec // G304: the pact file this scenario's mock server wrote.
	data, err := os.ReadFile(w.pactFilePath())
	if err != nil {
		return fmt.Errorf("want a pact file: %w", err)
	}

	var pact pactFile
	err = json.Unmarshal(data, &pact)
	if err != nil {
		return fmt.Errorf("parsing the pact file: %w", err)
	}
	w.pact = &pact

	return nil
}

func (w *world) pactFileNotWritten() error {
	_, err := w.finished()
	if err != nil {
		return err
	}

	_, err = os.Stat(w.pactFilePath())
	if err == nil {
		return fmt.Errorf("want no pact file, but %s was written", w.pactFilePath())
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking for a pact file: %w", err)
	}

	return nil
}

func (w *world) pactInteraction(word string) (pactInteraction, error) {
	if w.pact == nil {
		return pactInteraction{}, errors.New("no pact file has been read")
	}
	i, err := ordinal(word)
	if err != nil {
		return pactInteraction{}, err
	}
	if i >= len(w.pact.Interactions) {
		return pactInteraction{}, fmt.Errorf("the pact file has %d interactions, not a %s", len(w.pact.Interactions), word)
	}

	return w.pact.Interactions[i], nil
}

func (w *world) pactInteractionCount(n int) error {
	if w.pact == nil {
		return errors.New("no pact file has been read")
	}
	if len(w.pact.Interactions) != n {
		return fmt.Errorf("want %d interactions in the pact file, got %d", n, len(w.pact.Interactions))
	}

	return nil
}

func (w *world) pactRequestMethodIs(word, method string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	if !strings.EqualFold(in.Request.Method, method) {
		return fmt.Errorf("want a %s request, got %s", method, in.Request.Method)
	}

	return nil
}

func (w *world) pactRequestQueryIs(word, query string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	got, err := in.Request.query()
	if err != nil {
		return err
	}
	want, err := parseQueryCell(query)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("want query %v, got %v", want, got)
	}

	return nil
}

func (w *world) pactRequestHeaderIs(word, name, value string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	got, ok := in.Request.header(name)
	if !ok || got != value {
		return fmt.Errorf("want header %s: %s, got %q (present: %v)", name, value, got, ok)
	}

	return nil
}

func (w *world) pactRequestContentTypeIs(word, contentType string) error {
	return w.pactRequestHeaderIs(word, "Content-Type", contentType)
}

func (w *world) pactRequestDocumentIs(word, name string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}

	return w.documentIs(in.Request.Body, name)
}

func (w *world) pactResponseDocumentIs(word, name string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}

	return w.documentIs(in.Response.Body, name)
}

// documentIs checks a pact file body against a fixture: JSON fixtures as
// JSON, anything else as text.
func (w *world) documentIs(body json.RawMessage, name string) error {
	want, err := w.fixtures.File(name, "")
	if err != nil {
		return err
	}
	if strings.HasSuffix(name, ".json") {
		return jsonEqual(body, want.Contents)
	}

	var got string
	err = json.Unmarshal(body, &got)
	if err != nil {
		return fmt.Errorf("want a text body, got %s", body)
	}
	if got != string(want.Contents) {
		return fmt.Errorf("want body %q, got %q", want.Contents, got)
	}

	return nil
}

func (w *world) missingRequestFor(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.findMismatch("missing-request", in.Request.Method, in.Request.Path, func(model.Mismatch) bool { return true })
}

func (w *world) unexpectedRequestFor(method string, n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.findMismatch("request-not-found", method, in.Request.Path, func(m model.Mismatch) bool {
		return sameQuery(m.Query, in.Request.Query)
	})
}

func (w *world) unexpectedRequestAt(method, path string) error {
	return w.findMismatch("request-not-found", method, path, func(model.Mismatch) bool { return true })
}

func (w *world) findMismatch(kind, method, path string, also func(model.Mismatch) bool) error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	for _, m := range mismatches {
		if m.Type == kind && m.Method == method && m.Path == path && also(m) {
			return nil
		}
	}

	return fmt.Errorf("want a %s mismatch for %s %s, got %+v", kind, method, path, mismatches)
}

func sameQuery(a, b map[string][]string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	return reflect.DeepEqual(a, b)
}

func (w *world) mismatchWithError(kind, message string) error {
	return w.findDetail(kind, nil, message)
}

func (w *world) mismatchAtPathWithError(kind, location, message string) error {
	return w.findDetail(kind, &location, message)
}

// findDetail looks for a request mismatch detail of kind ("query",
// "header", "body" or "body-content-type") whose message contains message,
// and, when location is set, at that query parameter, header or body path.
func (w *world) findDetail(kind string, location *string, message string) error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	for _, m := range mismatches {
		if m.Type != "request-mismatch" {
			continue
		}
		for _, d := range m.Details {
			if detailIs(d, kind) && (location == nil || d.Location == *location) && strings.Contains(d.Message, message) {
				return nil
			}
		}
	}

	return fmt.Errorf("want a %s mismatch containing %q, got %+v", kind, message, mismatches)
}

// detailIs matches mismatch types the way the reference implementation does:
// by prefix, so "body" matches BodyMismatch, with "body-content-type"
// meaning BodyTypeMismatch.
func detailIs(d model.MismatchDetail, kind string) bool {
	if kind == "body-content-type" {
		return d.Type == "BodyTypeMismatch"
	}

	return strings.HasPrefix(strings.ToLower(d.Type), kind)
}
```

- [ ] **Step 4: Write the provider steps**

`compatibility-suite/steps/provider.go`:
```go
package steps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// providerVersion is the provider version reported to the broker, as in the
// reference implementation.
const providerVersion = "1.2.3"

// stateChangeFailed is the verifier's error when a state setup call fails.
const stateChangeFailed = "One or more of the setup state change handlers has failed"

// verificationErrorKinds maps the suite's error names to the mismatch types
// that produce them, as the reference implementation does.
var verificationErrorKinds = map[string]string{
	"Response status did not match": "StatusMismatch",
	"Headers had differences":       "HeaderMismatch",
	"Body had differences":          "BodyMismatch",
}

func registerProvider(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^a provider is started that returns the response from interaction (\d+)$`, w.startProviderWith)
	sc.Step(`^a provider is started that returns the responses from interactions "([^"]*)"$`, w.startProviderWithList)
	sc.Step(`^a provider is started that returns the response from interaction (\d+), with the following changes:$`, w.startChangedProvider)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified$`, w.verifyPactFile)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified with a provider state "([^"]*)" defined$`, w.verifyPactFileInState)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified from a Pact broker$`, w.verifyPactFromBroker)
	sc.Step(`^publishing of verification results is enabled$`, w.enablePublishing)
	sc.Step(`^a provider state callback is configured$`, w.configureStateCallback)
	sc.Step(`^a provider state callback is configured, but will return a failure$`, w.configureFailingStateCallback)
	sc.Step(`^a request filter is configured to make the following changes:$`, w.configureRequestFilter)
	sc.Step(`^the verification is run$`, w.runVerification)
	sc.Step(`^the verification will be successful$`, w.verificationPassed)
	sc.Step(`^the verification will NOT be successful$`, w.verificationFailed)
	sc.Step(`^the verification results will contain a "([^"]*)" error$`, w.verificationErrorIs)
	sc.Step(`^a verification result will NOT be published back$`, w.nothingPublished)
	sc.Step(`^a successful verification result will be published back for interaction \{(\d+)\}$`, w.publishedSuccess)
	sc.Step(`^a failed verification result will be published back for the interaction \{(\d+)\}$`, w.publishedFailure)
	sc.Step(`^the provider state callback will be called before the verification is run$`, w.stateSetupCalled)
	sc.Step(`^the provider state callback will receive a setup call with "([^"]*)" as the provider state parameter$`, w.stateSetupCalledFor)
	sc.Step(`^the provider state callback will be called after the verification is run$`, w.stateTeardownCalled)
	sc.Step(`^the provider state callback will receive a teardown call "([^"]*)" as the provider state parameter$`, w.stateTeardownCalledFor)
	sc.Step(`^the provider state callback will NOT receive a teardown call$`, w.stateTeardownNotCalled)
	sc.Step(`^a warning will be displayed that there was no provider state callback configured for provider state "([^"]*)"$`, w.noStateCallbackWarning)
	sc.Step(`^the request to the provider will contain the header "([^"]*)"$`, w.providerRequestsHaveHeader)
}

func (w *world) startProvider(interactions []model.Interaction) error {
	if w.provider != nil {
		return errors.New("a provider is already running")
	}
	w.provider = fakes.NewProvider(interactions)

	return nil
}

func (w *world) startProviderWith(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.startProvider([]model.Interaction{in})
}

func (w *world) startProviderWithList(list string) error {
	numbers, err := parseNumbers(list)
	if err != nil {
		return err
	}
	interactions := make([]model.Interaction, 0, len(numbers))
	for _, n := range numbers {
		in, err := w.interaction(n)
		if err != nil {
			return err
		}
		interactions = append(interactions, in)
	}

	return w.startProvider(interactions)
}

func (w *world) startChangedProvider(n int, table *godog.Table) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	err = applyResponseChanges(&in.Response, cells, w.fixtures, true)
	if err != nil {
		return err
	}

	return w.startProvider([]model.Interaction{in})
}

// pactFor renders the pact verifying interaction n, between consumer c_n and
// the provider, optionally in a provider state.
func (w *world) pactFor(n int, state string) ([]byte, error) {
	in, err := w.interaction(n)
	if err != nil {
		return nil, err
	}
	if state != "" {
		in.ProviderStates = []string{state}
	}
	w.states = append(w.states, in.ProviderStates...)

	return pactJSON(fmt.Sprintf("c_%d", n), providerName, []model.Interaction{in})
}

func (w *world) verifyPactFile(n int) error {
	return w.addPactFile(n, "")
}

func (w *world) verifyPactFileInState(n int, state string) error {
	return w.addPactFile(n, state)
}

func (w *world) addPactFile(n int, state string) error {
	pact, err := w.pactFor(n, state)
	if err != nil {
		return err
	}
	path := filepath.Join(w.dir, fmt.Sprintf("c_%d-%s.json", n, providerName))
	err = os.WriteFile(path, pact, 0o600)
	if err != nil {
		return fmt.Errorf("writing the pact file: %w", err)
	}
	w.pactFiles = append(w.pactFiles, path)

	return nil
}

func (w *world) verifyPactFromBroker(n int) error {
	pact, err := w.pactFor(n, "")
	if err != nil {
		return err
	}
	if w.broker == nil {
		w.broker = fakes.NewBroker(providerName)
	}
	w.broker.AddPact(fmt.Sprintf("c_%d", n), pact)

	return nil
}

func (w *world) enablePublishing() error {
	w.publish = true

	return nil
}

func (w *world) configureStateCallback() error {
	w.stateCalls = &stateRecorder{}

	return nil
}

func (w *world) configureFailingStateCallback() error {
	w.stateCalls = &stateRecorder{fail: true}

	return nil
}

func (w *world) configureRequestFilter(table *godog.Table) error {
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	for name, value := range cells {
		if name != "headers" {
			return fmt.Errorf("unsupported request filter change %q", name)
		}
		w.filterHeaders, err = parseHeaders(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *world) runVerification() error {
	if w.provider == nil {
		return errors.New("no provider has been started")
	}

	req := adapter.VerifyRequest{
		ProviderURL:     w.provider.URL(),
		PactFiles:       w.pactFiles,
		ProviderVersion: providerVersion,
		Publish:         w.publish,
		Headers:         w.filterHeaders,
	}
	if w.broker != nil {
		req.Broker = &adapter.BrokerSource{URL: w.broker.URL(), Provider: providerName}
	}
	if w.stateCalls != nil {
		// Interactions without a provider state are set up as state "".
		req.States = append([]string{""}, w.states...)
		req.StateCallback = w.stateCalls.call
	}

	result, err := w.adapter.Verifier.Verify(req)
	if err != nil {
		return fmt.Errorf("running the verification: %w", err)
	}
	w.result = &result

	return nil
}

func (w *world) verified() (adapter.VerifyResult, error) {
	if w.result == nil {
		return adapter.VerifyResult{}, errors.New("the verification has not been run")
	}

	return *w.result, nil
}

func (w *world) verificationPassed() error {
	result, err := w.verified()
	if err != nil {
		return err
	}
	if result.Err != nil {
		return fmt.Errorf("want the verification to pass, got %w: %+v", result.Err, result.Errors)
	}

	return nil
}

func (w *world) verificationFailed() error {
	result, err := w.verified()
	if err != nil {
		return err
	}
	if result.Err == nil {
		return errors.New("want the verification to fail, but it passed")
	}

	return nil
}

func (w *world) verificationErrorIs(name string) error {
	result, err := w.verified()
	if err != nil {
		return err
	}

	kind, isMismatch := verificationErrorKinds[name]
	message := name
	if name == "State change request failed" {
		message = stateChangeFailed
	}
	for _, e := range result.Errors {
		if isMismatch && slices.Contains(e.MismatchTypes, kind) || !isMismatch && e.Message == message {
			return nil
		}
	}

	return fmt.Errorf("want a %q error, got %+v", name, result.Errors)
}

func (w *world) nothingPublished() error {
	if w.broker == nil {
		return nil
	}
	if published := w.broker.Publications(); len(published) > 0 {
		return fmt.Errorf("want nothing published, got %+v", published)
	}

	return nil
}

func (w *world) publishedSuccess(n int) error {
	return w.publishedResult(n, true)
}

func (w *world) publishedFailure(n int) error {
	return w.publishedResult(n, false)
}

func (w *world) publishedResult(n int, success bool) error {
	if w.broker == nil {
		return errors.New("no broker was used")
	}
	consumer := fmt.Sprintf("c_%d", n)
	published := w.broker.Publications()
	for _, p := range published {
		got, ok := p.Body["success"].(bool)
		if p.Consumer == consumer && ok && got == success {
			return nil
		}
	}

	return fmt.Errorf("want a result with success=%t published for %s, got %+v", success, consumer, published)
}

// stateCall is one call to the provider state callback.
type stateCall struct {
	setup bool
	state string
}

// stateRecorder is a provider state callback that records its calls, and
// fails every call when fail is set.
type stateRecorder struct {
	mu    sync.Mutex
	fail  bool
	calls []stateCall
}

func (r *stateRecorder) call(setup bool, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, stateCall{setup: setup, state: state})
	if r.fail {
		return errors.New("the provider state callback was configured to fail")
	}

	return nil
}

// called reports whether the callback received a matching call.
func (w *world) called(match func(stateCall) bool) (bool, []stateCall, error) {
	if w.stateCalls == nil {
		return false, nil, errors.New("no provider state callback is configured")
	}
	w.stateCalls.mu.Lock()
	defer w.stateCalls.mu.Unlock()
	calls := slices.Clone(w.stateCalls.calls)

	return slices.ContainsFunc(calls, match), calls, nil
}

func (w *world) expectCall(what string, match func(stateCall) bool) error {
	found, calls, err := w.called(match)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("want %s, got calls %+v", what, calls)
	}

	return nil
}

func (w *world) stateSetupCalled() error {
	return w.expectCall("a setup call", func(c stateCall) bool { return c.setup })
}

func (w *world) stateSetupCalledFor(state string) error {
	return w.expectCall(fmt.Sprintf("a setup call for %q", state), func(c stateCall) bool { return c.setup && c.state == state })
}

func (w *world) stateTeardownCalled() error {
	return w.expectCall("a teardown call", func(c stateCall) bool { return !c.setup })
}

func (w *world) stateTeardownCalledFor(state string) error {
	return w.expectCall(fmt.Sprintf("a teardown call for %q", state), func(c stateCall) bool { return !c.setup && c.state == state })
}

func (w *world) stateTeardownNotCalled() error {
	found, calls, err := w.called(func(c stateCall) bool { return !c.setup })
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("want no teardown call, got %+v", calls)
	}

	return nil
}

// noStateCallbackWarning checks what the reference implementations check:
// that the verification ran with a provider state and no callback. The
// warning itself goes to the verifier's log, which neither they nor pact-go
// expose.
func (w *world) noStateCallbackWarning(state string) error {
	_, err := w.verified()
	if err != nil {
		return err
	}
	if w.stateCalls != nil {
		return errors.New("a provider state callback was configured")
	}
	if !slices.Contains(w.states, state) {
		return fmt.Errorf("no pact declared the provider state %q", state)
	}

	return nil
}

func (w *world) providerRequestsHaveHeader(header string) error {
	if w.provider == nil {
		return errors.New("no provider has been started")
	}
	name, value, ok := strings.Cut(header, ":")
	if !ok {
		return fmt.Errorf("header %q has no ':'", header)
	}
	name, value = strings.TrimSpace(name), strings.TrimSpace(value)

	requests := w.provider.Requests()
	if len(requests) == 0 {
		return errors.New("the provider received no requests")
	}
	for _, r := range requests {
		if r.Header.Get(name) != value {
			return fmt.Errorf("want %s: %s on every request, got %v on %s %s", name, value, r.Header, r.Method, r.Path)
		}
	}

	return nil
}
```

- [ ] **Step 5: Wire the suite**

Replace `compatibility-suite/suite_test.go` with:
```go
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
```

- [ ] **Step 6: Run the suite and read the violations**

```bash
go -C compatibility-suite test -count=1 -run TestCompatibility . > /tmp/compat-run.log 2>&1; tail -80 /tmp/compat-run.log
grep -E '^\s+suite_test.go:[0-9]+: (fixed|misclassified|regression|stale):' /tmp/compat-run.log | sed -E 's/.*: (fixed|misclassified|regression|stale): ([^:]+ :: [^:]+).*/\1 \2/' | sort | uniq -c | sort -rn | head -60
```
Expected:
- FAIL with violations: most V1 scenarios reported as `fixed`.
- `misclassified` for the four binary-body scenarios: V1 consumer "Request with a binary body (positive case)" and "(negative case)", and V1 provider "Response with binary body (positive case)" and "(negative case)". They are listed as failing but are now unsupported.
- **No `regression` and no `stale` entries.**
- Some V2–V4 scenarios that use only these steps may also show as `fixed`. That's expected.

Steps that are still undefined leave their scenarios failing as before. A `regression` would mean a scenario that isn't in the baseline failed, which can't happen here, because all 190 are listed.

- [ ] **Step 7: Lint and commit**

```bash
(cd compatibility-suite && golangci-lint run --config ../.golangci.yml ./...)
git add compatibility-suite/steps compatibility-suite/suite_test.go
git commit -m "feat(compat): add V1 HTTP consumer and provider step definitions

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
(The suite is red at this commit, since the baseline is stale until Task 9. Task 9 follows immediately on the same branch, before anything is pushed.)

---

### Task 9: Update the baseline, triage what remains, document

**Files:**
- Modify: `compatibility-suite/baseline.yaml`
- Modify: `docs/contributing/compatibility-suite.md`

**Interfaces:**
- Consumes: the whole suite (Task 8).
- Produces: a green `mise run compat`, with every V1 entry left in `baseline.yaml` carrying a specific reason.

- [ ] **Step 1: Update the baseline**

```bash
mise run compat --update-baseline
git diff --stat compatibility-suite/baseline.yaml
grep -c '^- feature:' compatibility-suite/baseline.yaml
grep -B1 -A2 'feature: V1/' compatibility-suite/baseline.yaml
```
Expected:
- The entry count drops from 190 by at least 46: 50 V1 scenarios minus the 4 binary ones.
- The remaining V1 entries are the four binary scenarios as `status: unsupported`. Their reason is the adapter's message: "…binary image/jpeg body…" for the consumer, and "…cannot be written to a V2 pact file…" for the provider.

- [ ] **Step 2: Triage every other V1 entry still listed**

For each remaining V1 entry with `status: failing`:
1. Find its error in the run log: `grep -n -A6 '<scenario name>' /tmp/compat-run.log`.
2. Decide the cause:
   - **A defect in the suite's own code** (steps, fakes, adapter): fix it with a failing test first in that package's `_test.go`, as in Tasks 3–7. Then re-run Step 1. Expected cases: wrong field names in pact JSON, or a header lookup that's case-sensitive.
   - **Real pact-go or pact_ffi behaviour that differs from the specification:** keep the entry as `failing` and replace `not yet triaged` with one sentence naming the behaviour. For example: `reason: pact-go writes the request Content-Type only when ...`.
3. No V1 entry may be left with `reason: not yet triaged`.

Then confirm:
```bash
grep -A3 'feature: V1/' compatibility-suite/baseline.yaml | grep -c 'not yet triaged'
```
Expected: `0`.

- [ ] **Step 3: Confirm the suite is green and stable**

```bash
mise run compat
cp compatibility-suite/baseline.yaml /tmp/baseline.before.yaml
mise run compat --update-baseline
cmp /tmp/baseline.before.yaml compatibility-suite/baseline.yaml && echo STABLE
```
Expected: PASS, then `STABLE`. A non-stable result means a flaky scenario; investigate it before going on.

- [ ] **Step 4: Document how scenarios are driven**

In `docs/contributing/compatibility-suite.md`, insert this section before `## Updating the baseline`:
```markdown
## How scenarios run against Pact Go

Step definitions (`compatibility-suite/steps`) never call Pact Go directly.
They go through an adapter (`compatibility-suite/adapter`), selected with
`-adapter` (default `v2`).

The v2 adapter drives every scenario through pact-go's public API
(`consumer.NewV2Pact`, `provider.NewVerifier`). pact-go has no V1
specification, so V1 features run as V2 pacts. Where pact-go does not return
a result as data, the adapter reads it back through `internal/native` to
check it:

- `native.MockServerMismatchesJSON`: the mock server's full mismatch report.
- `native.ResultObserver`: the verifier's JSON result.

Something pact-go's public API cannot express, such as a binary body with its
own content type, is recorded as `unsupported`, so the baseline also lists
pact-go's public API gaps.
```

- [ ] **Step 5: Final checks**

```bash
mise run lint && mise run verify && mise run compat && mise run test
git log --oneline master..HEAD
```
Expected: all green, with Tasks 1–9 committed.

- [ ] **Step 6: Commit**

```bash
git add compatibility-suite/baseline.yaml docs/contributing/compatibility-suite.md
git commit -m "test(compat): record the V1 HTTP baseline

V1 HTTP consumer and provider scenarios now run against pact-go. Binary
body scenarios are recorded as unsupported: pact-go cannot set a binary
body with its own content type, and V2 pact files cannot carry one.

Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>"
```
