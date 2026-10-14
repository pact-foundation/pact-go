// Package v2 runs the compatibility suite against pact-go v2. It drives
// scenarios only through pact-go's public API, and uses internal/native
// solely to read back results that the public API does not return as data.
package v2

import "github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"

// New returns the adapter for pact-go v2.
func New() adapter.Adapter {
	return adapter.Adapter{Consumer: consumerAdapter{}}
}
