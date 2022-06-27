//go:build provider
// +build provider

package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	examples "github.com/pact-foundation/pact-go/examples/types"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// The Provider verification
func TestExample_MuxProvider(t *testing.T) {
	go startProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("http://localhost:%d", port),
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/jmarie-loginprovider.json", pactDir))},
		StateHandlers:   stateHandlers,
		RequestFilter:   fixBearerToken,
		BeforeEach: func() error {
			return nil
		},
		AfterEach: func() error {
			return nil
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	// Pull from pact broker, used in e2e/integrated tests for pact-go release

	// Verify the Provider - Latest Published Pacts for any known consumers
	_, err = pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
		BrokerURL:                  os.Getenv("PACT_BROKER_BASE_URL"),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		StateHandlers:              stateHandlers,
		RequestFilter:              fixBearerToken,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Verify the Provider - Tag-based Published Pacts for any known consumers
	_, err = pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
		BrokerURL:                  os.Getenv("PACT_BROKER_BASE_URL"),
		Tags:                       []string{"prod"},
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		StateHandlers:              stateHandlers,
		RequestFilter:              fixBearerToken,
	})

	if err != nil {
		t.Fatal(err)
	}

}

var token = "" // token will be dynamic based on state etc.

// Simulates the neeed to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", token)
		next.ServeHTTP(w, r)
	})
}

var stateHandlers = types.StateHandlers{
	"User jmarie exists": func() error {
		userRepository = jmarieExists
		return nil
	},
	"User jmarie is authenticated": func() error {
		userRepository = jmarieExists
		token = fmt.Sprintf("Bearer %s", getAuthToken())
		return nil
	},
	"User jmarie is unauthorized": func() error {
		userRepository = jmarieUnauthorized
		token = "invalid"

		return nil
	},
	"User jmarie is unauthenticated": func() error {
		userRepository = jmarieUnauthorized
		token = "invalid"

		return nil
	},
	"User jmarie does not exist": func() error {
		userRepository = jmarieDoesNotExist
		return nil
	},
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startProvider() {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/", IsAuthenticated(GetUser))
	mux.HandleFunc("/login/", UserLogin)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", port, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))

}

// Set current provider state route.
var providerStateSetupFunc = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var state types.ProviderState

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	err = json.Unmarshal(body, &state)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Setup database for different states
	if state.State == "User jmarie exists" {
		userRepository = jmarieExists
	} else if state.State == "User jmarie is unauthorized" {
		userRepository = jmarieUnauthorized
	} else {
		userRepository = jmarieDoesNotExist
	}
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Provider States data sets
var jmarieExists = &examples.UserRepository{
	Users: map[string]*examples.User{
		"jmarie": &examples.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "jmarie",
			Password: "issilly",
			Type:     "admin",
			ID:       10,
		},
	},
}

var jmarieDoesNotExist = &examples.UserRepository{}

var jmarieUnauthorized = &examples.UserRepository{
	Users: map[string]*examples.User{
		"jmarie": &examples.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "jmarie",
			Password: "issilly1",
			Type:     "blocked",
			ID:       10,
		},
	},
}

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Provider: "loginprovider",
		LogDir:   logDir,
		PactDir:  pactDir,
		LogLevel: "DEBUG",
		// DisableToolValidityCheck: true,
	}
}
