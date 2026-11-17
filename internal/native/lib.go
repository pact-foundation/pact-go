// Package native contains the c bindings into the Pact Reference types.
package native

/*
#cgo darwin,arm64 LDFLAGS: -L/usr/local/lib -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo darwin,amd64 LDFLAGS: -L/usr/local/lib -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo windows,amd64 LDFLAGS: -lpact_ffi
#cgo linux,amd64 LDFLAGS: -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
#cgo linux,arm64 LDFLAGS: -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/usr/local/lib -lpact_ffi
*/
import "C"
