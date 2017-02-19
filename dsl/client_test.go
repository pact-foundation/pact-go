package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/daemon"
	"github.com/pact-foundation/pact-go/types"
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
func createDaemon(port int, success bool) (*daemon.Daemon, *daemon.ServiceMock) {
	execFunc := fakeExecSuccessCommand
	if !success {
		execFunc = fakeExecFailCommand
	}
	svc := &daemon.ServiceMock{
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

	d := daemon.NewDaemon(svc, svc)
	go d.StartDaemon(port, "tcp", "")
	return d, svc
}

func TestClient_List(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
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
	client.StartServer([]string{})
	list := client.ListServers()

	if len(list.Servers) != 0 {
		t.Fatalf("Expected 0 servers, got %d", len(list.Servers))
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StartServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	_, svc := createDaemon(port, true)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	client.StartServer([]string{})
	if svc.ServiceStartCount != 1 {
		t.Fatalf("Expected 1 server to have been started, got %d", svc.ServiceStartCount)
	}
}

func TestClient_RPCErrors(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)

	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)

	// Mock out the RPC client
	oldCommandStartServer := commandStartServer
	oldCommandStopServer := commandStopServer
	oldCommandVerifyProvider := commandVerifyProvider
	oldCommandListServers := commandListServers
	oldCommandStopDaemon := commandStopDaemon

	commandStartServer = "failcommand"
	commandStopServer = "failcommand"
	commandVerifyProvider = "failverifycommand"
	commandListServers = "failcommand"
	commandStopDaemon = "failcommand"

	defer func() {
		commandStartServer = oldCommandStartServer
		commandStopServer = oldCommandStopServer
		commandVerifyProvider = oldCommandVerifyProvider
		commandListServers = oldCommandListServers
		commandStopDaemon = oldCommandStopDaemon
	}()

	client := &PactClient{Port: port}
	testCases := map[interface{}]func() interface{}{
		"rpc: service/method request ill-formed: failcommand": func() interface{} {
			return client.StopDaemon().Error()
		},
		&types.MockServer{}: func() interface{} {
			return client.StopServer(&types.MockServer{})
		},
		&types.MockServer{}: func() interface{} {
			return client.StartServer([]string{})
		},
		&types.PactListResponse{}: func() interface{} {
			return client.ListServers()
		},
		"rpc: service/method request ill-formed: failverifycommand": func() interface{} {
			_, err := client.VerifyProvider(types.VerifyRequest{})
			return err.Error()
		},
	}

	for expected, testCase := range testCases {
		res := testCase()
		if !reflect.DeepEqual(expected, res) {
			t.Fatalf("Expected '%v' but got '%v'", expected, res)
		}
	}
}

func TestClient_getPort(t *testing.T) {
	testCases := map[string]int{
		"http://localhost:8000": 8000,
		"http://localhost":      80,
		"https://localhost":     443,
		":::::":                 -1,
	}

	for host, port := range testCases {
		if getPort(host) != port {
			t.Fatalf("Expected host '%s' to return port '%d' but got '%d'", host, port, getPort(host))
		}
	}
}

func TestClient_VerifyProvider(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	ms := setupMockServer(true, t)
	defer ms.Close()

	req := types.VerifyRequest{
		ProviderBaseURL:        ms.URL,
		PactURLs:               []string{"foo.json", "bar.json"},
		BrokerUsername:         "foo",
		BrokerPassword:         "foo",
		ProviderStatesURL:      "http://foo/states",
		ProviderStatesSetupURL: "http://foo/states/setup",
	}
	_, err := client.VerifyProvider(req)

	if err != nil {
		t.Fatal("Error: ", err)
	}
}

func TestClient_VerifyProviderFailValidation(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	req := types.VerifyRequest{}
	_, err := client.VerifyProvider(req)

	if err == nil {
		t.Fatal("Expected a error but got none")
	}

	if !strings.Contains(err.Error(), "ProviderBaseURL is mandatory") {
		t.Fatalf("Expected a proper error message but got '%s'", err.Error())
	}
}

func TestClient_VerifyProviderFailExecution(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, false)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	ms := setupMockServer(true, t)
	defer ms.Close()

	req := types.VerifyRequest{
		ProviderBaseURL: ms.URL,
		PactURLs:        []string{"foo.json", "bar.json"},
	}
	_, err := client.VerifyProvider(req)

	if err == nil {
		t.Fatal("Expected a error but got none")
	}

	if !strings.Contains(err.Error(), "COMMAND: oh noes!") {
		t.Fatalf("Expected a proper error message but got '%s'", err.Error())
	}
}

var oldTimeoutDuration = timeoutDuration

func TestClient_StartServerFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond

	client := &PactClient{ /* don't supply port */ }
	server := client.StartServer([]string{})
	if server.Port != 0 {
		t.Fatalf("Expected server to be empty %v", server)
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StopServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	_, svc := createDaemon(port, true)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{Port: port}

	client.StopServer(&types.MockServer{})
	if svc.ServiceStopCount != 1 {
		t.Fatalf("Expected 1 server to have been stopped, got %d", svc.ServiceStartCount)
	}
}

func TestClient_StopServerFail(t *testing.T) {
	timeoutDuration = 50 * time.Millisecond
	client := &PactClient{ /* don't supply port */ }
	res := client.StopServer(&types.MockServer{})
	should := &types.MockServer{}
	if !reflect.DeepEqual(res, should) {
		t.Fatalf("Expected nil object but got a difference: %v != %v", res, should)
	}
	timeoutDuration = oldTimeoutDuration
}

func TestClient_StopDaemon(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
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
var fakeExecFailCommand = func() *exec.Cmd {
	return fakeExecCommand("", false, "")
}

func fakeExecCommand(command string, success bool, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=%t", success)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	<-time.After(50 * time.Millisecond)

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		fmt.Fprintf(os.Stdout, "COMMAND: oh noes!\n")
		os.Exit(1)
	}

	// Success :)
	fmt.Fprintf(os.Stdout, "COMMAND: oh yays!\n")
	os.Exit(0)
}
