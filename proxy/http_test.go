package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func dummyHandler(header string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header, "true")
	})
}

func DummyMiddleware(header string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header, "true")
			next.ServeHTTP(w, r)
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	req, err := http.NewRequest("GET", "/x", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	loggingMiddleware(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Errorf("expected handler to set the header 'X-Dummy-Handler: true' but got '%v'",
			h)
	}
}

func TestChainHandlers(t *testing.T) {
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	headers := []string{
		"1",
		"2",
		"3",
		"X-Dummy-Handler",
	}
	mw := []Middleware{
		DummyMiddleware("1"),
		DummyMiddleware("2"),
		DummyMiddleware("3"),
		DummyMiddleware("X-Dummy-Handler"),
	}

	middlewareChain := chainHandlers(mw...)
	middlewareChain(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	for _, h := range headers {
		if v := rr.HeaderMap.Get(h); v != "true" {
			t.Errorf("expected handler to set the header '%v: true' but got '%v'",
				h, v)
		}
	}
}

func TestHTTPReverseProxy(t *testing.T) {

	// Setup target to proxy
	port, err := HTTPReverseProxy(Options{
		Middleware: []Middleware{
			DummyMiddleware("1"),
		},
		TargetScheme:  "http",
		TargetAddress: fmt.Sprintf("127.0.0.1:1234"),
	})

	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if port == 0 {
		t.Errorf("want non-zero port, got %v", port)
	}
}
