package daemon

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/utils"
)

// This guy mocks out the underlying Service provider in the Daemon,
// but executes actual Daemon code.
//
// Stubbing the exec.Cmd interface is hard, see fakeExec* functions for
// the magic.
func createMockedDaemon(success bool) (*Daemon, *ServiceMock) {
	execFunc := fakeExecSuccessCommand
	if !success {
		execFunc = fakeExecFailCommand
	}
	svc := &ServiceMock{
		Command:           "test",
		Args:              []string{},
		ServiceStopResult: true,
		ServiceStopError:  nil,
		ExecFunc:          execFunc,
		ServiceList: map[int]*exec.Cmd{
			1: fakeExecCommand("", success, ""),
			2: fakeExecCommand("", success, ""),
			3: fakeExecCommand("", success, ""),
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

	return NewDaemon(svc, svc), svc
}

func TestNewDaemon(t *testing.T) {
	var daemon interface{}
	daemon, _ = createMockedDaemon(true)

	if _, ok := daemon.(*Daemon); !ok {
		t.Fatalf("must be a Daemon")
	}
}

func TestStopDaemon(t *testing.T) {
	d, _ := createMockedDaemon(true)
	port, _ := utils.GetFreePort()
	go d.StartDaemon(port)
	connectToDaemon(port, t)
	var res string
	d.StopDaemon("", &res)
	waitForDaemonToShutdown(port, d, t)
}

func TestShutdownDaemon(t *testing.T) {
	d, _ := createMockedDaemon(true)
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
	daemon, _ := createMockedDaemon(true)
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)
}

func TestDaemonShutdown(t *testing.T) {
	daemon, manager := createMockedDaemon(true)
	daemon.Shutdown()

	if manager.ServiceStopCount != 3 {
		t.Fatalf("Expected Stop() to be called 3 times but got: %d", manager.ServiceStopCount)
	}
}

func TestStartServer(t *testing.T) {
	daemon, _ := createMockedDaemon(true)

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
	daemon, _ := createMockedDaemon(true)
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
	daemon, manager := createMockedDaemon(true)
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
	daemon, manager := createMockedDaemon(true)
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
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestStopServer_FailedStatus(t *testing.T) {
	daemon, manager := createMockedDaemon(true)
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

func TestVerifyProvider_MissingProviderBaseURL(t *testing.T) {
	daemon, _ := createMockedDaemon(true)

	req := VerifyRequest{}
	res := Response{}
	err := daemon.VerifyProvider(&req, &res)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if res.ExitCode != 1 {
		t.Fatalf("Expected non-zero exit code (1) but got: %d", res.ExitCode)
	}

	if !strings.Contains(res.Message, "ProviderBaseURL is mandatory") {
		t.Fatalf("Expected error message but got '%s'", res.Message)
	}
}

func TestVerifyProvider_MissingPactURLs(t *testing.T) {
	daemon, _ := createMockedDaemon(true)

	req := VerifyRequest{
		ProviderBaseURL: "http://foo.com",
	}
	res := Response{}
	err := daemon.VerifyProvider(&req, &res)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if res.ExitCode != 1 {
		t.Fatalf("Expected non-zero exit code (1) but got: %d", res.ExitCode)
	}

	if !strings.Contains(res.Message, "PactURLs is mandatory") {
		t.Fatalf("Expected error message but got '%s'", res.Message)
	}
}

func TestVerifyProvider_Valid(t *testing.T) {
	daemon, _ := createMockedDaemon(true)

	req := VerifyRequest{
		ProviderBaseURL: "http://foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	}
	res := Response{}
	err := daemon.VerifyProvider(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestVerifyProvider_FailedCommand(t *testing.T) {
	daemon, _ := createMockedDaemon(false)

	req := VerifyRequest{
		ProviderBaseURL: "http://foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	}
	res := Response{}
	err := daemon.VerifyProvider(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if res.ExitCode != 1 {
		t.Fatalf("Expected non-zero exit code (1) but got: %d", res.ExitCode)
	}

	if !strings.Contains(res.Message, "COMMAND: oh noes!") {
		t.Fatalf("Expected error message but got '%s'", res.Message)
	}
}

func TestVerifyProvider_ValidProviderStates(t *testing.T) {
	daemon, _ := createMockedDaemon(true)

	req := VerifyRequest{
		ProviderBaseURL:        "http://foo.com",
		PactURLs:               []string{"foo.json", "bar.json"},
		BrokerUsername:         "foo",
		BrokerPassword:         "foo",
		ProviderStatesURL:      "http://foo/states",
		ProviderStatesSetupURL: "http://foo/states/setup",
	}
	res := Response{}
	err := daemon.VerifyProvider(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestRPCClient_List(t *testing.T) {
	daemon, _ := createMockedDaemon(true)
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
	daemon, _ := createMockedDaemon(true)
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
	daemon, manager := createMockedDaemon(true)
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
	daemon, _ := createMockedDaemon(true)
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

func TestRPCClient_Verify(t *testing.T) {
	daemon, _ := createMockedDaemon(true)
	port, _ := utils.GetFreePort()
	defer waitForDaemonToShutdown(port, daemon, t)
	go daemon.StartDaemon(port)
	connectToDaemon(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	req := VerifyRequest{
		ProviderBaseURL: "http://foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	}
	res := Response{}

	err = client.Call("Daemon.VerifyProvider", req, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if res.ExitCode != 0 {
		t.Fatalf("Expected exit code of zero but got: %d", res.ExitCode)
	}

	if res.Message != "COMMAND: oh yays!" {
		t.Fatalf("Expected empty message but got: '%s'", res.Message)
	}
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
	<-time.After(1 * time.Second)

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		fmt.Fprintf(os.Stdout, "COMMAND: oh noes!")
		os.Exit(1)
	}

	// Success :)
	fmt.Fprintf(os.Stdout, "COMMAND: oh yays!")
	os.Exit(0)
}
