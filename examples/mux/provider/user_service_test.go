package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	examples "github.com/pact-foundation/pact-go/examples/types"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// The actual Provider test itself
func TestPact_Provider(t *testing.T) {
	go startInstrumentedProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", port),
		PactURLs:               []string{fmt.Sprintf("%s/billy-bobby.json", pactDir)},
		ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", port),
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/login", UserLogin)
	mux.HandleFunc("/setup", providerStateSetupFunc)

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
	if state.State == "User billy exists" {
		userRepository = billyExists
	} else if state.State == "User billy is unauthorized" {
		userRepository = billyUnauthorized
	} else {
		userRepository = billyDoesNotExist
	}
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Provider States data sets
var billyExists = &examples.UserRepository{
	Users: map[string]*examples.User{
		"billy": &examples.User{
			Name:     "billy",
			Username: "billy",
			Password: "issilly",
		},
	},
}

var billyDoesNotExist = &examples.UserRepository{}

var billyUnauthorized = &examples.UserRepository{
	Users: map[string]*examples.User{
		"billy": &examples.User{
			Name:     "billy",
			Username: "billy",
			Password: "issilly1",
		},
	},
}

// Setup the Pact client.
func createPact() dsl.Pact {
	// Create Pact connecting to local Daemon
	return dsl.Pact{
		Port:     6666,
		Consumer: "billy",
		Provider: "bobby",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
}
