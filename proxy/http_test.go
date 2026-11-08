package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/utils"
)

func dummyHandler(header string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	loggingMiddleware(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	if h := rr.Result().Header.Get("X-Dummy-Handler"); h != "true" {
		t.Errorf("expected handler to set the header 'X-Dummy-Handler: true' but got '%v'",
			h)
	}
}

func TestChainHandlers(t *testing.T) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health-check", nil)
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
		if v := rr.Result().Header.Get(h); v != "true" {
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
		TargetAddress: "127.0.0.1:1234",
	})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if port == 0 {
		t.Errorf("want non-zero port, got %v", port)
	}
}

func TestHTTPReverseProxy_PortInUseReturnsError(t *testing.T) {
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = l.Close()
	}()

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("expected *net.TCPAddr, got %T", l.Addr())
	}

	proxyPort, proxyErr := HTTPReverseProxy(Options{
		ProxyPort:     tcpAddr.Port,
		TargetScheme:  "http",
		TargetAddress: "127.0.0.1:1234",
	})
	if proxyErr == nil {
		t.Errorf("expected error when port %d is already in use, got nil", tcpAddr.Port)
	}
	if proxyPort != 0 {
		t.Errorf("expected port 0 on failure, got %d", proxyPort)
	}
}

func TestHTTPReverseProxy_ListeningWhenReturned(t *testing.T) {
	var dialer net.Dialer
	for i := range 200 {
		port, proxyErr := HTTPReverseProxy(Options{
			ProxyPort:     0,
			TargetScheme:  "http",
			TargetAddress: "127.0.0.1:1234",
		})
		if proxyErr != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, proxyErr)
		}
		if port == 0 {
			t.Fatalf("iteration %d: expected non-zero port", i)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn, dialErr := dialer.DialContext(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", port))
		cancel()
		if dialErr != nil {
			t.Fatalf("iteration %d: dial failed: %v", i, dialErr)
		}
		_ = conn.Close()
	}
}

func TestHTTPReverseProxy_ExplicitPort(t *testing.T) {
	freePort, err := utils.GetFreePort()
	if err != nil {
		t.Fatalf("getting free port: %v", err)
	}

	port, proxyErr := HTTPReverseProxy(Options{
		ProxyPort:     freePort,
		TargetScheme:  "http",
		TargetAddress: "127.0.0.1:1234",
	})
	if proxyErr != nil {
		t.Fatalf("unexpected error: %v", proxyErr)
	}
	if port != freePort {
		t.Errorf("expected port %d, got %d", freePort, port)
	}
}
