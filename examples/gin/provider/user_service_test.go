package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pact-foundation/pact-go/dsl"
	examples "github.com/pact-foundation/pact-go/examples/types"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// The actual Provider test itself
func TestPact_GinProvider(t *testing.T) {
	go startInstrumentedProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", port),
		PactURLs:               []string{filepath.ToSlash(fmt.Sprintf("%s/jmarie-loginprovider.json", pactDir))},
		ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", port),
	})

	if err != nil {
		t.Fatal(err)
	}

	// Pull from pact broker, used in e2e/integrated tests for pact-go release
	if os.Getenv("PACT_INTEGRATED_TESTS") != "" {
		var brokerHost = os.Getenv("PACT_BROKER_HOST")

		// Verify the Provider - Specific Published Pacts
		pact.VerifyProvider(t, types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
			PactURLs:                   []string{fmt.Sprintf("%s/pacts/provider/loginprovider/consumer/jmarie/latest/sit4", brokerHost)},
			ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", port),
			BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
			PublishVerificationResults: true,
			ProviderVersion:            "1.0.0",
		})

		// Verify the Provider - Latest Published Pacts for any known consumers
		pact.VerifyProvider(t, types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
			BrokerURL:                  brokerHost,
			ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", port),
			BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
			PublishVerificationResults: true,
			ProviderVersion:            "1.0.0",
		})

		// Verify the Provider - Tag-based Published Pacts for any known consumers
		pact.VerifyProvider(t, types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
			ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", port),
			BrokerURL:                  brokerHost,
			Tags:                       []string{"latest", "sit4"},
			BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
			PublishVerificationResults: true,
			ProviderVersion:            "1.0.0",
		})

	} else {
		t.Log("Skipping pulling from broker as PACT_INTEGRATED_TESTS is not set")
	}
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	router := gin.Default()
	router.POST("/users/login/:id", UserLogin)
	router.POST("/setup", providerStateSetup)

	router.Run(fmt.Sprintf(":%d", port))
}

// Set current provider state route.
func providerStateSetup(c *gin.Context) {
	var state types.ProviderState
	if c.BindJSON(&state) == nil {
		// Setup database for different states
		if state.State == "User jmarie exists" {
			userRepository = jmarieExists
		} else if state.State == "User jmarie is unauthorized" {
			userRepository = jmarieUnauthorized
		} else {
			userRepository = jmarieDoesNotExist
		}
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
		},
	},
}

// Setup the Pact client.
func createPact() dsl.Pact {
	// Create Pact connecting to local Daemon
	return dsl.Pact{
		Consumer: "jmarie",
		Provider: "loginprovider",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
}
