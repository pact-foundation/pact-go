package daemon

// Runs the RPC daemon for remote communication

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

// MockServer contains the settings to run a local Mock Server
type MockServer struct {
	Pid  int
	port int
}

// Port returns the allocated mock servers port.
func (m *MockServer) Port() int {
	return m.port
}

// Daemon wraps the commands for the RPC server.
type Daemon struct {
}

// StartServer starts a mock server and returns a pointer to a MockServer
// struct.
func (d *Daemon) StartServer(request *MockServer, reply *MockServer) error {
	*reply = *request

	return nil
}

// StopServer stops the given mock server.

// Publish publishes Pact files from a given location (file/http).

// Verify runs the Pact verification process against a given API Provider.

// StartDaemon starts the daemon RPC server.
func StartDaemon() {
	fmt.Println("Starting daemon on port 6666")
	server := new(Daemon)
	rpc.Register(server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":6666")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
