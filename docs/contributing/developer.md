# Developer documentation

## Tooling

Install [mise](https://mise.jdx.dev), then run `mise install` to get the
pinned Java and protoc versions. Go comes from your own toolchain; the
version floor is in `go.mod`.

Run `mise tasks` to list the available commands.

Docker is required for `mise run pact` and the containerised test tasks.

## Key Branches

### `1.x.x` 

The previous major version. Only bug fixes and security updates will be considered.

### `master`

The `2.x.x` release line. Current major version.