# Developer documentation

## Tooling

* Docker
* Java (>= 19) - required for the Avro plugin example

## Key Branches

### `1.x.x` 

The previous major version. Only bug fixes and security updates will be considered.

### `master`

The `2.x.x` release line. Current major version.


## Windows on ARM

WoA is not supported by Pact FFI atm.

Build errors

```ps1
PS W:\> go build -o build/pact-go.exe
error obtaining VCS status: exit status 128
        Use -buildvcs=false to disable VCS stamping.

```

Install ffi

```ps1
PS W:\> .\build\pact-go.exe -l DEBUG install --libDir /tmp
2023/10/03 19:02:08 [INFO] set lib dir target to /tmp
2023/10/03 19:02:08 [INFO] package libpact_ffi not found
2023/10/03 19:02:08 [INFO] downloading library from https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v0.4.5/pact_ffi-windows-x86_64.dll.gz to /tmp/pact_ffi.dll
&{}
2023/10/03 19:02:12 [DEBUG] obtaining hash for file /tmp/pact_ffi.dll
2023/10/03 19:02:12 [DEBUG] writing config {map[libpact_ffi:{libpact_ffi 0.4.5 f03507c43328add6e02215740c044988}]}
2023/10/03 19:02:12 [DEBUG] writing yaml config to file libraries:
  libpact_ffi:
    libname: libpact_ffi
    version: 0.4.5
    hash: f03507c43328add6e02215740c044988

2023/10/03 19:02:12 [INFO] package libpact_ffi found
2023/10/03 19:02:12 [INFO] checking version 0.4.5 of libpact_ffi against semver constraint >= 0.4.0, < 1.0.0
2023/10/03 19:02:12 [DEBUG] 0.4.5 satisfies constraints 0.4.5 >= 0.4.0, < 1.0.0
2023/10/03 19:02:12 [INFO] package libpact_ffi is correctly installed
2023/10/03 19:02:12 [DEBUG] skip checking ffi version() call because FFI not loaded. This is expected when running the 'pact-go' command.
```

Some WIP steps

```ps1
scoop install go
$env:GOARCH=amd64
go mod tidy
go build -o build/pact-go.exe -buildvcs=false
.\build\pact-go.exe -l DEBUG install --libDir /tmp
scoop install grep
go list -buildvcs=false ./... | grep -v vendor | grep -v examples
go test -v github.com/pact-foundation/pact-go/v2/installer
```