package goconsumer

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

// Common test data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var pact dsl.Pact
var loginRequest = ` { "username":"billy", "password": "issilly" }`
var form url.Values
var rr http.ResponseWriter
var req *http.Request

// Use this to control the setup and teardown of Pact
func TestMain(m *testing.M) {
	// Setup Pact and related test stuff
	setup()

	// Run all the tests
	code := m.Run()

	// Shutdown the Mock Service and Write pact files to disk
	pact.WritePact()
	pact.Teardown()

	os.Exit(code)
}

// Setup common test data
func setup() {
	pact = createPact()

	// Login form values
	form = url.Values{}
	form.Add("username", "billy")
	form.Add("password", "issilly")

	// Create a request to pass to our handler.
	req, _ = http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = form

	// Record response (satisfies http.ResponseWriter)
	rr = httptest.NewRecorder()
}

// Create Pact connecting to local Daemon
func createPact() dsl.Pact {
	pactDaemonPort := 6666
	return dsl.Pact{
		Port:     pactDaemonPort,
		Consumer: "billy",
		Provider: "bobby",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
}

func TestPactConsumerLoginHandler_UserExists(t *testing.T) {
	var testBillyExists = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
		}
		client.loginHandler(rr, req)

		// Expect User to be set on the Client
		if client.user == nil {
			return errors.New("Expected user not to be nil")
		}

		return nil
	}

	// Setup interactions on the Mock Service. Note that you can have multiple
	// interactions
	pact.
		AddInteraction().
		Given("User billy exists").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
			Body: `
				{
				  "user": {
				    "name": "billy"
				  }
				}
			`,
		})

	err := pact.Verify(testBillyExists)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func TestPactConsumerLoginHandler_UserDoesNotExist(t *testing.T) {
	var testBillyDoesNotExists = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
		}
		client.loginHandler(rr, req)

		if client.user != nil {
			return fmt.Errorf("Expected user to be nil but got: %v", client.user)
		}

		return nil
	}

	pact.
		AddInteraction().
		Given("User billy does not exist").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(dsl.Response{
			Status: 404,
			Headers: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
		})

	err := pact.Verify(testBillyDoesNotExists)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func TestPactConsumerLoginHandler_UserUnauthorised(t *testing.T) {
	var testBillyUnauthorized = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
		}
		client.loginHandler(rr, req)

		if client.user != nil {
			return fmt.Errorf("Expected user to be nil but got: %v", client.user)
		}

		return nil
	}

	pact.
		AddInteraction().
		Given("User billy is unauthorized").
		UponReceiving("A request to login with user 'billy'").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   "/users/login",
			Body:   loginRequest,
		}).
		WillRespondWith(dsl.Response{
			Status: 401,
			Headers: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
		})

	err := pact.Verify(testBillyUnauthorized)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
