package message

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/proxy"
)

type messageVerificationHandlerRequest struct {
	Description string                 `json:"description"`
	States      []models.ProviderState `json:"providerStates"`
}

var (
	PACT_MESSAGE_METADATA_HEADER  = "PACT_MESSAGE_METADATA"
	PACT_MESSAGE_METADATA_HEADER2 = "Pact-Message-Metadata"
)

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
		contentType, ok := stringMetadataValue(metadata, "contentType", "content-type", "Content-Type")
		if !ok {
			contentType = "application/json; charset=utf-8"
			log.Println("[WARN] no content type (key 'contentType') found in message metadata. Defaulting to", contentType)
		}
		w.Header().Set("Content-Type", contentType)
	}
}

// stringMetadataValue returns the first string value found in metadata for
// the given keys, in order. The second return value is false if none of the
// keys are present or the value found is not a string.
func stringMetadataValue(metadata Metadata, keys ...string) (string, bool) {
	for _, key := range keys {
		v, present := metadata[key]
		if !present || v == nil {
			continue
		}
		s, ok := v.(string)
		if !ok {
			log.Printf("[WARN] message metadata key %q is not a string (got %T), ignoring", key, v)
			continue
		}
		return s, true
	}
	return "", false
}

func CreateMessageHandler(messageHandlers Handlers) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/__messages" {
				log.Printf("[TRACE] message verification handler")

				// Extract message
				var message messageVerificationHandlerRequest
				body, err := io.ReadAll(r.Body)
				closeErr := r.Body.Close()
				if closeErr != nil {
					log.Println("[WARN] failed to close request body:", closeErr)
				}
				log.Printf("[TRACE] message verification handler received request: %s, %s", strconv.Quote(string(body)), strconv.Quote(r.URL.Path))

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
					log.Printf("[ERROR] error executive message handler %s", handlerErr)
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
			log.Println("[TRACE] skipping message handler for request", strconv.Quote(r.RequestURI))

			// Pass through to application
			next.ServeHTTP(w, r)
		})
	}
}
