//go:build consumer || provider

package grpc

import "os"

var dir, _ = os.Getwd()
