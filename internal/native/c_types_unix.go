//go:build darwin || linux

package native

import "C"

// CUlong is the platform C unsigned long type, used for pact_ffi calls
// that take a size_t/C.ulong (e.g. array indices, buffer lengths).
type CUlong = C.ulong
