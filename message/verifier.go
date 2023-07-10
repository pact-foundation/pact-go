package message

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
)

// VerifyMessageProviderRaw runs provider message verification.
//
// A Message Producer is analagous to Consumer in the HTTP Interaction model.
// It is the initiator of an interaction, and expects something on the other end
// of the interaction to respond - just in this case, not immediately.
// func (v *Verifier) verifyMessageProviderRaw(request VerifyMessageRequest, writer outputWriter) error {
// 	// request.handle = native.NewVerifier("pact-go", command.Version)

// 	err := v.validateConfig()

// 	if err != nil {
// 		return err
// 	}

// 	// Starts the message wrapper API with hooks back to the message handlers
// 	// This maps the 'description' field of a message pact, to a function handler
// 	// that will implement the message producer. This function must return an object and optionally
// 	// and error. The object will be marshalled to JSON for comparison.
// 	mux := http.NewServeMux()

// 	port, err := utils.GetFreePort()
// 	if err != nil {
// 		return fmt.Errorf("unable to allocate a port for verification: %v", err)
// 	}

// 	// Construct verifier request
// 	verificationRequest := provider.VerifyRequest{
// 		ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", port),
// 		PactURLs:                   request.PactURLs,
// 		PactFiles:                  request.PactFiles,
// 		PactDirs:                   request.PactDirs,
// 		BrokerURL:                  request.BrokerURL,
// 		Tags:                       request.Tags,
// 		BrokerUsername:             request.BrokerUsername,
// 		BrokerPassword:             request.BrokerPassword,
// 		BrokerToken:                request.BrokerToken,
// 		PublishVerificationResults: request.PublishVerificationResults,
// 		ProviderVersion:            request.ProviderVersion,
// 		Provider:                   request.Provider,
// 		ProviderTags:               request.ProviderTags,
// 		// CustomProviderHeaders:      request.CustomProviderHeaders,
// 		ConsumerVersionSelectors: request.ConsumerVersionSelectors,
// 		EnablePending:            request.EnablePending,
// 		FailIfNoPactsFound:       request.FailIfNoPactsFound,
// 		IncludeWIPPactsSince:     request.IncludeWIPPactsSince,
// 		ProviderStatesSetupURL:   fmt.Sprintf("http://localhost:%d%s", port, providerStatesSetupPath),
// 	}

// 	mux.HandleFunc(providerStatesSetupPath, messageStateHandler(request.StateHandlers))
// 	mux.HandleFunc("/", messageVerificationHandler(request.MessageHandlers))

// 	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer ln.Close()

// 	log.Printf("[DEBUG] API handler starting: port %d (%s)", port, ln.Addr())
// 	go http.Serve(ln, mux)

// 	portErr := provider.WaitForPort(port, "tcp", "localhost", v.ClientTimeout,
// 		fmt.Sprintf(`Timed out waiting for pact proxy on port %d - check for errors`, port))

// 	if portErr != nil {
// 		log.Fatal("Error:", err)
// 		return portErr
// 	}

// 	log.Println("[DEBUG] pact provider verification")

// 	return verificationRequest.Verify(writer)
// }

type messageVerificationHandlerRequest struct {
	Description string                 `json:"description"`
	States      []models.ProviderState `json:"providerStates"`
}

// type messageStateHandlerRequest struct {
// 	State  string                 `json:"state"`
// 	Params map[string]interface{} `json:"params"`
// 	Action string                 `json:"action"`
// }

// TODO: is this different to HTTP?
// var MessageStateHandler = func(stateHandlers models.StateHandlers) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 		log.Printf("[TRACE] message state handler %+v", r)

// 		// Extract message
// 		var message messageStateHandlerRequest
// 		body, err := ioutil.ReadAll(r.Body)
// 		r.Body.Close()
// 		log.Printf("[TRACE] message state handler received request: %s, %s", string(body), r.URL.Path)

// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		err = json.Unmarshal(body, &message)

// 		if err != nil {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		log.Println("[TRACE] message verification - setting up state for message", message)
// 		sf, stateFound := stateHandlers[message.State]

// 		if !stateFound {
// 			log.Printf("[WARN] state handler not found for state: %v", message.State)
// 		} else {
// 			// Execute state handler
// 			res, err := sf(message.Action == "setup", models.ProviderState{
// 				Name:       message.State,
// 				Parameters: message.Params,
// 			})

// 			if err != nil {
// 				log.Printf("[WARN] state handler for '%v' return error: %v", message.State, err)
// 				w.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}

// 			// Return provider state values for generator
// 			if res != nil {
// 				resBody, err := json.Marshal(res)

// 				if err != nil {
// 					log.Printf("[ERROR] state handler for '%v' errored: %v", message.State, err)
// 					w.WriteHeader(http.StatusInternalServerError)

// 					return
// 				}

// 				log.Printf("[INFO] state handler for '%v' finished", message.State)
// 				w.Write(resBody)
// 			}

// 		}

// 		// w.WriteHeader(http.StatusOK)
// 	}
// }

type messageWithMetadata struct {
	Contents []byte   `json:"pactMessageContents"`
	Metadata Metadata `json:"pactMessageMetadata"`
}

func appendMetadataToResponse(res interface{}, metadata Metadata) ([]byte, error) {
	data, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	withMetadata := &messageWithMetadata{
		Contents: data,
		Metadata: metadata,
	}

	return json.Marshal(withMetadata)
}

var PACT_MESSAGE_METADATA_HEADER = "PACT_MESSAGE_METADATA"
var PACT_MESSAGE_METADATA_HEADER2 = "Pact-Message-Metadata"

func appendMetadataToResponseHeaders(metadata Metadata, w http.ResponseWriter) {
	if len(metadata) > 0 {
		log.Println("[DEBUG] adding message metadata header", metadata)
		json, err := json.Marshal(metadata)
		if err != nil {
			log.Println("[WARN] invalid metadata", metadata, ". Unable to marshal to JSON:", err)
		}
		log.Println("[TRACE] encoded metadata to JSON:", string(json))

		encoded := base64.StdEncoding.EncodeToString(json)
		log.Println("[TRACE] encoded metadata to base64:", encoded)

		w.Header().Add(PACT_MESSAGE_METADATA_HEADER, encoded)
		w.Header().Add(PACT_MESSAGE_METADATA_HEADER2, encoded)

		// Content-Type must match the body content type in the pact.
		if metadata["contentType"] != nil {
			w.Header().Set("Content-Type", metadata["contentType"].(string))
		} else if metadata["content-type"] != nil {
			w.Header().Set("Content-Type", metadata["content-type"].(string))
		} else if metadata["Content-Type"] != nil {
			w.Header().Set("Content-Type", metadata["Content-Type"].(string))
		} else {
			defaultContentType := "application/json; charset=utf-8"
			log.Println("[WARN] no content type (key 'contentType') found in message metadata. Defaulting to", defaultContentType)
			w.Header().Set("Content-Type", defaultContentType)
		}
	}
}

func CreateMessageHandler(messageHandlers Handlers) proxy.Middleware {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/__messages" {
				log.Printf("[TRACE] message verification handler")

				// Extract message
				var message messageVerificationHandlerRequest
				body, err := ioutil.ReadAll(r.Body)
				r.Body.Close()
				log.Printf("[TRACE] message verification handler received request: %+s, %s", body, r.URL.Path)

				if err != nil {
					log.Printf("[ERROR] unable to parse message verification request: %s", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				err = json.Unmarshal(body, &message)

				if err != nil {
					log.Printf("[ERROR] unable to parse message verification request: %s", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// Lookup key in function mapping
				f, messageFound := messageHandlers[message.Description]

				if !messageFound {
					log.Printf("[ERROR] message handler not found for message description: %v", message.Description)
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Execute function handler
				res, metadata, handlerErr := f(message.States)

				if handlerErr != nil {
					log.Printf("[ERROR] error executive message handler %s", err)
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}

				// Write the body back
				appendMetadataToResponseHeaders(metadata, w)

				if bytes, ok := res.([]byte); ok {
					log.Println("[DEBUG] checking type of message is []byte")
					body = bytes
				} else {
					log.Println("[DEBUG] message body is not []byte, serialising as JSON")
					body, err = json.Marshal(res)
					if err != nil {
						w.WriteHeader(http.StatusServiceUnavailable)
						log.Println("[ERROR] error marshalling object:", err)
						return
					}
				}

				w.WriteHeader(http.StatusOK)
				w.Write(body)

				return
			}
			log.Println("[TRACE] skipping message handler for request", r.RequestURI)

			// Pass through to application
			next.ServeHTTP(w, r)

		})
	}
}

const providerStatesSetupPath = "/__setup"
