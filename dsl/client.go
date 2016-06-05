package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/pact-foundation/pact-go/daemon"
)

var timeoutDuration = 1 * time.Second

// Client is the simplified remote interface to the Pact Daemon.
type Client interface {
	StartServer() *daemon.PactMockServer
}

// PactClient is the default implementation of the Client interface.
type PactClient struct {
	// Port the daemon is running on.
	Port int
}

func getHTTPClient(port int) (*rpc.Client, error) {
	log.Println("[DEBUG] creating an HTTP client")
	err := waitForPort(port)
	if err != nil {
		return nil, err
	}
	return rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
}

// Use this to wait for a daemon to be running prior
// to running tests.
func waitForPort(port int) error {
	log.Println("[DEBUG] waiting for port", port, "to become available")
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-timeout:
			log.Printf("[ERROR] Expected server to start < %s", timeoutDuration)
			return fmt.Errorf("Expected server to start < %s", timeoutDuration)
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				return nil
			}
		}
	}
}

// StartServer starts a remote Pact Mock Server.
func (p *PactClient) StartServer() *daemon.PactMockServer {
	log.Println("[DEBUG] client: starting a server")
	var res daemon.PactMockServer
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call("Daemon.StartServer", daemon.PactMockServer{}, &res)
		if err != nil {
			log.Fatal("rpc error:", err)
		}
	}

	if err == nil {
		waitForPort(res.Port)
	}

	return &res
}

// ListServers starts a remote Pact Mock Server.
func (p *PactClient) ListServers() *daemon.PactListResponse {
	log.Println("[DEBUG] client: listing servers")
	var res daemon.PactListResponse
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call("Daemon.ListServers", daemon.PactMockServer{}, &res)
		if err != nil {
			log.Fatal("rpc error:", err)
		}
	}
	return &res
}

// StopServer stops a remote Pact Mock Server.
func (p *PactClient) StopServer(server *daemon.PactMockServer) *daemon.PactMockServer {
	log.Println("[DEBUG] client: stop server")
	var res daemon.PactMockServer
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call("Daemon.StopServer", server, &res)
		if err != nil {
			log.Fatal("rpc error:", err)
		}
	}
	return &res
}

// StopDaemon remotely shuts down the Pact Daemon.
func (p *PactClient) StopDaemon() error {
	log.Println("[DEBUG] client: stop daemon")
	var req, res string
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call("Daemon.StopDaemon", &req, &res)
		if err != nil {
			log.Fatal("rpc error:", err)
		}
	}
	return err
}
