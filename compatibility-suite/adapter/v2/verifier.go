package v2

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/proxy"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
)

type verifierAdapter struct{}

// Verify runs pact-go's provider verification, and reads the verifier's own
// result through internal/native because pact-go returns only a sentinel
// error.
func (verifierAdapter) Verify(req adapter.VerifyRequest) (adapter.VerifyResult, error) {
	request := provider.VerifyRequest{
		ProviderBaseURL:            req.ProviderURL,
		PactFiles:                  req.PactFiles,
		ProviderVersion:            req.ProviderVersion,
		PublishVerificationResults: req.Publish,
	}
	if req.Broker != nil {
		request.BrokerURL = req.Broker.URL
		request.Provider = req.Broker.Provider
	}
	if req.StateCallback != nil {
		request.StateHandlers = stateHandlers(req.States, req.StateCallback)
	}
	if len(req.Headers) > 0 {
		request.RequestFilter = addHeaders(req.Headers)
	}

	var resultJSON string
	native.ResultObserver = func(result string) { resultJSON = result }
	defer func() { native.ResultObserver = nil }()

	var verifyErr error
	sandbox("provider", func(t *testing.T) {
		t.Helper()
		verifyErr = provider.NewVerifier().VerifyProvider(t, request)
	})

	if resultJSON == "" {
		return adapter.VerifyResult{}, fmt.Errorf("the verifier reported no result (pact-go returned: %w)", verifyErr)
	}

	errs, err := parseVerifierResult([]byte(resultJSON))
	if err != nil {
		return adapter.VerifyResult{}, err
	}

	return adapter.VerifyResult{Err: verifyErr, Errors: errs}, nil
}

// stateHandlers registers callback for each state. pact-go looks handlers up
// by exact state name, so an interaction without a state needs "" here.
func stateHandlers(states []string, callback func(setup bool, state string) error) models.StateHandlers {
	handlers := make(models.StateHandlers, len(states))
	for _, state := range states {
		handlers[state] = func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return nil, callback(setup, s.Name)
		}
	}

	return handlers
}

// addHeaders is a pact-go request filter adding headers to every request.
func addHeaders(headers http.Header) proxy.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for name, values := range headers {
				for _, v := range values {
					r.Header.Add(name, v)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
