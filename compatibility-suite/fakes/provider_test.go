package fakes_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func get(t *testing.T, target string, header http.Header) (int, http.Header, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header = header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, res.Header, string(body)
}

func basicProvider(t *testing.T) *fakes.Provider {
	t.Helper()
	p := fakes.NewProvider([]model.Interaction{
		{
			Request: model.Request{Method: http.MethodGet, Path: "/basic"},
			Response: model.Response{
				Status:  http.StatusOK,
				Headers: http.Header{"X-Test": {"Something"}},
				Body:    &model.Body{ContentType: "application/json", Contents: []byte(`{"one":"a"}`)},
			},
		},
		{
			Request:  model.Request{Method: http.MethodGet, Path: "/with_params", Query: url.Values{"a": {"1"}}},
			Response: model.Response{Status: http.StatusNoContent},
		},
	})
	t.Cleanup(p.Close)

	return p
}

func TestProviderAnswersMatchingRequests(t *testing.T) {
	p := basicProvider(t)
	status, header, body := get(t, p.URL()+"/basic", http.Header{})
	if status != http.StatusOK || header.Get("Content-Type") != "application/json" ||
		header.Get("X-Test") != "Something" || body != `{"one":"a"}` {
		t.Fatalf("got %d %v %q", status, header, body)
	}

	status, _, _ = get(t, p.URL()+"/with_params?a=1", http.Header{})
	if status != http.StatusNoContent {
		t.Fatalf("want 204 for a matching query, got %d", status)
	}
}

func TestProviderRejectsUnmatchedRequests(t *testing.T) {
	p := basicProvider(t)
	for _, target := range []string{"/other", "/with_params?a=2", "/with_params", "/basic?x=1"} {
		status, _, _ := get(t, p.URL()+target, http.Header{})
		if status != http.StatusInternalServerError {
			t.Errorf("%s: want 500, got %d", target, status)
		}
	}
}

func TestProviderRecordsRequests(t *testing.T) {
	p := basicProvider(t)
	get(t, p.URL()+"/basic", http.Header{"A": {"1"}})

	got := p.Requests()
	if len(got) != 1 || got[0].Method != http.MethodGet || got[0].Path != "/basic" || got[0].Header.Get("A") != "1" {
		t.Fatalf("got %+v", got)
	}
}
