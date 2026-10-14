package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
)

// writePact writes a V2 pact expecting GET /basic to return 200 and the basic
// JSON document, optionally in a provider state.
func writePact(t *testing.T, state string) string {
	t.Helper()
	interaction := map[string]any{
		"description": "get basic",
		"request":     map[string]any{"method": "GET", "path": "/basic"},
		"response": map[string]any{
			"status":  200,
			"headers": map[string]string{"Content-Type": "application/json"},
			"body":    map[string]string{"one": "a", "two": "b"},
		},
	}
	if state != "" {
		interaction["providerState"] = state
	}
	pact := map[string]any{
		"consumer":     map[string]string{"name": "c"},
		"provider":     map[string]string{"name": "p"},
		"interactions": []any{interaction},
		"metadata":     map[string]any{"pactSpecification": map[string]string{"version": "2.0.0"}},
	}

	data, err := json.Marshal(pact)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "c-p.json")
	err = os.WriteFile(path, data, 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

// headerLog records the headers of each request a stub receives.
type headerLog struct {
	mu  sync.Mutex
	all []http.Header
}

func (l *headerLog) add(h http.Header) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.all = append(l.all, h.Clone())
}

func (l *headerLog) list() []http.Header {
	l.mu.Lock()
	defer l.mu.Unlock()

	return slices.Clone(l.all)
}

// stubProvider answers every request with status and the basic JSON document.
func stubProvider(t *testing.T, status int) (*httptest.Server, *headerLog) {
	t.Helper()
	log := &headerLog{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.add(r.Header)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(basicJSON)
	}))
	t.Cleanup(server.Close)

	return server, log
}

func TestVerifierReportsSuccess(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err != nil || len(result.Errors) != 0 {
		t.Fatalf("want a pass, got %v %+v", result.Err, result.Errors)
	}
}

func TestVerifierReportsMismatchTypes(t *testing.T) {
	server, _ := stubProvider(t, http.StatusBadRequest)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err == nil {
		t.Fatal("want the verification to fail")
	}
	if len(result.Errors) == 0 || !slices.Contains(result.Errors[0].MismatchTypes, "StatusMismatch") {
		t.Fatalf("want a StatusMismatch, got %+v", result.Errors)
	}
}

func TestVerifierCallsStateCallback(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	var (
		mu    sync.Mutex
		calls []string
	)
	result, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: server.URL,
		PactFiles:   []string{writePact(t, "state one")},
		States:      []string{"", "state one"},
		StateCallback: func(setup bool, state string) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, fmt.Sprintf("%t:%s", setup, state))

			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Err != nil {
		t.Fatalf("want a pass, got %v %+v", result.Err, result.Errors)
	}
	if !slices.Contains(calls, "true:state one") || !slices.Contains(calls, "false:state one") {
		t.Fatalf("want setup and teardown of 'state one', got %v", calls)
	}
}

func TestVerifierAddsHeaders(t *testing.T) {
	server, log := stubProvider(t, http.StatusOK)
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: server.URL,
		PactFiles:   []string{writePact(t, "")},
		Headers:     http.Header{"A": {"1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	received := log.list()
	if len(received) == 0 {
		t.Fatal("the provider received no requests")
	}
	for _, h := range received {
		if h.Get("A") != "1" {
			t.Fatalf("want header A: 1 on every request, got %v", h)
		}
	}
}

func TestVerifierReportsWhenNothingRan(t *testing.T) {
	// pact-go rejects a broker source without a provider name before it calls
	// the verifier, so no result is ever reported.
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{
		ProviderURL: "http://127.0.0.1:1",
		Broker:      &adapter.BrokerSource{URL: "http://127.0.0.1:1"},
	})
	if err == nil {
		t.Fatal("want an error when the verifier never ran")
	}
}

func TestVerifierClearsObserver(t *testing.T) {
	server, _ := stubProvider(t, http.StatusOK)
	_, err := verifierAdapter{}.Verify(adapter.VerifyRequest{ProviderURL: server.URL, PactFiles: []string{writePact(t, "")}})
	if err != nil {
		t.Fatal(err)
	}
	if native.ResultObserver != nil {
		t.Fatal("want ResultObserver cleared after a verification")
	}
}
