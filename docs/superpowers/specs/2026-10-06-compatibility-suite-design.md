# Pact Compatibility Suite for Pact Go — Design

Date: 2026-10-06
Status: Approved design, pending spec review

## Intent

Establish a stable, repeatable behavioural baseline for Pact Go using the shared
[pact-compatibility-suite](https://github.com/pact-foundation/pact-compatibility-suite)
before making major changes to the underlying architecture.

Success means: any behaviour change introduced by a future architectural change
surfaces as a specific, named scenario that previously passed and now fails —
not as an undifferentiated "suite is red".

### What was decided (by the maintainer)

- The nature of the upcoming architectural change is **not yet decided**, so the
  suite must be decoupled from both the FFI binding layer and the public API.
- Scenarios that fail today are tracked in a **ratchet file**.
- CI runs the suite on **Linux, latest supported Go**, as a **required** check.
- Work lands as a **stack of GitHub stacked PRs**.

### Assumptions

- The upstream suite is consumed as-is; we do not edit upstream feature files.
- The existing `mise run install` task is sufficient to provision `pact_ffi`.
- No Docker or network access is required to run the suite.

## Upstream suite (pinned at `dd08122`, 2026-09-23)

20 feature files, ~186 scenarios:

| Area | Files | Scenarios |
|---|---|---|
| HTTP consumer | V1–V4 `http_consumer.feature` | 44 |
| HTTP provider | V1–V4 `http_provider.feature` | 34 |
| HTTP matching / generators | V3 `http_matching`, `http_generators` | 14 |
| Matching rules (engine-level) | V3, V4 `matching_rules.feature` | 39 |
| Generators (engine-level) | V3, V4 `generators.feature` | 16 |
| Messages | V3, V4 `message_consumer`/`message_provider` | 31 |
| V4 synchronous messages | `synchronous_message_consumer.feature` | 11 |
| V4 misc | `v4.feature` | 1 |

The suite provides Gherkin and fixtures only; step definitions are ours.
Reference implementations: pact-reference (Rust) and pact-jvm.

## Approach

A nested Go module under `compatibility-suite/`, run with
[godog](https://github.com/cucumber/godog), with all Pact Go calls behind an
adapter interface.

Rejected alternatives:

- **Same module (`internal/compat`)**: puts godog in every user's dependency
  graph.
- **Translate Gherkin to table-driven Go tests**: needs re-translation on every
  upstream change, defeating the purpose of a shared suite.

## Layout

```
compatibility-suite/
  go.mod                      # module github.com/pact-foundation/pact-go/v2/compatibility-suite
                              # replace github.com/pact-foundation/pact-go/v2 => ../
  pact-compatibility-suite/   # git subtree of upstream, squashed, pinned SHA
  baseline.yaml               # the ratchet
  suite_test.go               # godog runner + ratchet enforcement
  steps/                      # step definitions, one file per feature area
  fixtures/                   # Go helpers that load upstream fixtures (XML bodies, matcher JSON)
  fakes/                      # in-process provider stub + fake Pact Broker
  adapter/
    adapter.go                # interfaces + ErrUnsupported
    v2/                       # implementation over today's public API
```

The nested `go.mod` means the directory is excluded from the published
`pact-go/v2` module zip, and its dependencies never reach pact-go's `go.mod`.

## Adapter seam

`steps/` never imports pact-go. It depends only on `adapter`:

- `HTTPConsumer`: define interactions (V1–V4), start mock server, return
  mismatches, write pact file.
- `HTTPVerifier`: verify a provider against pact files or a broker, with
  provider-state callbacks and result publishing.
- `MessageConsumer` / `MessageVerifier`: V3/V4 async and V4 sync messages.
- `Matcher`: compare actual request/response against expected with matching
  rules (optional capability).
- `Generator`: apply generators to a request/response (optional capability).

Any method may return `adapter.ErrUnsupported`. `adapter/v2` is the only package
importing pact-go. A future architecture gets `adapter/v3` and is run against the
same baseline; the diff between the two runs is the regression report.

The adapter is selected by a test flag (`-adapter=v2`, default `v2`).

### Engine-level scenarios

`matching_rules.feature` and `generators.feature` (~55 scenarios) exercise the
matching/generator engine directly. Pact Go exposes no public API for this.

- `Matcher` in `adapter/v2` routes comparisons through a mock server where
  expressible: register the expected request, send the actual, read mismatches.
- Anything that cannot be expressed via the public API returns
  `ErrUnsupported`, recorded as `unsupported` in the baseline.

## Ratchet (`baseline.yaml`)

```yaml
# key: <feature path relative to features/> :: <scenario name>
"V1/http_consumer.feature::When all requests are made to the mock server":
  status: failing        # failing | unsupported
  reason: "..."
```

After the godog run, `suite_test.go` compares results with the file:

| Result | In baseline? | Outcome |
|---|---|---|
| pass | no | ok |
| fail | `failing` / `unsupported` | ok |
| fail | no | **CI fails**: regression |
| pass | yes | **CI fails**: "remove from baseline" (lock in the improvement) |
| — | key not in suite | **CI fails**: stale entry |

A scenario whose steps hit `ErrUnsupported` but is listed as `failing` (or vice
versa) also fails CI, so the two categories stay honest.

`go test -run TestCompatibility -update-baseline` rewrites the file. Changes are
reviewed like any other diff.

Scenario Outlines are keyed per example row (godog expands them with a
row-specific name).

## External dependencies

- **Provider under test**: in-process `httptest` server in `fakes/`, configured
  from the interactions in the scenario.
- **Pact Broker**: in-process fake serving pacts-for-verification and recording
  verification-result publishes for assertion.
- **pact_ffi**: provisioned via existing `mise run install`.

No Docker, no network.

## Error handling

- Each scenario gets a fresh scenario context (godog `ScenarioContext`), with
  mock servers, fakes and temp dirs torn down in `After` hooks, even on
  failure.
- Pact files are written to a per-scenario `t.TempDir()`.
- Step failures carry the mismatch / verifier output so a regression is
  diagnosable from the CI log alone.

## Testing & CI

- `mise run compat`: `cd compatibility-suite && go test ./... -count=1`.
- New `compatibility` job in `.github/workflows/test.yml`: `ubuntu-latest`,
  Go `1.27`, `mise run install` then `mise run compat`. Added to `complete.needs`
  so it is required.
- godog emits JUnit XML, uploaded as an artifact.
- `mise run lint` covers the nested module.
- Fixture loaders and the ratchet comparison logic get ordinary unit tests.

## Upstream sync

- Initial import: `git subtree add --prefix compatibility-suite/pact-compatibility-suite <upstream> <sha> --squash`.
- Updates: `git subtree pull ... --squash` in a dedicated PR. New upstream
  scenarios fail as "unlisted"; run `-update-baseline` to record them, then
  work them down.
- Documented in `docs/contributing/compatibility-suite.md`.

## Delivery: stacked PRs

Each step is a PR stacked on the previous one using GitHub's stacked PRs
(`gh stack`, official extension `github/gh-stack`; install with
`gh extension install github/gh-stack`). Each PR is green on its own because the
baseline moves with it.

1. **Scaffold**: nested module, subtree import, adapter interfaces, godog
   runner, ratchet, mise task, CI job, contributing doc. Every scenario is
   seeded into `baseline.yaml` as `failing`.
2. **V1 HTTP**: consumer + provider steps, provider stub, fake broker.
3. **V2 + V3 HTTP**: including `http_matching` and `http_generators`.
4. **Messages**: V3/V4 async consumer/provider, V4 sync, `v4.feature`.
5. **Engine-level**: `Matcher` / `Generator` capabilities; remaining
   inexpressible scenarios reclassified to `unsupported`.

The baseline is declared **stable** when PRs 1–4 have merged. PR 5 improves
coverage but is not a prerequisite.

## Out of scope

- Fixing Pact Go behaviour gaps the suite uncovers (tracked separately; the
  ratchet records them).
- macOS / Windows runs (can be added later as a scheduled job).
- Plugin-based scenarios (none in the pinned suite).
