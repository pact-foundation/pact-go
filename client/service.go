/*
Package client is an internal package, implementing the raw interface to the
Pact CLI tools: The Pact Mock Service and Provider Verification "binaries."

See https://github.com/pact-foundation/pact-provider-verifier and
https://github.com/bethesque/pact-mock_service for more on the Ruby "binaries".

NOTE: The ultimate goal here is to replace the Ruby dependencies with a shared
library (Pact Reference - (https://github.com/pact-foundation/pact-reference/).
*/
package client

import (
	"os/exec"
)

// Service is a process wrapper for 3rd party binaries. It will spawn an instance
// of the binary and manage the life-cycle and IO of the process.
type Service interface {
	Setup()
	Stop(pid int) (bool, error)
	List() map[int]*exec.Cmd
	Command() *exec.Cmd
	Start() *exec.Cmd
	NewService(args []string) Service
}
