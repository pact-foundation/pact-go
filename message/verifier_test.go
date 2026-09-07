package message

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// erroringBody is an io.ReadCloser that fails on Read, used to
// characterise CreateMessageHandler's handling of a body it cannot read.
type erroringBody struct{}

func (erroringBody) Read(_ []byte) (int, error) { return 0, errors.New("read boom") }
func (erroringBody) Close() error               { return nil }

// TestCreateMessageHandler_PassThrough characterises the case where the
// request does not target the message-verification path: the middleware
// must not touch the response and must delegate straight to next.
func TestCreateMessageHandler_PassThrough(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	mw := CreateMessageHandler(Handlers{})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/not-the-messages-path", strings.NewReader(""))
	rr := httptest.NewRecorder()

	mw(next).ServeHTTP(rr, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusTeapot, rr.Code)
}

// TestCreateMessageHandler_UnreadableBody characterises the response when
// the request body cannot be read at all.
func TestCreateMessageHandler_UnreadableBody(t *testing.T) {
	mw := CreateMessageHandler(Handlers{})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", erroringBody{})
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateMessageHandler_InvalidJSON characterises the response to a body
// that is readable but not valid JSON.
func TestCreateMessageHandler_InvalidJSON(t *testing.T) {
	mw := CreateMessageHandler(Handlers{})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader("not-json"))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// TestCreateMessageHandler_MessageNotFound characterises the response when
// no handler is registered for the requested description.
func TestCreateMessageHandler_MessageNotFound(t *testing.T) {
	mw := CreateMessageHandler(Handlers{})
	body := `{"description":"an unregistered message"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// TestCreateMessageHandler_HandlerErrors characterises the response when
// the registered handler returns an error.
func TestCreateMessageHandler_HandlerErrors(t *testing.T) {
	handlers := Handlers{
		"a user event": func(_ []models.ProviderState) (Body, Metadata, error) {
			return nil, nil, errors.New("handler boom")
		},
	}
	mw := CreateMessageHandler(handlers)
	body := `{"description":"a user event"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

// TestCreateMessageHandler_ByteBody characterises the success path where
// the handler returns a []byte body: it is written through unmodified,
// without being re-encoded as JSON.
func TestCreateMessageHandler_ByteBody(t *testing.T) {
	var gotStates []models.ProviderState
	handlers := Handlers{
		"a user event": func(states []models.ProviderState) (Body, Metadata, error) {
			gotStates = states
			// Deliberately not valid JSON: proves the handler writes these
			// bytes through unmodified rather than re-encoding them.
			return []byte("raw-bytes-payload"), nil, nil
		},
	}
	mw := CreateMessageHandler(handlers)
	body := `{"description":"a user event","providerStates":[{"name":"User with id 127 exists"}]}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "raw-bytes-payload", rr.Body.String())
	require.Len(t, gotStates, 1)
	assert.Equal(t, "User with id 127 exists", gotStates[0].Name)
}

// TestCreateMessageHandler_NonByteBody characterises the success path where
// the handler returns a non-[]byte body: it is marshalled as JSON before
// being written.
func TestCreateMessageHandler_NonByteBody(t *testing.T) {
	handlers := Handlers{
		"a user event": func(_ []models.ProviderState) (Body, Metadata, error) {
			return map[string]any{"id": float64(127)}, nil, nil
		},
	}
	mw := CreateMessageHandler(handlers)
	body := `{"description":"a user event"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var got map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
	assert.InDelta(t, 127, got["id"], 0)
}

// TestCreateMessageHandler_MetadataHeaders characterises the metadata
// propagation: non-empty metadata is base64-encoded onto both metadata
// headers, and a "contentType" entry overrides the response Content-Type.
func TestCreateMessageHandler_MetadataHeaders(t *testing.T) {
	handlers := Handlers{
		"a user event": func(_ []models.ProviderState) (Body, Metadata, error) {
			return []byte("payload"), Metadata{"contentType": "text/plain"}, nil
		},
	}
	mw := CreateMessageHandler(handlers)
	body := `{"description":"a user event"}`
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/__messages", strings.NewReader(body))
	rr := httptest.NewRecorder()

	mw(nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
	assert.NotEmpty(t, rr.Header().Get(PACT_MESSAGE_METADATA_HEADER))
	assert.NotEmpty(t, rr.Header().Get(PACT_MESSAGE_METADATA_HEADER2))
}

var _ io.ReadCloser = erroringBody{}
