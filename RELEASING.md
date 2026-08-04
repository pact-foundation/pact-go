# Releasing

Releasing is fully automated with [release-please](https://github.com/googleapis/release-please). There is nothing to run locally, no version to bump, and no changelog to hand-edit.

## How it works

1. Every push to `master` runs the [`release-please`](https://github.com/pact-foundation/pact-go/actions/workflows/release-please.yml) workflow. It scans commits since the last release for [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `feat!:`/`fix!:` or a `BREAKING CHANGE:` footer, `perf:`, `deps:`, etc.) and keeps a **Release PR** up to date with:
   - The next version, computed via semver from the commit types seen (`feat` → minor, `fix`/`perf`/`deps` → patch, breaking change → major).
   - A generated `CHANGELOG.md` entry.
2. To cut a release, **merge the release PR**. On merge, release-please:
   - Tags `master` with the new version (e.g. `v2.6.0`).
   - Publishes a GitHub Release.
3. The tag push triggers [`release.yml`](.github/workflows/release.yml), which runs [GoReleaser](https://goreleaser.com/) to build the release binaries and attach them to that GitHub Release.

That's the whole process. Keep merging PRs into `master` with conventional commit messages, and merge the standing release PR whenever you're ready to ship what's accumulated on it.

## Notes

- No `command/version.go` bump is needed: `pact-go version` resolves itself at runtime, either from the GoReleaser `-ldflags` (release binaries) or from the Go module version (`go install .../pact-go/v2@vX.Y.Z`).
- Release-please's config lives in [`release-please-config.json`](release-please-config.json); the version it currently believes is released is tracked in [`.release-please-manifest.json`](.release-please-manifest.json).
- If a GoReleaser run fails after a tag is already published, re-run [`release.yml`](https://github.com/pact-foundation/pact-go/actions/workflows/release.yml) manually via `workflow_dispatch`, passing the existing tag - it won't create a new tag or PR.
- Commits that aren't `feat`/`fix`/etc. (e.g. `chore:`, `docs:`, `test:`) don't trigger a version bump on their own, but will still be picked up once a `feat`/`fix` commit lands.
