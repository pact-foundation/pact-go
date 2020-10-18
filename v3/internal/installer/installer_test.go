// Package install contains functions necessary for installing and checking
// if the necessary underlying shared libs have been properly installed
package installer

// Requirements
type Downloader struct{}

// 1. Be able to specify the path of the binary in advance (`PACT_GO_SHARED_LIBRARY_PATH`)
// 2. Automatically copy the library path to the required one for CGO to be able to reliably link
// 2. Check if the correct versions of the libs are present???
// 3. Download the appropriate libs
// 4. Disable the check

// libpact_mock_server_ffi-linux-x86_64.a.gz
// libpact_mock_server_ffi-linux-x86_64.so.gz
// libpact_mock_server_ffi-osx-x86_64.a.gz
// libpact_mock_server_ffi-osx-x86_64.dylib.gz
// libpact_mock_server_ffi-windows-x86_64.dll.gz
// libpact_mock_server_ffi-windows-x86_64.dll.lib.gz
// libpact_mock_server_ffi-windows-x86_64.lib.gz

// https://github.com/pact-foundation/pact-reference/releases/download/libpact_mock_server_ffi-v0.0.7/libpact_mock_server_ffi-osx-x86_64.dylib.gz
