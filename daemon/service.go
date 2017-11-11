package daemon

import (
	"io"
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
	Run(io.Writer) (*exec.Cmd, error)
	NewService(args []string) Service
}
