package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/command"
	"github.com/pact-foundation/pact-go/v2/internal/native"
	logging "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
	"github.com/pact-foundation/pact-go/v2/utils"
)

const MESSAGE_PATH = "/__messages"

// Verifier is used to verify the provider side of an HTTP API contract
type Verifier struct {
	// ClientTimeout specifies how long to wait for the provider to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration

	// Hostname to run any
	Hostname string

	handle *native.Verifier
}

func NewVerifier() *Verifier {
	native.Init(string(logging.LogLevel()))

	return &Verifier{
		handle: native.NewVerifier("pact-go", command.Version),
	}

}

func (v *Verifier) validateConfig() error {
	if v.ClientTimeout == 0 {
		v.ClientTimeout = 10 * time.Second
	}
	if v.Hostname == "" {
		v.Hostname = "localhost"
	}

	return nil
}

// If no HTTP server is given, we must start one up in order
// to provide a target for state changes etc.
func (v *Verifier) startDefaultHTTPServer(port int) {
	mux := http.NewServeMux()

	_ = http.ListenAndServe(fmt.Sprintf("%s:%d", v.Hostname, port), mux)
}

// VerifyProviderRaw reads the provided pact files and runs verification against
// a running Provider API, providing raw response from the Verification process.
//
// Order of events: BeforeEach, stateHandlers, requestFilter(pre <execute provider> post), AfterEach
func (v *Verifier) verifyProviderRaw(request VerifyRequest, writer outputWriter) error {

	// proxy target
	var u *url.URL

	err := v.validateConfig()
	if err != nil {
		return err
	}

	// Check if a provider has been given. If none, start a dummy service to attach the proxy to
	if request.ProviderBaseURL == "" {
		log.Println("[DEBUG] setting up a dummy server for verification, as none was provided")
		port, err := utils.GetFreePort()
		if err != nil {
			log.Panic("unable to allocate a port for verification:", err)
		}
		go v.startDefaultHTTPServer(port)

		request.ProviderBaseURL = fmt.Sprintf("http://localhost:%d", port)
	}

	u, err = url.Parse(request.ProviderBaseURL)
	if err != nil {
		log.Panic("unable to parse the provider URL", err)
	}

	m := []proxy.Middleware{}

	if request.BeforeEach != nil {
		m = append(m, beforeEachMiddleware(request.BeforeEach))
	}

	if len(request.StateHandlers) > 0 {
		m = append(m, stateHandlerMiddleware(request.StateHandlers, request.AfterEach))
	}

	if len(request.MessageHandlers) > 0 {
		m = append(m, message.CreateMessageHandler(request.MessageHandlers))
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

	if err != nil {
		return err
	}

	// Add any message targets
	if len(request.MessageHandlers) > 0 {
		request.Transports = append(request.Transports, Transport{
			Path:     MESSAGE_PATH,
			Protocol: "message",
			Port:     uint16(port),
		})
	}

	// Backwards compatibility, setup old provider states URL if given
	// Otherwise point to proxy
	if request.ProviderStatesSetupURL == "" && len(request.StateHandlers) > 0 {
		request.ProviderStatesSetupURL = fmt.Sprintf("http://localhost:%d%s", port, providerStatesSetupPath)
	}

	// Provider target should be the proxy
	request.ProviderBaseURL = fmt.Sprintf("http://localhost:%d", port)

	err = request.validate(v.handle)
	if err != nil {
		return err
	}

	portErr := WaitForPort(port, "tcp", "localhost", v.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for http verification proxy on port %d - check for errors`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return portErr
	}

	log.Println("[DEBUG] pact provider verification")

	return request.Verify(v.handle, writer)
}

// VerifyProvider accepts an instance of `*testing.T`
// running the provider verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
func (v *Verifier) VerifyProvider(t *testing.T, request VerifyRequest) error {
	err := v.verifyProviderRaw(request, t)

	// TODO: granular test reporting
	// runTestCases(t, res)

	t.Run("Provider pact verification", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
	})

	return err
}

// beforeEachMiddleware is invoked before any other, only on the __setup
// request (to avoid duplication)
func beforeEachMiddleware(BeforeEach Hook) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == providerStatesSetupPath {
				state, err := getStateFromRequest(r)

				// Before each should only fire on the "setup" phase
				if err == nil && state.Action == "setup" {
					log.Println("[DEBUG] executing before hook")
					err := BeforeEach()

					if err != nil {
						log.Println("[ERROR] error executing before hook:", err)
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// {"action":"teardown","id":"foo","state":"User foo exists"}
type stateHandlerAction struct {
	Action string `json:"action"`
	State  string `json:"state"`
	Params map[string]interface{}
}

func getStateFromRequest(r *http.Request) (stateHandlerAction, error) {
	var state stateHandlerAction
	buf := new(strings.Builder)
	tr := io.TeeReader(r.Body, buf)

	_, err := io.ReadAll(tr)
	if err != nil {
		log.Println("[ERROR] getStateFromRequest unable to read request body:", err)
		return stateHandlerAction{}, err
	}

	// Body is consumed above, need to put it back after ;P
	r.Body = ioutil.NopCloser(strings.NewReader(buf.String()))
	log.Println("[TRACE] getStateFromRequest received raw input", buf.String())

	err = json.Unmarshal([]byte(buf.String()), &state)
	log.Println("[TRACE] getStateFromRequest parsed input (without params)", state)

	if err != nil {
		log.Println("[ERROR] getStateFromRequest unable to decode incoming state change payload", err)
		return stateHandlerAction{}, err
	}

	return state, nil
}

// stateHandlerMiddleware responds to the various states that are
// given during provider verification
//
// statehandler accepts a state object from the verifier and executes
// any state handlers associated with the provider.
// It will not execute further middleware if it is the designted "state" request
func stateHandlerMiddleware(stateHandlers models.StateHandlers, afterEach Hook) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == providerStatesSetupPath {
				log.Println("[INFO] executing state handler middleware")
				var state stateHandlerAction
				buf := new(strings.Builder)
				tr := io.TeeReader(r.Body, buf)
				// TODO: should return an error if unable to read ?
				_, _ = io.ReadAll(tr)

				// Body is consumed above, need to put it back after ;P
				r.Body = ioutil.NopCloser(strings.NewReader(buf.String()))
				log.Println("[TRACE] state handler received raw input", buf.String())

				err := json.Unmarshal([]byte(buf.String()), &state)
				log.Println("[TRACE] state handler parsed input (without params)", state)

				if err != nil {
					log.Println("[ERROR] unable to decode incoming state change payload", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Extract the params from the payload. They are in the root, so we need to do some more work to achieve this
				var params models.ProviderStateResponse
				err = json.Unmarshal([]byte(buf.String()), &params)
				if err != nil {
					log.Println("[ERROR] unable to decode incoming state change payload", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// TODO: update rust code - params should be in a sub-property, to avoid top-level key conflicts
				// i.e. it's possible action/state are actually something a users wants to supply
				delete(params, "action")
				delete(params, "state")
				state.Params = params
				log.Println("[TRACE] state handler completed parsing input (with params)", state)

				// Find the provider state handler
				sf, stateFound := stateHandlers[state.State]

				if !stateFound {
					log.Printf("[WARN] no state handler found for state: %v", state.State)
				} else {
					// Execute state handler
					res, err := sf(state.Action == "setup", models.ProviderState{Name: state.State, Parameters: state.Params})

					if err != nil {
						log.Printf("[ERROR] state handler for '%v' errored: %v", state.State, err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					if state.Action == "teardown" && afterEach != nil {
						err := afterEach()

						if err != nil {
							log.Printf("[ERROR] after each hook for test errored: %v", err)
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
					}

					// Return provider state values for generator
					if res != nil {
						log.Println("[TRACE] returning values from provider state (raw)", res)
						resBody, err := json.Marshal(res)
						log.Println("[TRACE] returning values from provider state (JSON)", string(resBody))

						if err != nil {
							log.Printf("[ERROR] state handler for '%v' errored: %v", state.State, err)
							w.WriteHeader(http.StatusInternalServerError)

							return
						}

						w.Header().Add("content-type", "application/json")
						w.WriteHeader(http.StatusOK)
						// TODO: return interal server error if unable to write ?
						_, _ = w.Write(resBody)
						return
					}
				}

				w.WriteHeader(http.StatusOK)
				return
			}

			log.Println("[TRACE] skipping state handler for request", r.RequestURI)

			// Pass through to application
			next.ServeHTTP(w, r)
		})
	}
}

// Use this to wait for a port to be running prior
// to running tests.
func WaitForPort(port int, network string, address string, timeoutDuration time.Duration, message string) error {
	log.Println("[DEBUG] waiting for port", port, "to become available")
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-timeout:
			log.Printf("[ERROR] expected server to start < %s. %s", timeoutDuration, message)
			return fmt.Errorf("expected server to start < %s. %s", timeoutDuration, message)
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial(network, fmt.Sprintf("%s:%d", address, port))
			if err == nil {
				return nil
			}
		}
	}
}

const providerStatesSetupPath = "/__setup"
