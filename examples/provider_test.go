//go:build provider
// +build provider

// Package main contains a runnable Provider Pact test example.
package main

import (
	"fmt"
	l "log"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/version"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/pacts", dir)

var requestFilterCalled = false
var stateHandlerCalled = false

func TestV3HTTPProvider(t *testing.T) {
	log.SetLogLevel("TRACE")
	version.CheckVersion()

	// Start provider API in the background
	go startServer()

	verifier := provider.NewVerifier()

	// Authorization middleware
	// This is your chance to modify the request before it hits your provider
	// NOTE: this should be used very carefully, as it has the potential to
	// _change_ the contract
	f := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Println("[DEBUG] HOOK request filter")
			requestFilterCalled = true
			r.Header.Add("Authorization", "Bearer 1234-dynamic-value")
			next.ServeHTTP(w, r)
		})
	}

	// Verify the Provider with local Pact Files
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		ProviderBaseURL: "http://127.0.0.1:8111",
		Provider:        "V3Provider",
		ProviderVersion: os.Getenv("APP_SHA"),
		BrokerURL:       os.Getenv("PACT_BROKER_BASE_URL"),
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/PactGoV3Consumer-V3Provider.json", pactDir)),
			filepath.ToSlash(fmt.Sprintf("%s/PactGoV2ConsumerMatch-V2ProviderMatch.json", pactDir)),
		},
		ConsumerVersionSelectors: []provider.Selector{
			&provider.ConsumerVersionSelector{
				Tag: "master",
			},
			&provider.ConsumerVersionSelector{
				Tag: "prod",
			},
		},
		PublishVerificationResults: true,
		RequestFilter:              f,
		BeforeEach: func() error {
			l.Println("[DEBUG] HOOK before each")
			return nil
		},
		AfterEach: func() error {
			l.Println("[DEBUG] HOOK after each")
			return nil
		},
		StateHandlers: models.StateHandlers{
			"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
				stateHandlerCalled = true

				if setup {
					l.Println("[DEBUG] HOOK calling user foo exists state handler", s)
				} else {
					l.Println("[DEBUG] HOOK teardown the 'User foo exists' state")
				}

				// ... do something, such as create "foo" in the database

				// Optionally (if there are generators in the pact) return provider state values to be used in the verification
				return models.ProviderStateResponse{"uuid": "1234"}, nil
			},
		},
		DisableColoredOutput: true,
	})

	assert.NoError(t, err)
	assert.True(t, requestFilterCalled)
	assert.True(t, stateHandlerCalled)
}

func TestV3MessageProvider(t *testing.T) {
	log.SetLogLevel("TRACE")
	var user *User

	verifier := provider.NewVerifier()

	// Map test descriptions to message producer (handlers)
	functionMappings := message.Handlers{
		"a user event": func([]models.ProviderState) (message.Body, message.Metadata, error) {
			if user != nil {
				return user, message.Metadata{
					"Content-Type": "application/json",
				}, nil
			} else {
				return models.ProviderStateResponse{
					"message": "not found",
				}, nil, nil
			}
		},
	}

	// Setup any required states for the handlers
	stateMappings := models.StateHandlers{
		"User with id 127 exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			if setup {
				user = &User{
					ID:       127,
					Name:     "Billy",
					Date:     "2020-01-01",
					LastName: "Sampson",
				}
			}

			return models.ProviderStateResponse{"id": user.ID}, nil
		},
	}

	// Verify the Provider with local Pact Files
	verifier.VerifyProvider(t, provider.VerifyRequest{
		PactFiles:       []string{filepath.ToSlash(fmt.Sprintf("%s/PactGoV3MessageConsumer-V3MessageProvider.json", pactDir))},
		StateHandlers:   stateMappings,
		Provider:        "V3MessageProvider",
		ProviderVersion: os.Getenv("APP_SHA"),
		BrokerURL:       os.Getenv("PACT_BROKER_BASE_URL"),
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

	l.Fatal(http.ListenAndServe("127.0.0.1:8111", mux))
}

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=billy"`
	LastName string `json:"lastName" pact:"example=Sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
}
