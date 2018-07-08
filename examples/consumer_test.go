// Package main contains a runnable Consumer Pact test example.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
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
	pact := &dsl.Pact{
		Consumer: "MyConsumer",
		Provider: "MyProvider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port)
		req, err := http.NewRequest("GET", u, strings.NewReader(`{"name":"billy"}`))

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
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
		WithRequest(dsl.Request{
			Method:  "GET",
			Path:    dsl.String("/foobar"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body: map[string]string{
				"name": "billy",
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    dsl.Match(&User{LastName: "sampson"}),
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

	fmt.Println("Test Passed!")
}
