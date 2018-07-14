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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/pact-foundation/pact-go/install"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// Pact is the container structure to run the Consumer Pact test cases.
type Pact struct {
	// Current server for the consumer.
	Server *types.MockServer

	// Pact RPC Client.
	pactClient *PactClient

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
		p.pactClient = NewClient()
		p.pactClient.TimeoutDuration = p.ClientTimeout
	}

	if p.PactFileWriteMode == "" {
		p.PactFileWriteMode = "overwrite"
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
			Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
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

	// Check if we are verifying messages or if we actually have interactions
	if len(p.Interactions) == 0 {
		return errors.New("there are no interactions to be verified")
	}

	mockServer := &MockService{
		BaseURL:  fmt.Sprintf("http://%s:%d", p.Host, p.Server.Port),
		Consumer: p.Consumer,
		Provider: p.Provider,
	}

	for _, interaction := range p.Interactions {
		err := mockServer.AddInteraction(interaction)
		if err != nil {
			return err
		}
	}

	// Run the integration test
	err := integrationTest()
	if err != nil {
		return err
	}

	// Run Verification Process
	err = mockServer.Verify()
	if err != nil {
		return err
	}

	// Clear out interations
	p.Interactions = make([]*Interaction, 0)

	return mockServer.DeleteInteractions()
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
func (p *Pact) VerifyProviderRaw(request types.VerifyRequest) (types.ProviderVerifierResponse, error) {
	p.Setup(false)

	// If we provide a Broker, we go to it to find consumers
	if request.BrokerURL != "" {
		log.Println("[DEBUG] pact provider verification - finding all consumers from broker: ", request.BrokerURL)
		err := findConsumers(p.Provider, &request)
		if err != nil {
			return types.ProviderVerifierResponse{}, err
		}
	}

	log.Println("[DEBUG] pact provider verification")

	return p.pactClient.VerifyProvider(request)
}

// VerifyProvider accepts an instance of `*testing.T`
// running the provider verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
func (p *Pact) VerifyProvider(t *testing.T, request types.VerifyRequest) (types.ProviderVerifierResponse, error) {
	res, err := p.VerifyProviderRaw(request)

	for _, example := range res.Examples {
		t.Run(example.Description, func(st *testing.T) {
			st.Log(example.FullDescription)
			if example.Status != "passed" {
				t.Errorf("%s\n%s\n", example.FullDescription, example.Exception.Message)
			}
		})
	}

	return res, err
}

var installer = install.NewInstaller()

var checkCliCompatibility = func() {
	log.Println("[DEBUG] checking CLI compatability")
	err := installer.CheckInstallation()

	if err != nil {
		log.Fatal("[ERROR] CLI tools are out of date, please upgrade before continuing")
	}
}

var messageHandler = func(messageHandlers MessageHandlers, stateHandlers StateHandlers) http.HandlerFunc {
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

		json.Unmarshal(body, &message)

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
			fmt.Println("[ERROR] error marshalling objcet:", errM)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resBody)
	}
}

// VerifyMessageProvider accepts an instance of `*testing.T`
// running provider message verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
//
// A Message Producer is analagous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
func (p *Pact) VerifyMessageProvider(t *testing.T, request VerifyMessageRequest) (res types.ProviderVerifierResponse, err error) {
	res, err = p.VerifyMessageProviderRaw(request)

	for _, example := range res.Examples {
		t.Run(example.Description, func(st *testing.T) {
			st.Log(example.FullDescription)
			if example.Status != "passed" {
				st.Errorf("%s\n", example.Exception.Message)
				st.Error("Check to ensure that all message expectations have corresponding message handlers")
			}
		})
	}

	return
}

// VerifyMessageProviderRaw runs provider message verification.
//
// A Message Producer is analagous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
func (p *Pact) VerifyMessageProviderRaw(request VerifyMessageRequest) (types.ProviderVerifierResponse, error) {
	p.Setup(false)
	response := types.ProviderVerifierResponse{}

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
		PublishVerificationResults: request.PublishVerificationResults,
		ProviderVersion:            request.ProviderVersion,
	}

	mux.HandleFunc("/", messageHandler(request.MessageHandlers, request.StateHandlers))

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("[DEBUG] API handler starting: port %d (%s)", port, ln.Addr())
	go http.Serve(ln, mux)

	portErr := waitForPort(port, "tcp", "localhost", p.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for Daemon on port %d - are you sure it's running?`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return response, portErr
	}

	return p.VerifyProviderRaw(verificationRequest)
}

// VerifyMessageConsumerRaw creates a new Pact _message_ interaction to build a testable
// interaction.
//
//
// A Message Consumer is analagous to a Provider in the HTTP Interaction model.
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
		Message:           message,
		Consumer:          p.Consumer,
		Provider:          p.Provider,
		PactFileWriteMode: p.PactFileWriteMode,
		PactDir:           p.PactDir,
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
