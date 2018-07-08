// Package main contains a runnable Provider Pact test example.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)

// Example Provider Pact: How to run me!
// 1. cd <pact-go>/examples
// 2. go test -v -run TestProvider
func TestProvider(t *testing.T) {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "MyConsumer",
		Provider: "MyProvider",
	}

	// Start provider API in the background
	go startServer()

	// Verify the Provider with local Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        "http://localhost:8000",
		PactURLs:               []string{filepath.ToSlash(fmt.Sprintf("%s/myconsumer-myprovider.json", pactDir))},
		ProviderStatesSetupURL: "http://localhost:8000/setup",
		CustomProviderHeaders:  []string{"Authorization: basic e5e5e5e5e5e5e5"},
	})
}

func startServer() {
	mux := http.NewServeMux()
	lastName := "billy"

	mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, fmt.Sprintf(`{"name":"billy", "lastName":"%s"}`, lastName))

		// Break the API by replacing the above and uncommenting one of these
		// w.WriteHeader(http.StatusUnauthorized)
		// fmt.Fprintf(w, `{"s":"baz"}`)
	})

	// This function handles state requests for a particular test
	// In this case, we ensure that the user being requested is available
	// before the Verification process invokes the API.
	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		var s *types.ProviderState
		decoder := json.NewDecoder(req.Body)
		decoder.Decode(&s)
		if s.State == "User foo exists" {
			lastName = "bar"
		}

		w.Header().Add("Content-Type", "application/json")
	})
	log.Fatal(http.ListenAndServe(":8000", mux))
}
