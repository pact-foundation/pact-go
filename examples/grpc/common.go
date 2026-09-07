//go:build consumer || provider

// Package grpc contains a runnable Pact example verifying a gRPC service
// via a pact_ffi plugin.
package grpc

import "os"

var dir, _ = os.Getwd()
