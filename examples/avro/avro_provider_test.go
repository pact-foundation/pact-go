//go:build provider

package avro

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	"github.com/stretchr/testify/assert"
)

var pactDir = dir + "/../pacts"

func TestAvroHTTPProvider(t *testing.T) {
	httpPort, _ := utils.GetFreePort()

	// Start provider API in the background
	go startHTTPProvider(httpPort)

	verifier := provider.NewVerifier()

	// Verify the Provider with local Pact Files
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("http://127.0.0.1:%d", httpPort),
		Provider:        "AvroProvider",
		PactFiles: []string{
			filepath.ToSlash(pactDir + "/AvroConsumer-AvroProvider.json"),
		},
	})

	assert.NoError(t, err)
}

func startHTTPProvider(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/avro", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "avro/binary;record=User")

		user := &User{
			ID:       1,
			Username: "matt",
			// Username: "sally", // matching rules not supported?
		}

		codec := getCodec()
		binary, err := codec.BinaryFromNative(nil, map[string]any{
			"id":       user.ID,
			"username": user.Username,
		})
		if err != nil {
			log.Println("ERROR: ", err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			_, writeErr := w.Write(binary)
			if writeErr != nil {
				log.Println("ERROR writing response body: ", writeErr)
			}
		}
	})

	log.Printf("started HTTP server on port: %d\n", port)
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
