// Package native contains the c bindings into the Pact Reference types.
package mockserver

/*
#cgo darwin,amd64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_mock_server_ffi
#cgo windows,amd64 LDFLAGS: -lpact_mock_server_ffi
#cgo linux,amd64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_mock_server_ffi
*/
import "C"
