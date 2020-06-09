package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/pact-foundation/pact-go/utils"
)

// Middleware is a way to use composition to add functionality
// by intercepting the req/response cycle of the Reverse Proxy.
// Each handler must accept an http.Handler and also return an
// http.Handler, allowing a simple way to chain functionality together
type Middleware func(http.Handler) http.Handler

// Options for the Reverse Proxy configuration
type Options struct {

	// TargetScheme is one of 'http' or 'https'
	TargetScheme string

	// TargetAddress is the host:port component to proxy
	TargetAddress string

	// TargetPath is the path on the target to proxy
	TargetPath string

	// ProxyPort is the port to make available for proxying
	// Defaults to a random port
	ProxyPort int

	// Middleware to apply to the Proxy
	Middleware []Middleware

	// Internal request prefix for proxy to not rewrite
	InternalRequestPathPrefix string

	// Custom TLS Configuration for communicating with a Provider
	// Useful when verifying self-signed services, MASSL etc.
	CustomTLSConfig *tls.Config
}

// loggingMiddleware logs requests to the proxy
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] http reverse proxy received connection from %s on path %s\n", r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// chainHandlers takes a set of middleware and joins them together
// into a single Middleware, making it much simpler to compose middleware
// together
func chainHandlers(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last.ServeHTTP(w, r)
		})
	}
}

// HTTPReverseProxy provides a default setup for proxying
// internal components within the framework
func HTTPReverseProxy(options Options) (int, error) {
	log.Println("[DEBUG] starting new proxy with opts", options)
	port := options.ProxyPort
	var err error

	url := &url.URL{
		Scheme: options.TargetScheme,
		Host:   options.TargetAddress,
		Path:   options.TargetPath,
	}

	proxy := createProxy(url, options.InternalRequestPathPrefix)
	proxy.Transport = customTransport{tlsConfig: options.CustomTLSConfig}

	if port == 0 {
		port, err = utils.GetFreePort()
		if err != nil {
			log.Println("[ERROR] unable to start reverse proxy server:", err)
			return 0, err
		}
	}

	wrapper := chainHandlers(append(options.Middleware, loggingMiddleware)...)

	log.Println("[DEBUG] starting reverse proxy on port", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), wrapper(proxy))

	return port, nil
}

// https://stackoverflow.com/questions/52986853/how-to-debug-httputil-newsinglehostreverseproxy
// Set the proxy.Transport field to an implementation that dumps the request before delegating to the default transport:

type customTransport struct {
	tlsConfig *tls.Config
}

func (c customTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	log.Println("[TRACE] proxy outgoing request\n", string(b))

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if c.tlsConfig != nil {
		log.Println("[DEBUG] applying custom TLS config")
		transport.TLSClientConfig = c.tlsConfig
	}
	var DefaultTransport http.RoundTripper = transport

	res, err := DefaultTransport.RoundTrip(r)
	if err != nil {
		log.Println("[ERROR]", err)
		return nil, err
	}
	b, err = httputil.DumpResponse(res, true)
	log.Println("[TRACE] proxied server response\n", string(b))

	return res, err
}

// Adapted from https://github.com/golang/go/blob/master/src/net/http/httputil/reverseproxy.go
func createProxy(target *url.URL, ignorePrefix string) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		if !strings.HasPrefix(req.URL.Path, ignorePrefix) {
			log.Println("[DEBUG] setting proxy to target")
			log.Println("[DEBUG] incoming request", req.URL)
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host

			req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
			log.Println("[DEBUG] outgoing request to target", req.URL)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "Pact Go")
			}
		} else {
			log.Println("[DEBUG] setting proxy to internal server")
			req.URL.Scheme = "http"
			req.URL.Host = "localhost"
			req.Host = "localhost"
		}
	}
	return &httputil.ReverseProxy{Director: director}
}

// From httputil package
// https://github.com/golang/go/blob/master/src/net/http/httputil/reverseproxy.go
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
