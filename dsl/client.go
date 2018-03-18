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
	timeoutDuration = 10 * time.Second
)

// PactClient is the default implementation of the Client interface.
type PactClient struct {
	// Address the Daemon is listening on
	Address string

	// Track mock servers
	Servers []MockService

	// Track stub servers
	// Stubs []StubService
}

// StartServer starts a remote Pact Mock Server.
func (p *PactClient) StartServer(args []string, port int) *MockService {
	log.Println("[DEBUG] client: starting a server")

	return nil
}

// ListServers list all available Mock Servers
func (p *PactClient) ListServers(args []string, port int) []*types.MockServer {
	log.Println("[DEBUG] client: starting a server")

	return nil
}

// StopServer stops a remote Pact Mock Server.
func (p *PactClient) StopServer(server *types.MockServer) *types.MockServer {
	log.Println("[DEBUG] client: stop server")

	return nil
}

// RemoveAllServers stops all remote Pact Mock Servers.
func (p *PactClient) RemoveAllServers(server *types.MockServer) *[]types.MockServer {
	log.Println("[DEBUG] client: stop server")

	return nil
}

// VerifyProvider runs the verification process against a running Provider.
func (p *PactClient) VerifyProvider(request types.VerifyRequest) (types.ProviderVerifierResponse, error) {
	log.Println("[DEBUG] client: verifying a provider")

	return types.ProviderVerifierResponse{}, nil
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
