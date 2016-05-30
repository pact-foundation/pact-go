package daemon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

// func createDefaultDaemon() {
// 	getCommandPath = func() string {
// 		dir, _ := os.Getwd()
// 		return fmt.Sprintf(filepath.Join(dir, "../", "pact-mock-service", "bin", "pact-mock-service"))
// 	}
// }

// type MockService struct {
// }
//
// func (r MockService) Run(command string, args ...string) ([]byte, error) {
// 	cs := []string{"-test.run=TestHelperProcess", "--"}
// 	cs = append(cs, args...)
// 	cmd := exec.Command(os.Args[0], cs...)
// 	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
// 	out, err := cmd.CombinedOutput()
// 	return out, err
// }
//
// func TestHello(t *testing.T) {
// 	runner = MockService{}
// 	out := Hello()
// 	if out == "testing helper process" {
// 		t.Logf("out was eq to %s", string(out))
// 	}
// }
//
// func TestHelperProcess(*testing.T) {
// 	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
// 		return
// 	}
// 	defer os.Exit(0)
// 	fmt.Println("testing helper process")
// }

func createMockedDaemon() *Daemon {
	svc := &ServiceMock{
		Command:           "test",
		Args:              []string{},
		ServiceStopResult: true,
		ServiceStopError:  nil,
		ServiceList: map[int]*exec.Cmd{
			1: fakeExecCommand("", true, ""),
			2: fakeExecCommand("", true, ""),
			3: fakeExecCommand("", true, ""),
		},
		ServiceStartCmd: nil,
	}
	return NewDaemon(svc)
}

func TestNewDaemon(t *testing.T) {
	var daemon interface{}
	daemon = createMockedDaemon()

	if _, ok := daemon.(Daemon); !ok {
		t.Fatalf("must be a Daemon")
	}
}

func TestStartAndStopDaemon(t *testing.T) {
	daemon := createMockedDaemon()
	go daemon.StartDaemon()

	for {
		select {
		case <-time.After(1 * time.Second):
			t.Fatalf("Expected server to start < 1s.")
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", ":6666")
			if err == nil {
				daemon.signalChan <- os.Interrupt
				return
			}
		}
	}
}

// Adapted from http://npf.io/2015/06/testing-exec-command/
func fakeExecCommand(command string, success bool, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", fmt.Sprintf("GO_WANT_HELPER_PROCESS_TO_SUCCEED=%t", success)}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	fmt.Fprintln(os.Stdout, "HELLLlloooo")
	<-time.After(30 * time.Second)
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// some code here to check arguments perhaps?
	// Fail :(
	if os.Getenv("GO_WANT_HELPER_PROCESS_TO_SUCCEED") == "false" {
		os.Exit(1)
	}

	// Success :)
	os.Exit(0)
}

func TestDaemonShutdown(t *testing.T) {
	daemon := createMockedDaemon()
	manager := daemon.pactMockSvcManager.(*ServiceMock)

	// Start all processes to get the Pids!
	for _, s := range manager.ServiceList {
		s.Start()
	}

	daemon.Shutdown()

	if manager.ServiceStopCount != 3 {
		t.Fatalf("Expected Stop() to be called 3 times but got: %d", manager.ServiceStopCount)
	}
}

func TestStartDaemon_Fail(t *testing.T) {

}

func TestStartServer(t *testing.T) {
	daemon := createMockedDaemon()
	req := PactMockServer{Pid: 1234}
	res := PactMockServer{}
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
	var res []PactMockServer
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

	req := PactMockServer{Pid: 1234}
	res := PactMockServer{}
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
