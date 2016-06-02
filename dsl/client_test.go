package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
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

func createDaemon(port int) *daemon.Daemon {
	s := &daemon.PactMockService{}
	_, svc := s.NewService()
	d := daemon.NewDaemon(svc)
	go d.StartDaemon(port)
	return d
}

// func TestRPCClient_ListFail(t *testing.T) {
// 	client := &PactClient{ /* don't supply port */ }
//
// }

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_List(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)
	waitForPortInTest(port, t)
	defer waitForDaemonToShutdown(port, t)
	client := &PactClient{port: port}
	server := client.StartServer()

	waitForPortInTest(server.Port, t)

	s := client.ListServers()

	if len(s.Servers) != 1 {
		t.Fatalf("Expected 1 server to be running, got %d", len(s.Servers))
	}

	// client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	// var res daemon.PactMockServer
	// err = client.Call("Daemon.StartServer", daemon.PactMockServer{}, &res)
	// if err != nil {
	// 	log.Fatal("rpc error:", err)
	// }
	//
	// waitForPortInTest(res.Port, t)
	//
	// client, err = rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	// var res2 daemon.PactListResponse
	// err = client.Call("Daemon.ListServers", daemon.PactMockServer{}, &res2)
	// if err != nil {
	// 	log.Fatal("rpc error:", err)
	// }
}

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_StartServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)
	waitForPortInTest(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res daemon.PactMockServer
	err = client.Call("Daemon.StartServer", daemon.PactMockServer{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	<-time.After(10 * time.Second)
	waitForDaemonToShutdown(port, t)
}

/*
// Integration style test: Can a client hit each endpoint?
func TestRPCClient_StopServer(t *testing.T) {
	port, _ := utils.GetFreePort()
	// defer waitForDaemonToShutdown(port, daemon, t)
	d := createDaemon(port)
	waitForPortInTest(port, t)

	var cmd *exec.Cmd
	for _, s := range manager.List() {
		cmd = s
	}
	request := daemon.PactMockServer{
		Pid: cmd.Process.Pid,
	}

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res *daemon.PactMockServer
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

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_Verify(t *testing.T) {
	port, _ := utils.GetFreePort()
	d := createDaemon(port)
	// defer waitForDaemonToShutdown(port, daemon, t)
	waitForPortInTest(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res daemon.PactResponse
	err = client.Call("Daemon.Verify", &daemon.VerifyRequest{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if res.ExitCode != 0 {
		t.Fatalf("Expected exit code to be 0, got: %d", res.ExitCode)
	}
	if res.Message != "" {
		t.Fatalf("Expected message to be blank but got: %s", res.Message)
	}
}

// Integration style test: Can a client hit each endpoint?
func TestRPCClient_Publish(t *testing.T) {
	port, _ := utils.GetFreePort()
	d := createDaemon(port)
	// defer waitForDaemonToShutdown(port, daemon, t)
	go d.StartDaemon(port)
	waitForPortInTest(port, t)

	client, err := rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
	var res PactResponse
	err = client.Call("Daemon.Publish", &PublishRequest{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}

	if res.ExitCode != 0 {
		t.Fatalf("Expected exit code to be 0, got: %d", res.ExitCode)
	}

	if res.Message != "" {
		t.Fatalf("Expected message to be blank but got: %s", res.Message)
	}
}
*/
