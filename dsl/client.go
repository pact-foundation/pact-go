package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/mefellows/pact-go/daemon"
)

// Client is the simplified remote interface to the Pact Daemon
type Client interface {
	StartServer() *daemon.PactMockServer
}

// PactClient is the default implementation of the Client interface.
type PactClient struct {
	// Port the daemon is running on
	Port int
}

func getHTTPClient(port int) (*rpc.Client, error) {
	waitForPort(port)
	return rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
}

// Use this to wait for a daemon to be running prior
// to running tests
func waitForPort(port int) {
	fmt.Printf("client - Waiting for daemon port: %d", port)
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			log.Fatalf("Expected server to start < 1s.")
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				return
			}
		}
	}
}

// StartServer starts a remote Pact Mock Server.
func (p *PactClient) StartServer() *daemon.PactMockServer {
	var res daemon.PactMockServer
	client, err := getHTTPClient(p.Port)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	err = client.Call("Daemon.StartServer", daemon.PactMockServer{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	return &res
}

// ListServers starts a remote Pact Mock Server.
func (p *PactClient) ListServers() *daemon.PactListResponse {
	var res daemon.PactListResponse
	client, err := getHTTPClient(p.Port)
	err = client.Call("Daemon.ListServers", daemon.PactMockServer{}, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	return &res
}

// StopServer stops a remote Pact Mock Server.
func (p *PactClient) StopServer(server *daemon.PactMockServer) *daemon.PactMockServer {
	client, err := getHTTPClient(p.Port)
	var res daemon.PactMockServer
	err = client.Call("Daemon.StopServer", server, &res)
	if err != nil {
		log.Fatal("rpc error:", err)
	}
	return &res
}
