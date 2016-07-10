// Package main contains a runnable Consumer Pact test example.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pact-foundation/pact-go/dsl"
)

// Example Pact: How to run me!
// 1. Start the daemon with `./pact-go daemon`
// 2. cd <pact-go>/examples
// 3. go run consumer.go
func main() {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Port:     6666, // Ensure this port matches the daemon port!
		Consumer: "MyConsumer",
		Provider: "MyProvider",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port))
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}
		_, err = http.Get(fmt.Sprintf("http://localhost:%d/bazbat", pact.Server.Port))
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}

		return err
	}

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   "/foobar",
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})
	pact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   "/bazbat",
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
		})

	// Verify
	err := pact.Verify(test)
	if err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

	fmt.Println("Test Passed!")
}
