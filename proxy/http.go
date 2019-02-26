package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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

	// ProxyPort is the port to make available for proxying
	// Defaults to a random port
	ProxyPort int

	// Middleware to apply to the Proxy
	Middleware []Middleware
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
	port := options.ProxyPort
	var err error

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: options.TargetScheme,
		Host:   options.TargetAddress,
	})

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
