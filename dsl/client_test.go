package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/mefellows/pact-go/daemon"
	"github.com/mefellows/pact-go/utils"
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

// Use this to wait for a daemon to stop after running a test.
func waitForDaemonToShutdown(port int, t *testing.T) {
	req := ""
	res := ""
	// var req interface{}

	waitForPortInTest(port, t)

	fmt.Println("Sending remote shutdown signal...")
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))

	err = client.Call("Daemon.StopDaemon", &req, &res)
	// err = client.Call("Daemon.StopDaemon", req, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	fmt.Println(res)

	t.Logf("Waiting for deamon to shutdown before next test")
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatalf("Expected server to shutdown < 1s.")
		case <-time.After(50 * time.Millisecond):
			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			conn.SetReadDeadline(time.Now())
			defer conn.Close()
			if err != nil {
				return
			}
			buffer := make([]byte, 8)
			_, err = conn.Read(buffer)
			if err != nil {
				return
			}
		}
	}
}

// This guy mocks out the underlying Service provider in the Daemon,
// but executes actual Daemon code.
//
// Stubbing the exec.Cmd interface is hard, see fakeExec* functions for
// the magic.
func createDaemon(port int) (*daemon.Daemon, *daemon.ServiceMock) {
	svc := &daemon.ServiceMock{
		Command:           "test",
		Args:              []string{},
		ServiceStopResult: true,
		ServiceStopError:  nil,
		ExecFunc:          fakeExecSuccessCommand,
		ServiceList: map[int]*exec.Cmd{
			1: fakeExecCommand("", true, ""),
			2: fakeExecCommand("", true, ""),
			3: fakeExecCommand("", true, ""),
		},
		ServiceStartCmd: nil,
	}

	// Start all processes to get the Pids!
	for _, s := range svc.ServiceList {
		s.Start()
	}

	// Cleanup all Processes when we finish
	defer func() {
		for _, s := range svc.ServiceList {
			s.Process.Kill()
		}
	}()

	d := daemon.NewDaemon(svc)
	go d.StartDaemon(port)
	return d, svc
}

// func TestClient_Fail(t *testing.T) {
// 	client := NewPactClient{ /* don't supply port */ }
//
// }

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_List(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	waitForPortInTest(port, t)

	s := client.ListServers()

	if len(s.Servers) != 3 {
		t.Fatalf("Expected 3 server to be running, got %d", len(s.Servers))
	}
}

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_StartServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	_, svc := createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	waitForPortInTest(port, t)

	client.StartServer()
	if svc.ServiceStartCount != 1 {
		t.Fatalf("Expected 1 server to have been started, got %d", svc.ServiceStartCount)
	}
}

// Adapted from http://npf.io/2015/06/testing-exec-command/
var fakeExecSuccessCommand = func() *exec.Cmd {
	return fakeExecCommand("", true, "")
}

func fakeExecCommand(command string, success bool, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=%t", success)}
	return cmd
}
