package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/url"
	"strconv"
	"strings"
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

// Get a port given a URL
func getPort(rawURL string) int {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		if len(strings.Split(parsedURL.Host, ":")) == 2 {
			port, err := strconv.Atoi(strings.Split(parsedURL.Host, ":")[1])
			if err == nil {
				return port
			}
		}
		if parsedURL.Scheme == "https" {
			return 443
		}
		return 80
	}

	return -1
}

func getHTTPClient(port int) (*rpc.Client, error) {
	log.Println("[DEBUG] creating an HTTP client")
	err := waitForPort(port, fmt.Sprintf(`Timed out waiting for Daemon on port %d - are you
		sure it's running?`, port))
	if err != nil {
		return nil, err
	}
	return rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
}

// Use this to wait for a daemon to be running prior
// to running tests.
var waitForPort = func(port int, message string) error {
	log.Println("[DEBUG] waiting for port", port, "to become available")
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-timeout:
			log.Printf("[ERROR] Expected server to start < %s. %s", timeoutDuration, message)
			return fmt.Errorf("Expected server to start < %s. %s", timeoutDuration, message)
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
			log.Fatal("[ERROR] rpc:", err)
		}
	}

	if err == nil {
		waitForPort(res.Port, fmt.Sprintf(`Timed out waiting for Mock Server to
			start on port %d - are you sure it's running?`, res.Port))
	}

	return &res
}

// VerifyProvider runs the verification process against a running Provider.
func (p *PactClient) VerifyProvider(request *daemon.VerifyRequest) *daemon.Response {
	log.Println("[DEBUG] client: verifying a provider")

	port := getPort(request.ProviderBaseURL)

	waitForPort(port, fmt.Sprintf(`Timed out waiting for Provider API to start
		 on port %d - are you sure it's running?`, port))

	var res daemon.Response
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call("Daemon.VerifyProvider", request, &res)
		if err != nil {
			log.Println("[ERROR] rpc: ", err.Error())
		}
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
			log.Fatal("[ERROR] rpc:", err)
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
			log.Fatal("[ERROR] rpc:", err)
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
			log.Fatal("[ERROR] rpc:", err)
		}
	}
	return err
}
