package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/go-kit/kit/log"

	"context"

	"github.com/gorilla/mux"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Provider States data sets
var billyExists = &UserRepository{
	users: map[string]*User{
		"Jean-Marie de La Beaujardière😀😍": &User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			username: "Jean-Marie de La Beaujardière😀😍",
			password: "issilly",
			Type:     "admin",
		},
	},
}

var billyDoesNotExist = &UserRepository{}

var billyUnauthorized = &UserRepository{
	users: map[string]*User{
		"Jean-Marie de La Beaujardière😀😍": &User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			username: "Jean-Marie de La Beaujardière😀😍",
			password: "issilly1",
			Type:     "blocked",
		},
	},
}

// The actual Provider test itself
func TestPact_GoKitProvider(t *testing.T) {
	go startInstrumentedProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", port),
		PactURLs:               []string{filepath.ToSlash(fmt.Sprintf("%s/billy-bobby.json", pactDir))},
		ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", port),
	})

	// Pull from pact broker, used in e2e/integrated tests for pact-go release
	if os.Getenv("PACT_INTEGRATED_TESTS") != "" {
		var brokerHost = os.Getenv("PACT_BROKER_HOST")

		// Verify the Provider - Specific Published Pacts
		pact.VerifyProvider(t, types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
			PactURLs:                   []string{fmt.Sprintf("%s/pacts/provider/bobby/consumer/billy/latest/sit4", brokerHost)},
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

// Setup the Pact client.
func createPact() dsl.Pact {
	// Create Pact connecting to local Daemon
	pactDaemonPort := 6666
	return dsl.Pact{
		Port:     pactDaemonPort,
		Consumer: "billy",
		Provider: "bobby",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	ctx := context.Background()
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}
	var s Service
	{
		s = NewInmemService()
		s = LoggingMiddleware(logger)(s)
	}

	h := MakeHTTPHandler(ctx, s, logger)
	router := h.(*mux.Router)

	// Bolt on 2 extra endpoints to configure Provider states
	// http://docs.pact.io/documentation/provider_states.html
	router.Methods("POST").Path("/setup").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Log("[DEBUG] provider API: states setup")
		w.Header().Add("Content-Type", "application/json")

		var state types.ProviderState

		body, err := ioutil.ReadAll(req.Body)
		logger.Log(string(body))
		req.Body.Close()
		if err != nil {
			return
		}
		json.Unmarshal(body, &state)

		svc := s.(*loggingMiddleware).next.(*userService)

		// Setup database for different states
		if state.State == "User billy exists" {
			svc.userDatabase = billyExists
		} else if state.State == "User billy is unauthorized" {
			svc.userDatabase = billyUnauthorized
		} else {
			svc.userDatabase = billyDoesNotExist
		}

		logger.Log("[DEBUG] configured provider state: ", state.State)
	})

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log(fmt.Sprintf("[DEBUG] starting service on :%d", port))
		errs <- http.ListenAndServe(fmt.Sprintf(":%d", port), h)
	}()

	logger.Log("Exit", <-errs)
}
