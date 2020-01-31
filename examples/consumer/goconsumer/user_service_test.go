// +build consumer

package goconsumer

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/dsl"
	ex "github.com/pact-foundation/pact-go/examples/types"
	"github.com/pact-foundation/pact-go/types"
)

// Common test data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var pact dsl.Pact
var form url.Values
var rr http.ResponseWriter
var req *http.Request
var name = "jmarie"
var password = "issilly"
var like = dsl.Like
var eachLike = dsl.EachLike
var term = dsl.Term

type s = dsl.String
type request = dsl.Request

var loginRequest = ex.LoginRequest{
	Username: name,
	Password: password,
}
var commonHeaders = dsl.MapMatcher{
	"Content-Type": term("application/json; charset=utf-8", `application\/json`),
}

// Use this to control the setup and teardown of Pact
func TestMain(m *testing.M) {
	// Setup Pact and related test stuff
	setup()

	// Run all the tests
	code := m.Run()

	// Shutdown the Mock Service and Write pact files to disk
	pact.WritePact()
	pact.Teardown()

	// Enable when running E2E/integration tests before a release
	version := "1.0.0"
	if os.Getenv("TRAVIS_BUILD_NUMBER") != "" {
		version = fmt.Sprintf("1.0.%s-%d", os.Getenv("TRAVIS_BUILD_NUMBER"), time.Now().Unix())
	}

	// Publish the Pacts...
	p := dsl.Publisher{}

	err := p.Publish(types.PublishRequest{
		PactURLs:        []string{filepath.FromSlash(fmt.Sprintf("%s/jmarie-loginprovider.json", pactDir))},
		PactBroker:      fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		ConsumerVersion: version,
		Tags:            []string{"dev", "prod"},
		BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
	})

	if err != nil {
		log.Println("ERROR: ", err)
		os.Exit(1)
	}

	os.Exit(code)
}

// Setup common test data
func setup() {
	pact = createPact()

	// Proactively start service to get access to the port
	pact.Setup(true)

	// Login form values
	form = url.Values{}
	form.Add("username", name)
	form.Add("password", password)

	// Create a request to pass to our handler.
	req, _ = http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.PostForm = form

	rr = httptest.NewRecorder()
}

// Create Pact connecting to local Daemon
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer:                 "jmarie",
		Provider:                 "loginprovider",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
	}
}

func TestExampleConsumerLoginHandler_UserExists(t *testing.T) {
	var testJmarieExists = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
			// This header will be dynamically replaced at verification time
			token: "Bearer 1234",
		}
		client.loginHandler(rr, req)

		// Expect User to be set on the Client
		if client.user == nil {
			return errors.New("Expected user not to be nil")
		}

		return nil
	}

	// Pull from pact broker, used in e2e/integrated tests for pact-go release
	// Setup interactions on the Mock Service. Note that you can have multiple
	// interactions
	pact.
		AddInteraction().
		Given("User jmarie exists").
		UponReceiving("A request to login with user 'jmarie'").
		WithRequest(request{
			Method: "POST",
			Path:   term("/login/10", "/login/[0-9]+"),
			Query: dsl.MapMatcher{
				"foo": term("bar", "[a-zA-Z]+"),
			},
			Body:    loginRequest,
			Headers: commonHeaders,
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: dsl.Match(ex.LoginResponse{
				User: &ex.User{},
			}),
			Headers: dsl.MapMatcher{
				"X-Api-Correlation-Id": dsl.Like("100"),
				"Content-Type":         term("application/json; charset=utf-8", `application\/json`),
				"X-Auth-Token":         dsl.Like("1234"),
			},
		})

	err := pact.Verify(testJmarieExists)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func TestExampleConsumerLoginHandler_UserDoesNotExist(t *testing.T) {
	var testJmarieDoesNotExists = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
			// This header will be dynamically replaced at verification time
			token: "Bearer 1234",
		}
		client.loginHandler(rr, req)

		if client.user != nil {
			return fmt.Errorf("Expected user to be nil but in stead got: %v", client.user)
		}

		return nil
	}

	pact.
		AddInteraction().
		Given("User jmarie does not exist").
		UponReceiving("A request to login with user 'jmarie'").
		WithRequest(request{
			Method:  "POST",
			Path:    s("/login/10"),
			Body:    loginRequest,
			Headers: commonHeaders,
			Query: dsl.MapMatcher{
				"foo": s("anything"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  404,
			Headers: commonHeaders,
		})

	err := pact.Verify(testJmarieDoesNotExists)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func TestExampleConsumerLoginHandler_UserUnauthorised(t *testing.T) {
	var testJmarieUnauthorized = func() error {
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
		Given("User jmarie is unauthorized").
		UponReceiving("A request to login with user 'jmarie'").
		WithRequest(request{
			Method:  "POST",
			Path:    s("/login/10"),
			Body:    loginRequest,
			Headers: commonHeaders,
		}).
		WillRespondWith(dsl.Response{
			Status:  401,
			Headers: commonHeaders,
		})

	err := pact.Verify(testJmarieUnauthorized)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}

func TestExampleConsumerGetUser_Authenticated(t *testing.T) {
	var testJmarieUnauthenticated = func() error {
		client := Client{
			Host:  fmt.Sprintf("http://localhost:%d", pact.Server.Port),
			token: "Bearer 1234",
		}
		client.getUser("10")

		if client.user != nil {
			return fmt.Errorf("Expected user to be nil but got: %v", client.user)
		}

		return nil
	}

	pact.
		AddInteraction().
		Given("User jmarie is authenticated").
		UponReceiving("A request to get user 'jmarie'").
		WithRequest(request{
			Method: "GET",
			Path:   s("/users/10"),
			Headers: dsl.MapMatcher{
				"Authorization": s("Bearer 1234"),
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: commonHeaders,
			Body:    dsl.Match(ex.User{}),
		})

	err := pact.Verify(testJmarieUnauthenticated)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

}
func TestExampleConsumerGetUser_Unauthenticated(t *testing.T) {
	var testJmarieUnauthenticated = func() error {
		client := Client{
			Host: fmt.Sprintf("http://localhost:%d", pact.Server.Port),
		}
		client.getUser("10")

		if client.user != nil {
			return fmt.Errorf("Expected user to be nil but got: %v", client.user)
		}

		return nil
	}

	pact.
		AddInteraction().
		Given("User jmarie is unauthenticated").
		UponReceiving("A request to get with user 'jmarie'").
		WithRequest(request{
			Method:  "GET",
			Path:    s("/users/10"),
			Headers: commonHeaders,
		}).
		WillRespondWith(dsl.Response{
			Status:  401,
			Headers: commonHeaders,
		})

	err := pact.Verify(testJmarieUnauthenticated)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

}
