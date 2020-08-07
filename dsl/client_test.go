package dsl

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

func TestClient_List(t *testing.T) {
	client, _ := createMockClient(true)
	servers := client.ListServers()

	if len(servers) != 3 {
		t.Fatalf("Expected 3 server to be running, got %d", len(servers))
	}
}

func TestClient_StartServer(t *testing.T) {
	client, svc := createMockClient(true)
	defer stubPorts()()

	port, _ := utils.GetFreePort()
	client.StartServer([]string{}, port)
	if svc.ServiceStartCount != 1 {
		t.Fatalf("Expected 1 server to have been started, got %d", svc.ServiceStartCount)
	}
}

func TestClient_StartServerFail(t *testing.T) {
	client, _ := createMockClient(false)
	server := client.StartServer([]string{}, 0)
	if server.Port != 0 {
		t.Fatalf("Expected server to be empty %v", server)
	}
}

func TestClient_StopServer(t *testing.T) {
	client, svc := createMockClient(true)

	client.StopServer(&types.MockServer{})
	if svc.ServiceStopCount != 1 {
		t.Fatalf("Expected 1 server to have been stopped, got %d", svc.ServiceStartCount)
	}
}

func TestClient_StopServerFail(t *testing.T) {
	client, _ := createMockClient(true)
	res, err := client.StopServer(&types.MockServer{})
	should := &types.MockServer{}
	if !reflect.DeepEqual(res, should) {
		t.Fatalf("Expected nil object but got a difference: %v != %v", res, should)
	}
	if err != nil {
		t.Fatalf("wanted error, got none")
	}
}

func TestClient_VerifyProvider(t *testing.T) {
	client, _ := createMockClient(true)

	ms := setupMockServer(true, t)
	defer ms.Close()

	req := types.VerifyRequest{
		ProviderBaseURL:        ms.URL,
		PactURLs:               []string{"foo.json", "bar.json"},
		BrokerUsername:         "foo",
		BrokerPassword:         "foo",
		ProviderStatesSetupURL: "http://foo/states/setup",
	}
	_, err := client.VerifyProvider(req)

	if err != nil {
		t.Fatal("Error: ", err)
	}
}

func TestClient_VerifyProviderFailValidation(t *testing.T) {
	client, _ := createMockClient(true)

	req := types.VerifyRequest{}
	_, err := client.VerifyProvider(req)

	if err == nil {
		t.Fatal("Expected a error but got none")
	}

	if !strings.Contains(err.Error(), "One of 'PactURLs' or 'BrokerURL' must be specified") {
		t.Fatalf("Expected a proper error message but got '%s'", err.Error())
	}
}

func TestClient_VerifyProviderFailExecution(t *testing.T) {
	client, _ := createMockClient(false)

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

func TestClient_getAddress(t *testing.T) {
	testCases := map[string]string{
		"http://localhost:8000": "localhost",
		"http://localhost":      "localhost",
		"http://127.0.0.1":      "127.0.0.1",
		":::::":                 "",
	}

	for host, address := range testCases {
		if getAddress(host) != address {
			t.Fatalf("Expected host '%s' to return address '%s' but got '%s'", host, address, getAddress(host))
		}
	}
}

func TestClient_sanitiseRubyResponse(t *testing.T) {
	var tests = map[string]string{
		"this is a sentence with a hash # so it should be in tact":                                           "this is a sentence with a hash # so it should be in tact",
		"this is a sentence with a hash and newline\n#so it should not be in tact":                           "this is a sentence with a hash and newline",
		"this is a sentence with a ruby statement bundle exec rake pact:verify so it should not be in tact":  "",
		"this is a sentence with a ruby statement\nbundle exec rake pact:verify so it should not be in tact": "this is a sentence with a ruby statement",
		"this is a sentence with multiple new lines \n\n\n\n\nit should not be in tact":                      "this is a sentence with multiple new lines \nit should not be in tact",
	}
	for k, v := range tests {
		test := sanitiseRubyResponse(k)
		if !strings.EqualFold(strings.TrimSpace(test), strings.TrimSpace(v)) {
			log.Fatalf("Got `%s', Expected `%s`", strings.TrimSpace(test), strings.TrimSpace(v))
		}
	}
}

// This guy mocks out the underlying Service provider in the client,
// but executes actual client code. This means we don't spin up the real
// mock service but execute our code in isolation.
//
// Use this when you want too exercise the client code, but not shell out to Ruby.
// Where possible, outside of these tess, you should consider creating a mockClient{} object and
// stubbing out the required behaviour.
//
// Stubbing the exec.Cmd interface is hard, see fakeExec* functions for
// the magic.
func createMockClient(success bool) (*PactClient, *ServiceMock) {
	execFunc := fakeExecSuccessCommand
	if !success {
		execFunc = fakeExecFailCommand
	}
	svc := &ServiceMock{
		Cmd:               "test",
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

	d := newClient(svc, svc, svc, svc)
	d.TimeoutDuration = 100 * time.Millisecond
	return d, svc
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
	<-time.After(50 * time.Millisecond)

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		fmt.Fprintf(os.Stdout, "COMMAND: oh noes!\n")
		os.Exit(1)
	}

	// Success :)
	fmt.Fprintf(os.Stdout, "{\"summary_line\":\"1 examples, 0 failures\"}\n{\"summary_line\":\"1 examples, 0 failures\"}")
	os.Exit(0)
}

func Test_sanitiseRubyResponse(t *testing.T) {
	var tests = map[string]string{
		"this is a sentence with a hash # so it should be in tact":                                           "this is a sentence with a hash # so it should be in tact",
		"this is a sentence with a hash and newline\n#so it should not be in tact":                           "this is a sentence with a hash and newline",
		"this is a sentence with a ruby statement bundle exec rake pact:verify so it should not be in tact":  "",
		"this is a sentence with a ruby statement\nbundle exec rake pact:verify so it should not be in tact": "this is a sentence with a ruby statement",
		"this is a sentence with multiple new lines \n\n\n\n\nit should not be in tact":                      "this is a sentence with multiple new lines \nit should not be in tact",
	}
	for k, v := range tests {
		test := sanitiseRubyResponse(k)
		if !strings.EqualFold(strings.TrimSpace(test), strings.TrimSpace(v)) {
			log.Fatalf("Got `%s', Expected `%s`", strings.TrimSpace(test), strings.TrimSpace(v))
		}
	}
}
