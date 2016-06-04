package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"reflect"
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

	waitForPortInTest(port, t)

	t.Log("Sending remote shutdown signal...\n")
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))

	err = client.Call("Daemon.StopDaemon", &req, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

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
// but executes actual Daemon code. This means we don't spin up the real
// mock service but execute our code in isolation.
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

func TestClient_List(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	s := client.ListServers()

	if len(s.Servers) != 3 {
		t.Fatalf("Expected 3 server to be running, got %d", len(s.Servers))
	}
}

func TestClient_ListFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond
	client := &PactClient{ /* don't supply port */ }
	client.StartServer()
	list := client.ListServers()

	if len(list.Servers) != 0 {
		t.Fatalf("Expected 0 servers, got %d", len(list.Servers))
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StartServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	_, svc := createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	client.StartServer()
	if svc.ServiceStartCount != 1 {
		t.Fatalf("Expected 1 server to have been started, got %d", svc.ServiceStartCount)
	}
}

var oldTimeoutDuration = timeoutDuration

func TestClient_StartServerFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond

	client := &PactClient{ /* don't supply port */ }
	server := client.StartServer()
	if server.Port != 0 {
		t.Fatalf("Expected server to be empty %v", server)
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StopServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	_, svc := createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	client.StopServer(&daemon.PactMockServer{})
	if svc.ServiceStopCount != 1 {
		t.Fatalf("Expected 1 server to have been stopped, got %d", svc.ServiceStartCount)
	}
}

func TestClient_StopServerFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond
	client := &PactClient{ /* don't supply port */ }
	res := client.StopServer(&daemon.PactMockServer{})
	should := &daemon.PactMockServer{}
	if !reflect.DeepEqual(res, should) {
		t.Fatalf("Expected nil object but got a difference: %v != %v", res, should)
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StopDaemon(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)
	waitForPortInTest(port, t)
	client := &PactClient{Port: port}

	err := client.StopDaemon()
	if err != nil {
		t.Fatalf("Err: %v", err)
	}
	waitForDaemonToShutdown(port, t)
}

func TestClient_StopDaemonFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond
	client := &PactClient{ /* don't supply port */ }
	err := client.StopDaemon()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
	timeoutDuration = oldTimeoutDuration
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
