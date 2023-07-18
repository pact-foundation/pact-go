//go:build provider
// +build provider

package avro

import (
	"fmt"
	"log"
	"net/http"

	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)

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
			filepath.ToSlash(fmt.Sprintf("%s/AvroConsumer-AvroProvider.json", pactDir)),
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
		binary, err := codec.BinaryFromNative(nil, map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
		})
		if err != nil {
			log.Println("ERROR: ", err)
			w.WriteHeader(500)
		} else {
			fmt.Fprintf(w, string(binary))
			w.WriteHeader(200)
		}
	})

	log.Printf("started HTTP server on port: %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux))
}
