// Package fakes provides in-process stand-ins for the services a provider
// verification talks to.
package fakes

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sync"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// RecordedRequest is a request a fake received.
type RecordedRequest struct {
	Method string
	Path   string
	Header http.Header
}

// Provider is an HTTP provider that answers each request with the response
// of the first interaction whose method, path and query match it, and with a
// 500 otherwise, as a Pact mock server would. It records every request.
type Provider struct {
	server       *httptest.Server
	interactions []model.Interaction
	mu           sync.Mutex
	requests     []RecordedRequest
}

// NewProvider starts a provider serving interactions.
func NewProvider(interactions []model.Interaction) *Provider {
	p := &Provider{interactions: interactions}
	p.server = httptest.NewServer(http.HandlerFunc(p.serve))

	return p
}

// URL is the provider's base URL.
func (p *Provider) URL() string { return p.server.URL }

// Close stops the provider.
func (p *Provider) Close() { p.server.Close() }

// Requests returns the requests received so far.
func (p *Provider) Requests() []RecordedRequest {
	p.mu.Lock()
	defer p.mu.Unlock()

	return slices.Clone(p.requests)
}

func (p *Provider) serve(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	p.requests = append(p.requests, RecordedRequest{Method: r.Method, Path: r.URL.Path, Header: r.Header.Clone()})
	p.mu.Unlock()

	for _, in := range p.interactions {
		if matches(in.Request, r) {
			writeResponse(w, in.Response)

			return
		}
	}
	http.Error(w, "no interaction matches "+r.Method+" "+r.URL.RequestURI(), http.StatusInternalServerError)
}

func matches(want model.Request, r *http.Request) bool {
	if want.Method != r.Method || want.Path != r.URL.Path {
		return false
	}
	got := r.URL.Query()
	if len(want.Query) == 0 {
		return len(got) == 0
	}

	return reflect.DeepEqual(map[string][]string(want.Query), map[string][]string(got))
}

func writeResponse(w http.ResponseWriter, res model.Response) {
	for name, values := range res.Headers {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}
	if res.Body != nil && res.Body.ContentType != "" && w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", res.Body.ContentType)
	}
	w.WriteHeader(res.Status)
	if res.Body != nil {
		_, _ = w.Write(res.Body.Contents)
	}
}
