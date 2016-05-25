package daemon

// Runs the RPC daemon for remote communication

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
)

// MockServer contains the RPC client interface to a Mock Server
type MockServer struct {
	Pid    int
	port   int
	status int
}

// Port returns the allocated mock servers port.
func (m *MockServer) Port() int {
	return m.port
}

// Status returns exit code of th eserver.
func (m *MockServer) Status() int {
	return m.status
}

// PublishRequest contains the details required to Publish Pacts to a broker.
type PublishRequest struct {
	// Array of local Pact files or directories containing them. Required.
	PactUrls []url.URL

	// URL to fetch the provider states for the given provider API. Optional.
	PactBroker url.URL

	// Username for Pact Broker basic authentication. Optional
	PactBrokerUsername string

	// Password for Pact Broker basic authentication. Optional
	PactBrokerPassword string
}

// PublishResponse contains the exit status and any message from the Broker.
type PublishResponse struct {
	// System exit code from the Publish task.
	ExitCode int

	// Error message (if any) from the publish process.
	Message string
}

// Stop stops the given mock server and captures the exit status.
func (m *MockServer) Stop() *MockServer {
	m.status = 0
	m.Pid = 0
	return m
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

// ListServers returns a slice of all running MockServers.
func (d *Daemon) ListServers(request interface{}, reply *[]MockServer) error {
	*reply = []MockServer{
		MockServer{Pid: 1},
	}

	return nil
}

// StopServer stops the given mock server.
func (d *Daemon) StopServer(request *MockServer, reply *MockServer) error {
	request.Stop()
	*reply = *request
	return nil
}

// Publish publishes Pact files from a given location (file/http).
func (d *Daemon) Publish(request *PublishRequest, reply *PublishResponse) error {
	*reply = *&PublishResponse{
		ExitCode: 0,
		Message:  "Success",
	}
	return nil
}

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
