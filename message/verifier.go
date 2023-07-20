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

type messageVerificationHandlerRequest struct {
	Description string                 `json:"description"`
	States      []models.ProviderState `json:"providerStates"`
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
				_, err = w.Write(body)
				if err != nil {
					log.Println("[ERROR] failed to write body response:", err)
				}

				return
			}
			log.Println("[TRACE] skipping message handler for request", r.RequestURI)

			// Pass through to application
			next.ServeHTTP(w, r)

		})
	}
}
