
# Releasing

Once your changes are in master, the latest release should be set as a draft at https://github.com/pact-foundation/pact-go/releases/.

Once you've tested that it works as expected:

1. Bump version in `command/version.go`.
1. Run `make release` to generate release notes and release commit.
1. Edit the release notes at https://github.com/pact-foundation/pact-go/releases/edit/v<VERSION>.
