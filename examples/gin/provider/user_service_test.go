package provider

import (
	"fmt"
	"net/http"
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
	go startProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: fmt.Sprintf("http://localhost:%d", port),
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/jmarie-loginprovider.json", pactDir))},
		StateHandlers:   stateHandlers,
		RequestFilter:   fixBearerToken,
		BeforeEach: func() error {
			fmt.Println("before hook")
			return nil
		},
		AfterEach: func() error {
			fmt.Println("after hook")
			return nil
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	// Pull from pact broker, used in e2e/integrated tests for pact-go release
	if os.Getenv("PACT_INTEGRATED_TESTS") != "" {
		var brokerHost = os.Getenv("PACT_BROKER_HOST")

		// Verify the Provider - Latest Published Pacts for any known consumers
		_, err := pact.VerifyProvider(t, types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
			BrokerURL:                  brokerHost,
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
			BrokerURL:                  brokerHost,
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

	} else {
		t.Log("Skipping pulling from broker as PACT_INTEGRATED_TESTS is not set")
	}
}

var token = "" // token will be dynamic based on state etc.

// Provider state handlers
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
		fmt.Println("state handler")
		userRepository = jmarieDoesNotExist
		return nil
	},
}

// Simulates the neeed to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("request filter - before")
		r.Header.Set("Authorization", token)
		next.ServeHTTP(w, r)
		fmt.Println("request filter - after")
	})
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startProvider() {
	router := gin.Default()
	router.POST("/login/:id", UserLogin)
	router.GET("/users/:id", IsAuthenticated(), GetUser)

	router.Run(fmt.Sprintf(":%d", port))
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
		Consumer:                 "jmarie",
		Provider:                 "loginprovider",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
		LogLevel:                 "DEBUG",
	}
}
