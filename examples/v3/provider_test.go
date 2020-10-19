// +build provider

// Package main contains a runnable Provider Pact test example.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	v3 "github.com/pact-foundation/pact-go/v3"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)

// Example Provider Pact: How to run me!
// 1. cd <pact-go>/examples/v3
// 2. go test -v -run TestProvider
func TestV3HTTPProvider(t *testing.T) {
	v3.SetLogLevel("INFO")

	// Start provider API in the background
	go startServer()

	verifier := v3.HTTPVerifier{}

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
	err := verifier.VerifyProvider(t, v3.VerifyRequest{
		ProviderBaseURL: "http://localhost:8000",
		PactFiles:       []string{filepath.ToSlash(fmt.Sprintf("%s/V3Consumer-V3Provider.json", pactDir))},
		RequestFilter:   f,
		StateHandlers: v3.StateHandlers{
			"User foo exists": func(s v3.ProviderStateV3) error {
				log.Println("[DEBUG] calling user foo exists state handler")

				// ... do something

				return nil
			},
		},
	})

	assert.NoError(t, err)
}

func TestV3MessageProvider(t *testing.T) {
	v3.SetLogLevel("TRACE")
	var user *User

	verifier := v3.MessageVerifier{}

	// Map test descriptions to message producer (handlers)
	functionMappings := v3.MessageHandlers{
		"a user event": func([]v3.ProviderStateV3) (interface{}, error) {
			if user != nil {
				return user, nil
			} else {
				return map[string]string{
					"message": "not found",
				}, nil
			}
		},
	}

	stateMappings := v3.StateHandlers{
		"User with id 127 exists": func(v3.ProviderStateV3) error {
			user = &User{
				ID:       44,
				Name:     "Baz",
				Date:     "2020-01-01",
				LastName: "sampson",
			}

			return nil
		},
	}

	// Verify the Provider with local Pact Files
	verifier.Verify(t, v3.VerifyMessageRequest{
		VerifyRequest: v3.VerifyRequest{
			PactFiles:     []string{filepath.ToSlash(fmt.Sprintf("%s/V3MessageConsumer-V3MessageProvider.json", pactDir))},
			StateHandlers: stateMappings,
		},
		MessageHandlers: functionMappings,
	})
}

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
			{
				"accountBalance": 123.76,
				"dateTime": "2020-01-01",
				"equality": "a thing",
				"id": 12,
				"itemsMin": [
					"thereshouldbe3ofthese",
					"thereshouldbe3ofthese",
					"thereshouldbe3ofthese"
				],
				"itemsMinMax": [
					27,
					27,
					27,
					27
				],
				"lastName": "LastName",
				"name": "FirstName",
				"superstring": "foo"
			}`,
		)
	})

	log.Fatal(http.ListenAndServe("localhost:8000", mux))
}

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=billy"`
	LastName string `json:"lastName" pact:"example=sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
	// Date     string `json:"datetime" pact:"example=20200101,regex=[0-9a-z-A-Z]+"`
}
