// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	v3 "github.com/pact-foundation/pact-go/v3"
)

// Example Pact: How to run me!
// 1. cd <pact-go>/examples
// 2. go test -v -run TestConsumer
func TestConsumer(t *testing.T) {
	type User struct {
		Name     string `json:"name" pact:"example=billy"`
		LastName string `json:"lastName" pact:"example=sampson"`
	}

	// Create Pact connecting to local Daemon
	pact := &v3.MockProvider{
		Consumer: "MyConsumer",
		Provider: "MyProvider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func(config v3.MockServerConfig) error {
		u := fmt.Sprintf("http://localhost:%d/foobar", pact.ServerPort)
		req, err := http.NewRequest("GET", u, strings.NewReader(`{"name":"billy"}`))

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 1234")

		if err != nil {
			return err
		}
		if _, err = http.DefaultClient.Do(req); err != nil {
			return err
		}

		return err
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to get foo").
		WithRequest(v3.Request{
			Method:  "GET",
			Path:    v3.String("/foobar"),
			Headers: v3.MapMatcher{"Content-Type": v3.String("application/json"), "Authorization": v3.String("Bearer 1234")},
			Body: map[string]string{
				"name": "billy",
			},
		}).
		WillRespondWith(v3.Response{
			Status:  200,
			Headers: v3.MapMatcher{"Content-Type": v3.String("application/json")},
			Body:    v3.Match(&User{}),
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}
