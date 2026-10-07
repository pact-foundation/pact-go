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
