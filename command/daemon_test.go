package command

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/utils"
)

// Use this to wait for a daemon to be running prior
// to running tests
func waitForPortInTest(port int, t *testing.T) {
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatalf("Expected server to start < 1s.")
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				return
			}
		}
	}
}

func TestDaemonCommand(t *testing.T) {
	args := []string{"daemon"}
	p, _ := utils.GetFreePort()
	port = p
	network = "tcp"
	go daemonCmd.Run(nil, args)

	waitForPortInTest(port, t)
	daemonCmdInstance.Shutdown()
}
