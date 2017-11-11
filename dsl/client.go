package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pact-foundation/pact-go/types"
)

var (
	timeoutDuration       = 10 * time.Second
	commandStartServer    = "Daemon.StartServer"
	commandStopServer     = "Daemon.StopServer"
	commandVerifyProvider = "Daemon.VerifyProvider"
	commandListServers    = "Daemon.ListServers"
	commandStopDaemon     = "Daemon.StopDaemon"
)

// Client is the simplified remote interface to the Pact Daemon.
type Client interface {
	StartServer() *types.MockServer
}

// PactClient is the default implementation of the Client interface.
type PactClient struct {
	// Port the daemon is running on.
	Port int

	// Network Daemon is listening on
	Network string

	// Address the Daemon is listening on
	Address string
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

func getHTTPClient(port int, network string, address string) (*rpc.Client, error) {
	log.Println("[DEBUG] creating an HTTP client")
	err := waitForPort(port, network, address, fmt.Sprintf(`Timed out waiting for Daemon on port %d - are you
		sure it's running?`, port))
	if err != nil {
		return nil, err
	}
	return rpc.DialHTTP(network, fmt.Sprintf("%s:%d", address, port))
}

// Use this to wait for a daemon to be running prior
// to running tests.
var waitForPort = func(port int, network string, address string, message string) error {
	log.Println("[DEBUG] waiting for port", port, "to become available")
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-timeout:
			log.Printf("[ERROR] Expected server to start < %s. %s", timeoutDuration, message)
			return fmt.Errorf("Expected server to start < %s. %s", timeoutDuration, message)
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial(network, fmt.Sprintf("%s:%d", address, port))
			if err == nil {
				return nil
			}
		}
	}
}

// StartServer starts a remote Pact Mock Server.
func (p *PactClient) StartServer(args []string, port int) *types.MockServer {
	log.Println("[DEBUG] client: starting a server")
	var res types.MockServer
	client, err := getHTTPClient(p.Port, p.getNetworkInterface(), p.Address)
	if err == nil {
		args = append(args, []string{"--port", strconv.Itoa(port)}...)
		err = client.Call(commandStartServer, types.MockServer{Args: args}, &res)
		res.Port = port
		if err != nil {
			log.Println("[ERROR] rpc:", err.Error())
		}
	}

	if err == nil {
		waitForPort(port, p.getNetworkInterface(), p.Address, fmt.Sprintf(`Timed out waiting for Mock Server to
			start on port %d - are you sure it's running?`, port))
	}

	return &res
}

// VerifyProvider runs the verification process against a running Provider.
func (p *PactClient) VerifyProvider(request types.VerifyRequest) (types.ProviderVerifierResponse, error) {
	log.Println("[DEBUG] client: verifying a provider")

	port := getPort(request.ProviderBaseURL)

	waitForPort(port, p.getNetworkInterface(), p.Address, fmt.Sprintf(`Timed out waiting for Provider API to start
		 on port %d - are you sure it's running?`, port))

	var res types.ProviderVerifierResponse
	client, err := getHTTPClient(p.Port, p.getNetworkInterface(), p.Address)
	if err == nil {
		err = client.Call(commandVerifyProvider, request, &res)
		if err != nil {
			log.Println("[ERROR] rpc: ", err.Error())
		}
	}

	return res, err
}

// sanitiseRubyResponse removes Ruby-isms from the response content
// making the output much more human readable
func sanitiseRubyResponse(response string) string {
	log.Println("[TRACE] response from Ruby process pre-sanitisation:", response)

	r := regexp.MustCompile("(?m)^\\s*#.*$")
	s := r.ReplaceAllString(response, "")

	r = regexp.MustCompile("(?m).*bundle exec rake pact:verify.*$")
	s = r.ReplaceAllString(s, "")

	r = regexp.MustCompile("\\n+")
	s = r.ReplaceAllString(s, "\n")

	return s
}

// ListServers lists all running Pact Mock Servers.
func (p *PactClient) ListServers() *types.PactListResponse {
	log.Println("[DEBUG] client: listing servers")
	var res types.PactListResponse
	client, err := getHTTPClient(p.Port, p.getNetworkInterface(), p.Address)
	if err == nil {
		err = client.Call(commandListServers, types.MockServer{}, &res)
		if err != nil {
			log.Println("[ERROR] rpc:", err.Error())
		}
	}
	return &res
}

// StopServer stops a remote Pact Mock Server.
func (p *PactClient) StopServer(server *types.MockServer) *types.MockServer {
	log.Println("[DEBUG] client: stop server")
	var res types.MockServer
	client, err := getHTTPClient(p.Port, p.getNetworkInterface(), p.Address)
	if err == nil {
		err = client.Call(commandStopServer, server, &res)
		if err != nil {
			log.Println("[ERROR] rpc:", err.Error())
		}
	}
	return &res
}

// StopDaemon remotely shuts down the Pact Daemon.
func (p *PactClient) StopDaemon() error {
	log.Println("[DEBUG] client: stop daemon")
	var req, res string
	client, err := getHTTPClient(p.Port, p.getNetworkInterface(), p.Address)
	if err == nil {
		err = client.Call(commandStopDaemon, &req, &res)
		if err != nil {
			log.Println("[ERROR] rpc:", err.Error())
		}
	}
	return err
}

// getNetworkInterface returns a default interface to communicate to the Daemon
// if none specified
func (p *PactClient) getNetworkInterface() string {
	if p.Network == "" {
		return "tcp"
	}
	return p.Network
}
