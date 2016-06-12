package consumer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

func createPact() *dsl.Pact {
	// Create Pact connecting to local Daemon
	pactDaemonPort := 6666
	return &dsl.Pact{
		Port:     pactDaemonPort,
		Consumer: "billy",
		Provider: "bobby",
		LogLevel: "DEBUG",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
}

// Format a JSON document to make comparison easier.
func formatJSON(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}

func TestPact_Consumer(t *testing.T) {
	pact := createPact()
	defer pact.Teardown()
	loginRequest := formatJSON(`
    {
      "username":"billy",
      "password": "issilly"
    }`)

	// Pass in test case
	var test = func() error {
		res, err := http.Post(fmt.Sprintf("http://localhost:%d/users/login", pact.Server.Port), "application/json", bytes.NewReader([]byte(loginRequest)))
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}
		fmt.Println("Response: ")
		fmt.Println(res)

		return err
	}

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("User billy exists").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(&dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(&dsl.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

	// Verify Collaboration Test interactions (Consumer side)
	err := pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

	pact.
		AddInteraction().
		Given("User billy does not exist").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(&dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(&dsl.Response{
			Status: 404,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

	// Verify Collaboration Test interactions (Consumer side)
	err = pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

	pact.
		AddInteraction().
		Given("User billy is unauthorized").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(&dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(&dsl.Response{
			Status: 403,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})

	// Verify Collaboration Test interactions (Consumer side)
	err = pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

	// Write the Pact file out to file
	pact.WritePact()
}
