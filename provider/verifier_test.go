package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStateHandlerMiddleware_PassThrough characterises the case where the
// request does not target the state-setup path: the middleware must not
// touch the response and must delegate straight to next.
func TestStateHandlerMiddleware_PassThrough(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	mw := stateHandlerMiddleware(models.StateHandlers{}, nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/not-the-setup-path", strings.NewReader(""))
	rr := httptest.NewRecorder()

	mw(next).ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}

// TestStateHandlerMiddleware_InvalidPayload characterises the response to a
// request body that cannot be decoded as a stateHandlerAction.
func TestStateHandlerMiddleware_InvalidPayload(t *testing.T) {
	mw := stateHandlerMiddleware(models.StateHandlers{}, nil)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader("not-json"))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestStateHandlerMiddleware_StateNotFound characterises the "no handler
// registered for this state" path: it returns 200 without invoking any
// handler, sharing the trailing w.WriteHeader(http.StatusOK) with the
// success path inside the "found" branch.
func TestStateHandlerMiddleware_StateNotFound(t *testing.T) {
	handlerCalled := false
	handlers := models.StateHandlers{
		"a different state": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			handlerCalled = true
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"setup","state":"unregistered state"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

// TestStateHandlerMiddleware_SetupNoResponse characterises the setup path
// for a found state handler that returns no provider state values: it
// falls through to the same trailing 200 as the "not found" case.
func TestStateHandlerMiddleware_SetupNoResponse(t *testing.T) {
	var gotSetup bool
	var gotState models.ProviderState
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			gotSetup = setup
			gotState = s
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"setup","state":"User foo exists","id":"foo"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.True(t, gotSetup)
	assert.Equal(t, "User foo exists", gotState.Name)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

// TestStateHandlerMiddleware_TeardownNoAfterEach characterises the teardown
// path when no AfterEach hook is configured: the handler runs with
// setup=false and afterEach is simply skipped.
func TestStateHandlerMiddleware_TeardownNoAfterEach(t *testing.T) {
	var gotSetup bool
	handlerCalled := false
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			handlerCalled = true
			gotSetup = setup
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"teardown","state":"User foo exists"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.True(t, handlerCalled)
	assert.False(t, gotSetup)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestStateHandlerMiddleware_TeardownAfterEachSucceeds characterises the
// teardown path when an AfterEach hook is configured and succeeds.
func TestStateHandlerMiddleware_TeardownAfterEachSucceeds(t *testing.T) {
	afterEachCalled := false
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, func() error {
		afterEachCalled = true
		return nil
	})
	body := `{"action":"teardown","state":"User foo exists"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.True(t, afterEachCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestStateHandlerMiddleware_TeardownAfterEachErrors characterises the
// teardown path when the AfterEach hook itself fails: the middleware must
// respond 500 and must not fall through to the trailing 200.
func TestStateHandlerMiddleware_TeardownAfterEachErrors(t *testing.T) {
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, func() error {
		return errors.New("after each boom")
	})
	body := `{"action":"teardown","state":"User foo exists"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestStateHandlerMiddleware_HandlerErrors characterises the case where the
// state handler itself returns an error: the middleware must respond 500.
func TestStateHandlerMiddleware_HandlerErrors(t *testing.T) {
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return nil, errors.New("state handler boom")
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"setup","state":"User foo exists"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestStateHandlerMiddleware_ReturnsProviderStateValues characterises the
// success path where the handler returns provider state values: the
// middleware must marshal them as JSON, set the content-type header, and
// respond 200 with the body — a distinct exit from the shared trailing
// w.WriteHeader(http.StatusOK) used by the other success paths.
func TestStateHandlerMiddleware_ReturnsProviderStateValues(t *testing.T) {
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			return models.ProviderStateResponse{"uuid": "1234"}, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"setup","state":"User foo exists"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("content-type"))

	var got models.ProviderStateResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	assert.Equal(t, "1234", got["uuid"])
}

// TestStateHandlerMiddleware_ParamsExtracted characterises the extraction
// of extra top-level payload fields into ProviderState.Parameters, with
// "action" and "state" excluded.
func TestStateHandlerMiddleware_ParamsExtracted(t *testing.T) {
	var gotState models.ProviderState
	handlers := models.StateHandlers{
		"User foo exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
			gotState = s
			return nil, nil
		},
	}

	mw := stateHandlerMiddleware(handlers, nil)
	body := `{"action":"setup","state":"User foo exists","id":"foo"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, providerStatesSetupPath, strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, "foo", gotState.Parameters["id"])
	_, hasAction := gotState.Parameters["action"]
	_, hasState := gotState.Parameters["state"]
	assert.False(t, hasAction)
	assert.False(t, hasState)
}
