package steps

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestSendRequest(t *testing.T) {
	var got *http.Request
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(r.Context())
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("X-Reply", "yes")
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	}))
	t.Cleanup(server.Close)

	//nolint:bodyclose // sendRequest reads and closes the body before returning
	res, body, err := sendRequest(server.URL, model.Request{
		Method:  http.MethodPut,
		Path:    "/basic",
		Query:   url.Values{"b": {"2"}, "a": {"1"}},
		Headers: http.Header{"X-Test": {"Compatibility"}},
		Body:    &model.Body{ContentType: "application/json", Contents: []byte(`{"one":"a"}`)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusTeapot || res.Header.Get("X-Reply") != "yes" || string(body) != "short and stout" {
		t.Fatalf("response: %d %v %q", res.StatusCode, res.Header, body)
	}
	if got.Method != http.MethodPut || got.URL.Path != "/basic" || got.URL.RawQuery != "a=1&b=2" ||
		got.Header.Get("X-Test") != "Compatibility" || got.Header.Get("Content-Type") != "application/json" ||
		string(gotBody) != `{"one":"a"}` {
		t.Fatalf("request: %s %s %v %q", got.Method, got.URL, got.Header, gotBody)
	}
}

func TestJSONEqual(t *testing.T) {
	err := jsonEqual([]byte(`{"one":"a","two":"b"}`), []byte("{\n  \"two\": \"b\",\n  \"one\": \"a\"\n}"))
	if err != nil {
		t.Fatal(err)
	}
	err = jsonEqual([]byte(`{"one":"a"}`), []byte(`{"one":"b"}`))
	if err == nil {
		t.Fatal("want an error for different documents")
	}
	err = jsonEqual([]byte(`not json`), []byte(`{}`))
	if err == nil {
		t.Fatal("want an error for a body that is not JSON")
	}
}
