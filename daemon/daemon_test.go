package daemon

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/mefellows/pact-go/utils"
)

// This guy mocks out the underlying Service provider in the Daemon,
// but executes actual Daemon code.
//
// Stubbing the exec.Cmd interface is hard, see fakeExec* functions for
// the magic.
func createMockedDaemon() (*Daemon, *ServiceMock) {
	svc := &ServiceMock{
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

	return NewDaemon(svc), svc
}

func TestNewDaemon(t *testing.T) {
	var daemon interface{}
	daemon, _ = createMockedDaemon()

	if _, ok := daemon.(*Daemon); !ok {
		t.Fatalf("must be a Daemon")
	}
}

func TestStopDaemon(t *testing.T) {
	d, _ := createMockedDaemon()
	port, _ := utils.GetFreePort()
	go d.StartDaemon(port)
	connectToDaemon(port, t)
	var res string
	d.StopDaemon("", &res)
	waitForDaemonToShutdown(port, d, t)
}

func TestShutdownDaemon(t *testing.T) {
	d, _ := createMockedDaemon()
	port, _ := utils.GetFreePort()
	go d.StartDaemon(port)
	connectToDaemon(port, t)
	d.Shutdown()
}

// Use this to wait for a daemon to be running prior
// to running tests
func connectToDaemon(port int, t *testing.T) {
	for {
		select {
		case <-time.After(1 * time.Second):
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
func waitForDaemonToShutdown(port int, daemon *Daemon, t *testing.T) {
	if daemon != nil {
		daemon.signalChan <- os.Interrupt
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

func TestStartAndStopDaemon(t *testing.T) {
	port, _ := utils.GetFreePort()
	daemon, _ := createMockedDaemon()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)
}

func TestDaemonShutdown(t *testing.T) {
	daemon, manager := createMockedDaemon()
	daemon.Shutdown()

	if manager.ServiceStopCount != 3 {
		t.Fatalf("Expected Stop() to be called 3 times but got: %d", manager.ServiceStopCount)
	}
}

func TestStartServer(t *testing.T) {
	daemon, _ := createMockedDaemon()

	req := PactMockServer{Pid: 1234}
	res := PactMockServer{}
	err := daemon.StartServer(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if res.Pid == 0 {
		t.Fatalf("Expected non-zero Pid but got: %d", res.Pid)
	}

	if res.Port != 0 {
		t.Fatalf("Expected non-zero port but got: %d", res.Port)
	}
}

func TestListServers(t *testing.T) {
	daemon, _ := createMockedDaemon()
	var res PactListResponse
	err := daemon.ListServers(PactMockServer{}, &res)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(res.Servers) != 3 {
		t.Fatalf("Expected array of len 3, got: %d", len(res.Servers))
	}
}

func TestStopServer(t *testing.T) {
	daemon, manager := createMockedDaemon()
	var cmd *exec.Cmd
	var res PactMockServer

	for _, s := range manager.List() {
		cmd = s
	}
	request := PactMockServer{
		Pid: cmd.Process.Pid,
	}

	err := daemon.StopServer(&request, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if res.Pid != cmd.Process.Pid {
		t.Fatalf("Expected PID to be 0 but got: %d", res.Pid)
	}

	if res.Status != 0 {
		t.Fatalf("Expected exit status to be 0 but got: %d", res.Status)
	}
}

func TestStopServer_Fail(t *testing.T) {
	daemon, manager := createMockedDaemon()
	var cmd *exec.Cmd
	var res PactMockServer

	for _, s := range manager.List() {
		cmd = s
	}
	request := PactMockServer{
		Pid: cmd.Process.Pid,
	}

	manager.ServiceStopError = errors.New("failed to stop server")

	err := daemon.StopServer(&request, &res)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestStopServer_FailedStatus(t *testing.T) {
	daemon, manager := createMockedDaemon()
	var cmd *exec.Cmd
	var res PactMockServer

	for _, s := range manager.List() {
		cmd = s
	}
	request := PactMockServer{
		Pid: cmd.Process.Pid,
	}

	manager.ServiceStopResult = false

	daemon.StopServer(&request, &res)

	if res.Status != 1 {
		t.Fatalf("Expected exit status to be 1 but got: %d", res.Status)
	}
}

func TestRPCClient_List(t *testing.T) {
	daemon, _ := createMockedDaemon()
	port, _ := utils.GetFreePort()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res PactListResponse
	err = client.Call("Daemon.ListServers", PactMockServer{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if len(res.Servers) != 3 {
		t.Fatalf("Expected 3 servers to be listed, got: %d", len(res.Servers))
	}
}

func TestRPCClient_StartServer(t *testing.T) {
	daemon, _ := createMockedDaemon()
	port, _ := utils.GetFreePort()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res PactMockServer
	err = client.Call("Daemon.StartServer", PactMockServer{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if res.Pid == 0 {
		t.Fatalf("Expected non-zero Pid but got: %d", res.Pid)
	}

	if res.Port != 0 {
		t.Fatalf("Expected non-zero port but got: %d", res.Port)
	}
}

func TestRPCClient_StopServer(t *testing.T) {
	daemon, manager := createMockedDaemon()
	port, _ := utils.GetFreePort()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)

	var cmd *exec.Cmd
	for _, s := range manager.List() {
		cmd = s
	}
	request := PactMockServer{
		Pid: cmd.Process.Pid,
	}

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res *PactMockServer
	err = client.Call("Daemon.StopServer", request, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if res.Pid != cmd.Process.Pid {
		t.Fatalf("Expected PID to match request %d but got: %d", cmd.Process.Pid, res.Pid)
	}

	if res.Port != 0 {
		t.Fatalf("Expected non-zero port but got: %d", res.Port)
	}
}
func TestRPCClient_StopDaemon(t *testing.T) {
	daemon, _ := createMockedDaemon()
	port, _ := utils.GetFreePort()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res string
	err = client.Call("Daemon.StopDaemon", "", &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	waitForDaemonToShutdown(port, nil, t)
}

// Adapted from http://npf.io/2015/06/testing-exec-command/
var fakeExecSuccessCommand = func() *exec.Cmd {
	return fakeExecCommand("", true, "")
}

var fakeExecFailCommand = func() *exec.Cmd {
	return fakeExecCommand("", false, "")
}

func fakeExecCommand(command string, success bool, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=%t", success)}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	<-time.After(30 * time.Second)

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		os.Exit(1)
	}

	// Success :)
	os.Exit(0)
}
