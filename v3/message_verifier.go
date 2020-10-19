package v3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/utils"
)

type MessageVerifier struct {
	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration
}

func (v *MessageVerifier) validateConfig() error {
	if v.ClientTimeout == 0 {
		v.ClientTimeout = 10 * time.Second
	}

	return nil
}
func (v *MessageVerifier) Verify(t *testing.T, request VerifyMessageRequest) error {
	err := v.verifyMessageProviderRaw(request, t)

	t.Run("Provider pact verification", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
	})

	return err
}

// VerifyMessageProviderRaw runs provider message verification.
//
// A Message Producer is analagous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
func (v *MessageVerifier) verifyMessageProviderRaw(request VerifyMessageRequest, writer outputWriter) error {
	err := v.validateConfig()

	if err != nil {
		return err
	}
	// Starts the message wrapper API with hooks back to the message handlers
	// This maps the 'description' field of a message pact, to a function handler
	// that will implement the message producer. This function must return an object and optionally
	// and error. The object will be marshalled to JSON for comparison.
	mux := http.NewServeMux()

	port, err := utils.GetFreePort()
	if err != nil {
		return fmt.Errorf("unable to allocate a port for verification: %v", err)
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
		ProviderTags:               request.ProviderTags,
		// CustomProviderHeaders:      request.CustomProviderHeaders,
		// ConsumerVersionSelectors:   request.ConsumerVersionSelectors,
		// EnablePending:              request.EnablePending,
		FailIfNoPactsFound: request.FailIfNoPactsFound,
		// IncludeWIPPactsSince:       request.IncludeWIPPactsSince,
	}

	mux.HandleFunc("/", messageVerificationHandler(request.MessageHandlers, request.StateHandlers))

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("[DEBUG] API handler starting: port %d (%s)", port, ln.Addr())
	go http.Serve(ln, mux)

	portErr := waitForPort(port, "tcp", "localhost", v.ClientTimeout,
		fmt.Sprintf(`Timed out waiting for pact proxy on port %d - check for errors`, port))

	if portErr != nil {
		log.Fatal("Error:", err)
		return portErr
	}

	log.Println("[DEBUG] pact provider verification")

	return verificationRequest.verify(writer)
}

type messageVerificationHandlerRequest struct {
	Description string            `json:"description"`
	States      []ProviderStateV3 `json:"providerStates"`
}

var messageVerificationHandler = func(messageHandlers MessageHandlers, stateHandlers StateHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		log.Printf("[TRACE] message verification handler")

		// Extract message
		var message messageVerificationHandlerRequest
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		log.Printf("[TRACE] message verification handler received request: %+s", body)

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
		res, handlerErr := f(message.States)

		if handlerErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Write the body back
		resBody, errM := json.Marshal(res)
		if errM != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			log.Println("[ERROR] error marshalling objcet:", errM)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(resBody)
	}
}
