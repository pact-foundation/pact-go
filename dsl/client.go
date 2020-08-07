package dsl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pact-foundation/pact-go/client"
	"github.com/pact-foundation/pact-go/types"
)

// Client is the interface
type Client interface {
	// StartServer starts a remote Pact Mock Server.
	StartServer(args []string, port int) *types.MockServer

	// ListServers lists all known Mock Servers
	ListServers() []*types.MockServer

	// StopServer stops a remote Pact Mock Server.
	StopServer(server *types.MockServer) (*types.MockServer, error)

	// RemoveAllServers stops all remote Pact Mock Servers.
	RemoveAllServers(server *types.MockServer) []*types.MockServer

	// VerifyProvider runs the verification process against a running Provider.
	VerifyProvider(request types.VerifyRequest) ([]types.ProviderVerifierResponse, error)

	// UpdateMessagePact adds a pact message to a contract file
	UpdateMessagePact(request types.PactMessageRequest) error

	// ReifyMessage takes a structured object, potentially containing nested Matchers
	// and returns an object with just the example (generated) content
	// The object may be a simple JSON primitive e.g. string or number or a complex object
	ReifyMessage(request *types.PactReificationRequest) (res *types.ReificationResponse, err error)

	// PublishPacts publishes pact files to a Pact Broker
	PublishPacts(request types.PublishRequest) error
}

// PactClient is the main interface into starting/stopping
// the underlying Pact CLI subsystem
type PactClient struct {
	pactMockSvcManager     client.Service
	verificationSvcManager client.Service
	messageSvcManager      client.Service
	publishSvcManager      client.Service

	// Track mock servers
	Servers []MockService

	// Network Daemon is listening on
	Network string

	// Address the Daemon is listening on
	Address string

	// TimeoutDuration specifies how long to wait for Pact CLI to start
	TimeoutDuration time.Duration
}

// newClient creates a new Pact client manager with the provided services
func newClient(MockServiceManager client.Service, verificationServiceManager client.Service, messageServiceManager client.Service, publishServiceManager client.Service) *PactClient {
	MockServiceManager.Setup()
	verificationServiceManager.Setup()
	messageServiceManager.Setup()
	publishServiceManager.Setup()

	return &PactClient{
		pactMockSvcManager:     MockServiceManager,
		verificationSvcManager: verificationServiceManager,
		messageSvcManager:      messageServiceManager,
		publishSvcManager:      publishServiceManager,
		TimeoutDuration:        10 * time.Second,
	}
}

// NewClient creates a new Pact client manager with defaults
func NewClient() *PactClient {
	return newClient(&client.MockService{}, &client.VerificationService{}, &client.MessageService{}, &client.PublishService{})
}

// StartServer starts a remote Pact Mock Server.
func (p *PactClient) StartServer(args []string, port int) *types.MockServer {
	log.Println("[DEBUG] client: starting a server with args:", args, "port:", port)
	args = append(args, []string{"--port", strconv.Itoa(port)}...)
	svc := p.pactMockSvcManager.NewService(args)
	cmd := svc.Start()

	waitForPort(port, p.getNetworkInterface(), p.Address, p.TimeoutDuration,
		fmt.Sprintf(`Timed out waiting for Mock Server to start on port %d - are you sure it's running?`, port))

	return &types.MockServer{
		Pid:  cmd.Process.Pid,
		Port: port,
	}
}

// ListServers lists all known Mock Servers
func (p *PactClient) ListServers() []*types.MockServer {
	log.Println("[DEBUG] client: starting a server")

	var servers []*types.MockServer

	for port, s := range p.pactMockSvcManager.List() {
		servers = append(servers, &types.MockServer{
			Pid:  s.Process.Pid,
			Port: port,
		})
	}

	return servers
}

// StopServer stops a remote Pact Mock Server.
func (p *PactClient) StopServer(server *types.MockServer) (*types.MockServer, error) {
	log.Println("[DEBUG] client: stop server")

	// TODO: Need to be able to get a non-zero exit code here!
	_, server.Error = p.pactMockSvcManager.Stop(server.Pid)
	return server, server.Error
}

// RemoveAllServers stops all remote Pact Mock Servers.
func (p *PactClient) RemoveAllServers(server *types.MockServer) []*types.MockServer {
	log.Println("[DEBUG] client: stop server")

	for _, s := range p.verificationSvcManager.List() {
		if s != nil {
			p.pactMockSvcManager.Stop(s.Process.Pid)
		}
	}
	return nil
}

// VerifyProvider runs the verification process against a running Provider.
// TODO: extract/refactor the stdout/error streaems from these functions
func (p *PactClient) VerifyProvider(request types.VerifyRequest) ([]types.ProviderVerifierResponse, error) {
	log.Println("[DEBUG] client: verifying a provider")
	response := make([]types.ProviderVerifierResponse, 0)

	// Convert request into flags, and validate request
	err := request.Validate()
	if err != nil {
		return response, err
	}

	address := getAddress(request.ProviderBaseURL)
	port := getPort(request.ProviderBaseURL)

	err = waitForPort(port, p.getNetworkInterface(), address, p.TimeoutDuration,
		fmt.Sprintf(`Timed out waiting for Provider API to start on port %d - are you sure it's running?`, port))

	if err != nil {
		return response, err
	}

	// Run command, splitting out stderr and stdout. The command can fail for
	// several reasons:
	// 1. Command is unable to run at all.
	// 2. Command runs, but fails for unknown reason.
	// 3. Command runs, and returns exit status 1 because the tests fail.
	//
	// First, attempt to decode the response of the stdout.
	// If that is successful, we are at case 3. Return stdout as message, no error.
	// Else, return an error, include stderr and stdout in both the error and message.
	svc := p.verificationSvcManager.NewService(request.Args)
	cmd := svc.Command()

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return response, err
	}
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return response, err
	}

	// Buffered channel: wait for all reading to complete
	var wg sync.WaitGroup
	verifications := []string{}
	var stdErr strings.Builder

	// Split by lines, as the content is JSONL formatted
	// Each pact is verified by line, and the results (as JSON) sent to stdout.
	// See https://github.com/pact-foundation/pact-go/issues/88#issuecomment-404686337
	stdOutScanner := bufio.NewScanner(stdOutPipe)
	go func() {
		wg.Add(1)
		defer wg.Done()
		stdOutBuf := make([]byte, bufio.MaxScanTokenSize)
		stdOutScanner.Buffer(stdOutBuf, 64*1024*1024)

		for stdOutScanner.Scan() {
			verifications = append(verifications, stdOutScanner.Text())
		}
	}()

	// Scrape errors
	stdErrScanner := bufio.NewScanner(stdErrPipe)
	go func() {
		wg.Add(1)
		defer wg.Done()
		for stdErrScanner.Scan() {
			stdErr.WriteString(fmt.Sprintf("%s\n", stdErrScanner.Text()))
		}

	}()

	err = cmd.Start()
	if err != nil {
		return response, err
	}

	// Wait for watch goroutine before Cmd.Wait(), race condition!
	err = cmd.Wait()
	wg.Wait()

	var verification types.ProviderVerifierResponse
	for _, v := range verifications {
		v = strings.TrimSpace(v)

		// TODO: fix once https://github.com/pact-foundation/pact-provider-verifier/issues/26
		//       is addressed
		// logging to stdout breaks the JSON response
		// https://github.com/pact-foundation/pact-ruby/commit/06fa61581512ba5570c315d089f2c0fc23c8cb11
		if v != "" && strings.Index(v, "INFO") != 0 {
			dErr := json.Unmarshal([]byte(v), &verification)

			response = append(response, verification)

			if dErr != nil {
				err = dErr
			}
		}
	}

	if err == nil {
		return response, err
	}

	return response, fmt.Errorf("error verifying provider: %s\n\nSTDERR:\n%s\n\nSTDOUT:\n%s", err, stdErr.String(), strings.Join(verifications, "\n"))
}

// UpdateMessagePact adds a pact message to a contract file
func (p *PactClient) UpdateMessagePact(request types.PactMessageRequest) error {
	log.Println("[DEBUG] client: adding pact message...")

	// Convert request into flags, and validate request
	err := request.Validate()
	if err != nil {
		return err
	}

	svc := p.messageSvcManager.NewService(request.Args)
	cmd := svc.Command()

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	stdOut, err := ioutil.ReadAll(stdOutPipe)
	if err != nil {
		return err
	}
	stdErr, err := ioutil.ReadAll(stdErrPipe)
	if err != nil {
		return err
	}

	err = cmd.Wait()

	if err == nil {
		return nil
	}

	return fmt.Errorf("error creating message: %s\n\nSTDERR:\n%s\n\nSTDOUT:\n%s", err, stdErr, stdOut)
}

// PublishPacts publishes a set of pacts to a pact broker
func (p *PactClient) PublishPacts(request types.PublishRequest) error {
	svc := p.publishSvcManager.NewService(request.Args)
	log.Println("[DEBUG] about to publish pacts")
	cmd := svc.Start()

	log.Println("[DEBUG] waiting for response")
	err := cmd.Wait()

	log.Println("[DEBUG] response from publish", err)

	return err
}

// ReifyMessage takes a structured object, potentially containing nested Matchers
// and returns an object with just the example (generated) content
// The object may be a simple JSON primitive e.g. string or number or a complex object
func (p *PactClient) ReifyMessage(request *types.PactReificationRequest) (res *types.ReificationResponse, err error) {
	log.Println("[DEBUG] client: adding pact message...")

	var responseObject interface{}
	res = &types.ReificationResponse{
		Response: responseObject,
	}

	// Convert request into flags, and validate request
	err = request.Validate()
	if err != nil {
		return
	}

	svc := p.messageSvcManager.NewService(request.Args)
	cmd := svc.Command()

	stdOutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stdErrPipe, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return
	}
	stdOut, err := ioutil.ReadAll(stdOutPipe)
	if err != nil {
		return
	}
	stdErr, err := ioutil.ReadAll(stdErrPipe)
	if err != nil {
		return
	}

	err = cmd.Wait()

	res.ResponseRaw = stdOut
	decoder := json.NewDecoder(bytes.NewReader(stdOut))

	dErr := decoder.Decode(&res.Response)
	if dErr == nil {
		return
	}

	if err == nil {
		err = dErr
	}

	if err == nil {
		return
	}

	err = fmt.Errorf("error creating message: %s\n\nSTDERR:\n%s\n\nSTDOUT:\n%s", err, stdErr, stdOut)

	return
}

// Get a port given a URL
func getPort(rawURL string) int {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		splitHost := strings.Split(parsedURL.Host, ":")
		if len(splitHost) == 2 {
			port, err := strconv.Atoi(splitHost[1])
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

// Get the address given a URL
func getAddress(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	splitHost := strings.Split(parsedURL.Host, ":")
	return splitHost[0]
}

// Use this to wait for a port to be running prior
// to running tests.
var waitForPort = func(port int, network string, address string, timeoutDuration time.Duration, message string) error {
	log.Println("[DEBUG] waiting for port", port, "to become available on", address, "after", timeoutDuration)
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

// getNetworkInterface returns a default interface to communicate to the Daemon
// if none specified
func (p *PactClient) getNetworkInterface() string {
	if p.Network == "" {
		return "tcp"
	}
	return p.Network
}
