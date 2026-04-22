
# Releasing

Once your changes are in master, the latest release should be set as a draft at https://github.com/pact-foundation/pact-go/releases/.

Once you've tested that it works as expected:

1. Bump version in `command/version.go`.
2. Run `make release` to generate release notes and release commit.
3. Push tags `git push --follow-tags`
4. The pipeline will automatically trigger a release
   1. Edit the release notes at https://github.com/pact-foundation/pact-go/releases/edit/v<VERSION> if needed.
