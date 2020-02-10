// Package main contains a runnable Provider Pact test example.
package main

import (
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

	// Authorization middleware
	// This is your chance to modify the request before it hits your provider
	// NOTE: this should be used very carefully, as it has the potential to
	// _change_ the contract
	f := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("Authorization", "Bearer 1234-dynamic-value")
			next.ServeHTTP(w, r)
		})
	}

	// Verify the Provider with local Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:       "http://localhost:8000",
		PactURLs:              []string{filepath.ToSlash(fmt.Sprintf("%s/myconsumer-myprovider.json", pactDir))},
		CustomProviderHeaders: []string{"X-API-Token: abcd"},
		RequestFilter:         f,
		StateHandlers: types.StateHandlers{
			"User foo exists": func() error {
				name = "billy"
				return nil
			},
		},
	})
}

var name = "some other name"

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, fmt.Sprintf(`{"name":"%s", "lastName": "jones"}`, name))

		// Break the API by replacing the above and uncommenting one of these
		// w.WriteHeader(http.StatusUnauthorized)
		// fmt.Fprintf(w, `{"s":"baz"}`)
	})

	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}
