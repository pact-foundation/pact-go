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
