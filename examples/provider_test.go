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

	. "github.com/pact-foundation/pact-go/v2/sugar"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)

// Example Provider Pact: How to run me!
// 1. cd <pact-go>/examples/v3
// 2. go test -v -run TestProvider
func TestV3HTTPProvider(t *testing.T) {
	SetLogLevel("TRACE")
	CheckVersion()

	// Start provider API in the background
	go startServer()

	verifier := HTTPVerifier{}

	// Authorization middleware
	// This is your chance to modify the request before it hits your provider
	// NOTE: this should be used very carefully, as it has the potential to
	// _change_ the contract
	f := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("[DEBUG] HOOK request filter")
			r.Header.Add("Authorization", "Bearer 1234-dynamic-value")
			next.ServeHTTP(w, r)
		})
	}

	// Verify the Provider with local Pact Files
	err := verifier.VerifyProvider(t, VerifyRequest{
		ProviderBaseURL: "http://localhost:8111",
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/V3Consumer-V3Provider.json", pactDir)),
			filepath.ToSlash(fmt.Sprintf("%s/V2ConsumerMatch-V2ProviderMatch.json", pactDir)),
		},
		RequestFilter: f,
		BeforeEach: func() error {
			log.Println("[DEBUG] HOOK before each")
			return nil
		},
		AfterEach: func() error {
			log.Println("[DEBUG] HOOK after each")
			return nil
		},
		StateHandlers: StateHandlers{
			"User foo exists": func(setup bool, s ProviderStateV3) (ProviderStateV3Response, error) {

				if setup {
					log.Println("[DEBUG] HOOK calling user foo exists state handler", s)
				} else {
					log.Println("[DEBUG] HOOK teardown the 'User foo exists' state")
				}

				// ... do something, such as create "foo" in the database

				// Optionally (if there are generators in the pact) return provider state values to be used in the verification
				return ProviderStateV3Response{"uuid": "1234"}, nil
			},
		},
	})

	assert.NoError(t, err)
}

func TestV3MessageProvider(t *testing.T) {
	SetLogLevel("TRACE")
	var user *User

	verifier := MessageVerifier{}

	// Map test descriptions to message producer (handlers)
	functionMappings := MessageHandlers{
		"a user event": func([]ProviderStateV3) (interface{}, error) {
			if user != nil {
				return user, nil
			} else {
				return ProviderStateV3Response{
					"message": "not found",
				}, nil
			}
		},
	}

	stateMappings := StateHandlers{
		"User with id 127 exists": func(setup bool, s ProviderStateV3) (ProviderStateV3Response, error) {
			if setup {
				user = &User{
					ID:       127,
					Name:     "Billy",
					Date:     "2020-01-01",
					LastName: "Sampson",
				}
			}

			return ProviderStateV3Response{"id": user.ID}, nil
		},
	}

	// Verify the Provider with local Pact Files
	verifier.Verify(t, VerifyMessageRequest{
		VerifyRequest: VerifyRequest{
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
				"datetime": "2020-01-01",
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
					27,
					27
				],
				"lastName": "Sampson",
				"name": "Billy",
				"superstring": "foo",
				"arrayContaining": [
					"string",
					1,
					{
						"foo": "bar"
					}
				]
			}`,
		)
	})

	log.Fatal(http.ListenAndServe("localhost:8111", mux))
}

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=billy"`
	LastName string `json:"lastName" pact:"example=Sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
	// Date     string `json:"datetime" pact:"example=20200101,regex=[0-9a-z-A-Z]+"`
}
