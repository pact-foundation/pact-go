package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
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
	router := gin.Default()
	router.POST("/users/login", UserLogin)
	router.POST("/setup", providerStateSetup)

	router.Run(fmt.Sprintf(":%d", port))
}

// Set current provider state route.
func providerStateSetup(c *gin.Context) {
	var state types.ProviderState
	if c.BindJSON(&state) == nil {
		// Setup database for different states
		if state.State == "User billy exists" {
			userRepository = billyExists
		} else if state.State == "User billy is unauthorized" {
			userRepository = billyUnauthorized
		} else {
			userRepository = billyDoesNotExist
		}
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
