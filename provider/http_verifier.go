package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
)

// HTTPVerifier is used to verify the provider side of an HTTP API contract
type HTTPVerifier struct {
	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration
}

func (v *HTTPVerifier) validateConfig() error {
	if v.ClientTimeout == 0 {
		v.ClientTimeout = 10 * time.Second
	}

	return nil
}

// VerifyProviderRaw reads the provided pact files and runs verification against
// a running Provider API, providing raw response from the Verification process.
//
// Order of events: BeforeEach, stateHandlers, requestFilter(pre <execute provider> post), AfterEach
func (v *HTTPVerifier) verifyProviderRaw(request VerifyRequest, writer outputWriter) error {
	err := v.validateConfig()

	if err != nil {
		return err
	}

	u, err := url.Parse(request.ProviderBaseURL)

	m := []proxy.Middleware{}

	if request.BeforeEach != nil {
		m = append(m, beforeEachMiddleware(request.BeforeEach))
	}

	if request.AfterEach != nil {
		m = append(m, afterEachMiddleware(request.AfterEach))
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
	verificationRequest := VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", port),
		PactURLs:                   request.PactURLs,
		PactFiles:                  request.PactFiles,
		PactDirs:                   request.PactDirs,
		BrokerURL:                  request.BrokerURL,
		Tags:                       request.Tags,
		BrokerUsername:             request.BrokerUsername,
		BrokerPassword:             request.BrokerPassword,
		BrokerToken:                request.BrokerToken,
		PublishVerificationResults: request.PublishVerificationResults,
		ProviderVersion:            request.ProviderVersion,
		Provider:                   request.Provider,
		ProviderStatesSetupURL:     setupURL,
		ProviderTags:               request.ProviderTags,
		// CustomProviderHeaders:      request.CustomProviderHeaders,
		ConsumerVersionSelectors: request.ConsumerVersionSelectors,
		EnablePending:            request.EnablePending,
		FailIfNoPactsFound:       request.FailIfNoPactsFound,
		IncludeWIPPactsSince:     request.IncludeWIPPactsSince,
	}

	portErr := WaitForPort(port, "tcp", "localhost", v.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for http verification proxy on port %d - check for errors`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return portErr
	}

	log.Println("[DEBUG] pact provider verification")

	return verificationRequest.Verify(writer)
}

// VerifyProvider accepts an instance of `*testing.T`
// running the provider verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
func (v *HTTPVerifier) VerifyProvider(t *testing.T, request VerifyRequest) error {
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

// afterEachMiddleware is invoked after any other, and is the last
// function to be called prior to returning to the test suite. It is
// therefore not invoked on __setup
func afterEachMiddleware(AfterEach Hook) proxy.Middleware {
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

// {"action":"teardown","id":"foo","state":"User foo exists"}
type stateHandlerAction struct {
	Action string `json:"action"`
	State  string `json:"state"`
	Params map[string]interface{}
}

// stateHandlerMiddleware responds to the various states that are
// given during provider verification
//
// statehandler accepts a state object from the verifier and executes
// any state handlers associated with the provider.
// It will not execute further middleware if it is the designted "state" request
func stateHandlerMiddleware(stateHandlers models.StateHandlers) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == providerStatesSetupPath {
				log.Println("[INFO] executing state handler middleware")
				var state stateHandlerAction
				buf := new(strings.Builder)
				io.Copy(buf, r.Body)
				log.Println("[TRACE] state handler received raw input", buf.String())

				err := json.Unmarshal([]byte(buf.String()), &state)
				log.Println("[TRACE] state handler parsed input (without params)", state)

				if err != nil {
					log.Println("[ERROR] unable to decode incoming state change payload", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Extract the params from the payload. They are in the root, so we need to do some more work to achieve this
				var params models.ProviderStateV3Response
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
					res, err := sf(state.Action == "setup", models.ProviderStateV3{Name: state.State, Parameters: state.Params})

					if err != nil {
						log.Printf("[ERROR] state handler for '%v' errored: %v", state.State, err)
						w.WriteHeader(http.StatusInternalServerError)
						return
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
						w.Write(resBody)
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

const providerStatesSetupPath = "/__setup/"
