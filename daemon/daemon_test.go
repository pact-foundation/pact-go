package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

type MockService struct {
}

func (r MockService) Run(command string, args ...string) ([]byte, error) {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	out, err := cmd.CombinedOutput()
	return out, err
}

func TestHello(t *testing.T) {
	runner = MockService{}
	out := Hello()
	if out == "testing helper process" {
		t.Logf("out was eq to %s", string(out))
	}
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	fmt.Println("testing helper process")
}

func TestStartDaemon(t *testing.T) {

}

func TestStartDaemon_Fail(t *testing.T) {

}

func TestStartServer(t *testing.T) {
	daemon := CreateDaemon()
	req := MockServer{Pid: 1234}
	res := MockServer{}
	err := daemon.StartServer(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if !reflect.DeepEqual(req, res) {
		t.Fatalf("Req and Res did not match")
	}
}

func TestListServers(t *testing.T) {
	daemon := &Daemon{}
	var res []MockServer
	err := daemon.ListServers(nil, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("Expected array of len 1, got: %d", len(res))
	}
}

func TestStopServer(t *testing.T) {
	daemon := &Daemon{}

	req := MockServer{Pid: 1234}
	res := MockServer{}
	err := daemon.StopServer(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if res.Pid != 0 {
		t.Fatalf("Expected PID to be 0 but got: %d", res.Pid)
	}

	if res.Status != 0 {
		t.Fatalf("Expected exit status to be 0 but got: %d", res.Status)
	}
}

func TestStartServer_Fail(t *testing.T) {

}

func TestVerification(t *testing.T) {

}

func TestVerification_Fail(t *testing.T) {

}

func TestPublish(t *testing.T) {
	daemon := &Daemon{}
	req := PublishRequest{}
	var res PactResponse
	err := daemon.Publish(&req, &res)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if res.ExitCode != 0 {
		t.Fatalf("Expected exit code to be 0 but got: %d", res.ExitCode)
	}

	if res.Message != "Success" {
		t.Fatalf("Expected message to be 'Success' but got: %s", res.Message)
	}
}

func TestPublish_Fail(t *testing.T) {

}
