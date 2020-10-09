/*
Package dsl contains the main Pact DSL used in the Consumer
collaboration test cases, and Provider contract test verification.
*/
package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/install"
	"github.com/pact-foundation/pact-go/proxy"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *types.MockServer

	// Pact RPC Client.
	pactClient Client

	// Consumer is the name of the Consumer/Client.
	Consumer string

	// Provider is the name of the Providing service.
	Provider string

	// Interactions contains all of the Mock Service Interactions to be setup.
	Interactions []*Interaction

	// MessageInteractions contains all of the Message based interactions to be setup.
	MessageInteractions []*Message

	// Log levels.
	LogLevel string

	// Used to detect if logging has been configured.
	logFilter *logutils.LevelFilter

	// Location of Pact external service invocation output logging.
	// Defaults to `<cwd>/logs`.
	LogDir string

	// Pact files will be saved in this folder.
	// Defaults to `<cwd>/pacts`.
	PactDir string

	// PactFileWriteMode specifies how to write to the Pact file, for the life
	// of a Mock Service.
	// "overwrite" will always truncate and replace the pact after each run
	// "merge" will append to the pact file, which is useful if your tests
	// are split over multiple files and instantiations of a Mock Server
	// See https://github.com/pact-foundation/pact-ruby/blob/master/documentation/configuration.md#pactfile_write_mode
	PactFileWriteMode string

	// Specify which version of the Pact Specification should be used (1 or 2).
	// Defaults to 2.
	SpecificationVersion int

	// Host is the address of the Mock and Verification Service runs on
	// Examples include 'localhost', '127.0.0.1', '[::1]'
	// Defaults to 'localhost'
	Host string

	// Network is the network of the Mock and Verification Service
	// Examples include 'tcp', 'tcp4', 'tcp6'
	// Defaults to 'tcp'
	Network string

	// Ports MockServer can be deployed to, can be CSV or Range with a dash
	// Example "1234", "12324,5667", "1234-5667"
	AllowedMockServerPorts string

	// DisableToolValidityCheck prevents CLI version checking - use this carefully!
	// The ideal situation is to check the tool installation with  before running
	// the tests, which should speed up large test suites significantly
	DisableToolValidityCheck bool

	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration

	// Check if CLI tools are up to date
	toolValidityCheck bool
}

// AddMessage creates a new asynchronous consumer expectation
func (p *Pact) AddMessage() *Message {
	p.setupLogging()
	log.Println("[DEBUG] pact add message")

	m := &Message{}
	p.MessageInteractions = append(p.MessageInteractions, m)
	return m
}

// AddInteraction creates a new Pact interaction, initialising all
// required things. Will automatically start a Mock Service if none running.
func (p *Pact) AddInteraction() *Interaction {
	p.Setup(true)
	log.Println("[DEBUG] pact add interaction")
	i := &Interaction{}
	p.Interactions = append(p.Interactions, i)
	return i
}

// Setup starts the Pact Mock Server. This is usually called before each test
// suite begins. AddInteraction() will automatically call this if no Mock Server
// has been started.
func (p *Pact) Setup(startMockServer bool) *Pact {
	p.setupLogging()
	log.Println("[DEBUG] pact setup")
	dir, _ := os.Getwd()

	if p.Network == "" {
		p.Network = "tcp"
	}

	if !p.toolValidityCheck && !(p.DisableToolValidityCheck || os.Getenv("PACT_DISABLE_TOOL_VALIDITY_CHECK") != "") {
		checkCliCompatibility()
		p.toolValidityCheck = true
	}

	if p.Host == "" {
		p.Host = "localhost"
	}

	if p.LogDir == "" {
		p.LogDir = fmt.Sprintf(filepath.Join(dir, "logs"))
	}

	if p.PactDir == "" {
		p.PactDir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	}

	if p.SpecificationVersion == 0 {
		p.SpecificationVersion = 2
	}

	if p.ClientTimeout == 0 {
		p.ClientTimeout = 10 * time.Second
	}

	if p.pactClient == nil {
		c := NewClient()
		c.TimeoutDuration = p.ClientTimeout
		p.pactClient = c
	}

	if p.PactFileWriteMode == "" {
		p.PactFileWriteMode = "overwrite"
	}

	// Need to predefine due to scoping
	var port int
	var perr error
	if p.AllowedMockServerPorts != "" {
		port, perr = utils.FindPortInRange(p.AllowedMockServerPorts)
	} else {
		port, perr = utils.GetFreePort()
	}
	if perr != nil {
		log.Println("[ERROR] unable to find free port, mockserver will fail to start")
	}

	if p.Server == nil && startMockServer {
		log.Println("[DEBUG] starting mock service on port:", port)
		args := []string{
			"--pact-specification-version",
			fmt.Sprintf("%d", p.SpecificationVersion),
			"--pact-dir",
			filepath.FromSlash(p.PactDir),
			"--log",
			filepath.FromSlash(p.LogDir + "/" + "pact.log"),
			"--consumer",
			p.Consumer,
			"--provider",
			p.Provider,
			"--pact-file-write-mode",
			p.PactFileWriteMode,
		}

		p.Server = p.pactClient.StartServer(args, port)
	}

	return p
}

// Configure logging
func (p *Pact) setupLogging() {
	if p.logFilter == nil {
		if p.LogLevel == "" {
			p.LogLevel = "INFO"
		}
		p.logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"},
			MinLevel: logutils.LogLevel(p.LogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(p.logFilter)
	}
	log.Println("[DEBUG] pact setup logging")
}

// Teardown stops the Pact Mock Server. This usually is called on completion
// of each test suite.
func (p *Pact) Teardown() *Pact {
	log.Println("[DEBUG] teardown")
	if p.Server != nil {
		server, err := p.pactClient.StopServer(p.Server)

		if err != nil {
			log.Println("error:", err)
		}
		p.Server = server
	}
	return p
}

// Verify runs the current test case against a Mock Service.
// Will cleanup interactions between tests within a suite.
func (p *Pact) Verify(integrationTest func() error) error {
	p.Setup(true)
	log.Println("[DEBUG] pact verify")
	var err error

	// Check if we are verifying messages or if we actually have interactions
	if len(p.Interactions) == 0 {
		return errors.New("there are no interactions to be verified")
	}

	mockServer := &MockService{
		BaseURL:  fmt.Sprintf("http://%s:%d", p.Host, p.Server.Port),
		Consumer: p.Consumer,
		Provider: p.Provider,
	}

	// Cleanup all interactions
	defer func(mockServer *MockService) {
		log.Println("[DEBUG] clearing interactions")

		p.Interactions = make([]*Interaction, 0)
		err = mockServer.DeleteInteractions()
	}(mockServer)

	for _, interaction := range p.Interactions {
		err = mockServer.AddInteraction(interaction)
		if err != nil {
			return err
		}
	}

	// Run the integration test
	err = integrationTest()
	if err != nil {
		return err
	}

	// Run Verification Process
	err = mockServer.Verify()
	if err != nil {
		return err
	}

	return err
}

// WritePact should be called writes when all tests have been performed for a
// given Consumer <-> Provider pair. It will write out the Pact to the
// configured file.
func (p *Pact) WritePact() error {
	p.Setup(true)
	log.Println("[DEBUG] pact write Pact file")
	mockServer := MockService{
		BaseURL:           fmt.Sprintf("http://%s:%d", p.Host, p.Server.Port),
		Consumer:          p.Consumer,
		Provider:          p.Provider,
		PactFileWriteMode: p.PactFileWriteMode,
	}
	err := mockServer.WritePact()
	if err != nil {
		return err
	}

	return nil
}

// VerifyProviderRaw reads the provided pact files and runs verification against
// a running Provider API, providing raw response from the Verification process.
//
// Order of events: BeforeEach, stateHandlers, requestFilter(pre <execute provider> post), AfterEach
func (p *Pact) VerifyProviderRaw(request types.VerifyRequest) ([]types.ProviderVerifierResponse, error) {
	p.Setup(false)
	res := make([]types.ProviderVerifierResponse, 0)

	u, err := url.Parse(request.ProviderBaseURL)

	if err != nil {
		return res, err
	}

	m := []proxy.Middleware{}

	if request.BeforeEach != nil {
		m = append(m, BeforeEachMiddleware(request.BeforeEach))
	}

	if request.AfterEach != nil {
		m = append(m, AfterEachMiddleware(request.AfterEach))
	}

	if len(request.StateHandlers) > 0 {
		m = append(m, stateHandlerMiddleware(request.StateHandlers))
	}

	if request.RequestFilter != nil {
		m = append(m, request.RequestFilter)
	}

	// Configure HTTP Verification Proxy
	opts := proxy.Options{
		TargetAddress:             fmt.Sprintf("%s:%s", u.Hostname(), u.Port()),
		TargetScheme:              u.Scheme,
		TargetPath:                u.Path,
		Middleware:                m,
		InternalRequestPathPrefix: providerStatesSetupPath,
		CustomTLSConfig:           request.CustomTLSConfig,
	}

	// Starts the message wrapper API with hooks back to the state handlers
	// This maps the 'description' field of a message pact, to a function handler
	// that will implement the message producer. This function must return an object and optionally
	// and error. The object will be marshalled to JSON for comparison.
	port, err := proxy.HTTPReverseProxy(opts)

	// Backwards compatibility, setup old provider states URL if given
	// Otherwise point to proxy
	setupURL := request.ProviderStatesSetupURL
	if request.ProviderStatesSetupURL == "" && len(request.StateHandlers) > 0 {
		setupURL = fmt.Sprintf("http://localhost:%d%s", port, providerStatesSetupPath)
	}

	// Construct verifier request
	verificationRequest := types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", port),
		PactURLs:                   request.PactURLs,
		BrokerURL:                  request.BrokerURL,
		Tags:                       request.Tags,
		BrokerUsername:             request.BrokerUsername,
		BrokerPassword:             request.BrokerPassword,
		BrokerToken:                request.BrokerToken,
		PublishVerificationResults: request.PublishVerificationResults,
		ProviderVersion:            request.ProviderVersion,
		Provider:                   request.Provider,
		ProviderStatesSetupURL:     setupURL,
		CustomProviderHeaders:      request.CustomProviderHeaders,
		ConsumerVersionSelectors:   request.ConsumerVersionSelectors,
		EnablePending:              request.EnablePending,
		ProviderTags:               request.ProviderTags,
		Verbose:                    request.Verbose,
		FailIfNoPactsFound:         request.FailIfNoPactsFound,
		IncludeWIPPactsSince:       request.IncludeWIPPactsSince,
	}

	if request.Provider == "" {
		verificationRequest.Provider = p.Provider
	}

	portErr := waitForPort(port, "tcp", "localhost", p.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for http verification proxy on port %d - check for errors`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return res, portErr
	}

	log.Println("[DEBUG] pact provider verification")

	return p.pactClient.VerifyProvider(verificationRequest)
}

// VerifyProvider accepts an instance of `*testing.T`
// running the provider verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
func (p *Pact) VerifyProvider(t *testing.T, request types.VerifyRequest) ([]types.ProviderVerifierResponse, error) {
	res, err := p.VerifyProviderRaw(request)

	if len(res) == 0 {
		var message = "no pacts found to verify"
		if err != nil {
			message = "error verifying the provider: see returned error for detail"
		}

		if len(request.Tags) > 0 {
			message = fmt.Sprintf("%s. Check the tags provided (%s) for your broker (%s) are correct", message, strings.Join(request.Tags, ","), request.BrokerURL)
		}

		if request.FailIfNoPactsFound {
			t.Errorf(message)
		} else {
			t.Logf(message)
		}
	}

	runTestCases(t, res)

	return res, err
}

var installer = install.NewInstaller()

var checkCliCompatibility = func() {
	log.Println("[DEBUG] checking CLI compatibility")
	err := installer.CheckInstallation()

	if err != nil {
		log.Fatal("[ERROR] CLI tools are out of date, please upgrade before continuing")
	}
}

// BeforeEachMiddleware is invoked before any other, only on the __setup
// request (to avoid duplication)
func BeforeEachMiddleware(BeforeEach types.Hook) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == providerStatesSetupPath {

				log.Println("[DEBUG] executing before hook")
				err := BeforeEach()

				if err != nil {
					log.Println("[ERROR] error executing before hook:", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AfterEachMiddleware is invoked after any other, and is the last
// function to be called prior to returning to the test suite. It is
// therefore not invoked on __setup
func AfterEachMiddleware(AfterEach types.Hook) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			if r.URL.Path != providerStatesSetupPath {
				log.Println("[DEBUG] executing after hook")
				err := AfterEach()

				if err != nil {
					log.Println("[ERROR] error executing after hook:", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}
		})
	}
}

// stateHandlerMiddleware responds to the various states that are
// given during provider verification
//
// statehandler accepts a state object from the verifier and executes
// any state handlers associated with the provider.
// It will not execute further middleware if it is the designted "state" request
func stateHandlerMiddleware(stateHandlers types.StateHandlers) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == providerStatesSetupPath {
				var s *types.ProviderState
				decoder := json.NewDecoder(r.Body)
				decoder.Decode(&s)

				// Setup any provider state
				for _, state := range s.States {
					sf, stateFound := stateHandlers[state]

					if !stateFound {
						log.Printf("[WARN] state handler not found for state: %v", state)
					} else {
						// Execute state handler
						if err := sf(); err != nil {
							log.Printf("[ERROR] state handler for '%v' errored: %v", state, err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
					}
				}

				w.WriteHeader(http.StatusOK)
				return
			}

			log.Println("[DEBUG] skipping state handler for request", r.RequestURI)

			// Pass through to application
			next.ServeHTTP(w, r)
		})
	}
}

var messageVerificationHandler = func(messageHandlers MessageHandlers, stateHandlers StateHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Extract message
		var message Message
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(body, &message)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Setup any provider state
		for _, state := range message.States {
			sf, stateFound := stateHandlers[state.Name]

			if !stateFound {
				log.Printf("[WARN] state handler not found for state: %v", state.Name)
			} else {
				// Execute state handler
				if err = sf(state); err != nil {
					log.Printf("[WARN] state handler for '%v' return error: %v", state.Name, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		// Lookup key in function mapping
		f, messageFound := messageHandlers[message.Description]

		if !messageFound {
			log.Printf("[ERROR] message handler not found for message description: %v", message.Description)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Execute function handler
		res, handlerErr := f(message)

		if handlerErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		wrappedResponse := map[string]interface{}{
			"contents": res,
		}

		// Write the body back
		resBody, errM := json.Marshal(wrappedResponse)
		if errM != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			log.Println("[ERROR] error marshalling objcet:", errM)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resBody)
	}
}

func generateTestCaseName(res types.ProviderVerifierResponse) string {
	if len(res.Examples) > 1 {
		return fmt.Sprintf("Pact between %s and %s %s", res.Examples[0].Pact.ConsumerName, res.Examples[0].Pact.ProviderName, res.Examples[0].Pact.ShortDescription)
	}
	return "Running pact test"
}

// VerifyMessageProvider accepts an instance of `*testing.T`
// running provider message verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
//
// A Message Producer is analogous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
func (p *Pact) VerifyMessageProvider(t *testing.T, request VerifyMessageRequest) (res []types.ProviderVerifierResponse, err error) {
	res, err = p.VerifyMessageProviderRaw(request)

	runTestCases(t, res)

	return
}

func runTestCases(t *testing.T, res []types.ProviderVerifierResponse) {
	for _, test := range res {
		t.Run(generateTestCaseName(test), func(pactTest *testing.T) {
			for _, notice := range test.Summary.Notices {
				if notice.When == "before_verification" {
					t.Logf("notice: %s", notice.Text)
				}
			}
			for _, example := range test.Examples {
				testCase := example.Description
				if example.Status == "pending" {
					testCase = fmt.Sprintf("Pending %s", example.Description)
				}

				t.Run(testCase, func(st *testing.T) {
					st.Log(example.FullDescription)

					if example.Status != "passed" {
						if example.Status == "pending" {
							st.Skip(example.Exception.Message)
						} else {
							st.Errorf("%s\n%s\n", example.FullDescription, example.Exception.Message)
						}
					}
				})
			}
			for _, notice := range test.Summary.Notices {
				if notice.When == "after_verification" {
					t.Logf("notice: %s", notice.Text)
				}
			}
		})
	}
}

// VerifyMessageProviderRaw runs provider message verification.
//
// A Message Producer is analogous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
func (p *Pact) VerifyMessageProviderRaw(request VerifyMessageRequest) ([]types.ProviderVerifierResponse, error) {
	p.Setup(false)
	response := make([]types.ProviderVerifierResponse, 0)

	// Starts the message wrapper API with hooks back to the message handlers
	// This maps the 'description' field of a message pact, to a function handler
	// that will implement the message producer. This function must return an object and optionally
	// and error. The object will be marshalled to JSON for comparison.
	mux := http.NewServeMux()

	port, err := utils.GetFreePort()
	if err != nil {
		return response, fmt.Errorf("unable to allocate a port for verification: %v", err)
	}

	// Construct verifier request
	verificationRequest := types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", port),
		PactURLs:                   request.PactURLs,
		BrokerURL:                  request.BrokerURL,
		Tags:                       request.Tags,
		BrokerUsername:             request.BrokerUsername,
		BrokerPassword:             request.BrokerPassword,
		BrokerToken:                request.BrokerToken,
		PublishVerificationResults: request.PublishVerificationResults,
		ProviderVersion:            request.ProviderVersion,
		ProviderTags:               request.ProviderTags,
		Provider:                   p.Provider,
	}

	mux.HandleFunc("/", messageVerificationHandler(request.MessageHandlers, request.StateHandlers))

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("[DEBUG] API handler starting: port %d (%s)", port, ln.Addr())
	go http.Serve(ln, mux)

	portErr := waitForPort(port, "tcp", "localhost", p.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for pact proxy on port %d - check for errors`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return response, portErr
	}

	log.Println("[DEBUG] pact provider verification")
	return p.pactClient.VerifyProvider(verificationRequest)
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
//
// A Message Consumer is analogous to a Provider in the HTTP Interaction model.
// It is the receiver of an interaction, and needs to be able to handle whatever
// request was provided.
func (p *Pact) VerifyMessageConsumerRaw(message *Message, handler MessageConsumer) error {
	log.Printf("[DEBUG] verify message")
	p.Setup(false)

	// Reify the message back to its "example/generated" form
	reified, err := p.pactClient.ReifyMessage(&types.PactReificationRequest{
		Message: message.Content,
	})

	if err != nil {
		return fmt.Errorf("unable to convert consumer test to a valid JSON representation: %v", err)
	}

	t := reflect.TypeOf(message.Type)
	if t != nil && t.Name() != "interface" {
		log.Println("[DEBUG] narrowing type to", t.Name())
		err = json.Unmarshal(reified.ResponseRaw, &message.Type)

		if err != nil {
			return fmt.Errorf("unable to narrow type to %v: %v", t.Name(), err)
		}
	}

	// Yield message, and send through handler function
	generatedMessage :=
		Message{
			Content:     message.Type,
			States:      message.States,
			Description: message.Description,
			Metadata:    message.Metadata,
		}

	err = handler(generatedMessage)
	if err != nil {
		return err
	}

	// If no errors, update Message Pact
	return p.pactClient.UpdateMessagePact(types.PactMessageRequest{
		Message:  message,
		Consumer: p.Consumer,
		Provider: p.Provider,
		PactDir:  p.PactDir,
	})
}

// VerifyMessageConsumer is a test convience function for VerifyMessageConsumerRaw,
// accepting an instance of `*testing.T`
func (p *Pact) VerifyMessageConsumer(t *testing.T, message *Message, handler MessageConsumer) error {
	err := p.VerifyMessageConsumerRaw(message, handler)

	if err != nil {
		t.Errorf("VerifyMessageConsumer failed: %v", err)
	}

	return err
}

const providerStatesSetupPath = "/__setup"
