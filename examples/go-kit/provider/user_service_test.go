package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/go-kit/kit/log"

	"golang.org/x/net/context"

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
		"billy": &User{
			Name:     "billy",
			username: "billy",
			password: "issilly",
		},
	},
}

var billyDoesNotExist = &UserRepository{}

var billyUnauthorized = &UserRepository{
	users: map[string]*User{
		"billy": &User{
			Name:     "billy",
			username: "billy",
			password: "issilly1",
		},
	},
}

// The actual Provider test itself
func TestPact_Provider(t *testing.T) {
	go startInstrumentedProvider()

	pact := createPact()

	// Verify the Provider with local Pact Files
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", port),
		PactURLs:               []string{fmt.Sprintf("%s/billy-bobby.json", pactDir)},
		ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", port),
		ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", port),
	})

	if err != nil {
		t.Fatal("Error:", err)
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
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
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

	// This path returns all states available for
	router.Methods("GET").Path("/states").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Log("[DEBUG] returning available provider states")
		w.Header().Add("Content-Type", "application/json")
		var states types.ProviderStates
		states = map[string][]string{
			"billy": []string{
				"User billy exists",
				"User billy does not exist",
				"User billy is unauthorized"},
		}
		data, _ := json.Marshal(states)
		fmt.Fprintf(w, string(data))
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
