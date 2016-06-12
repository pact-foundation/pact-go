package provider

import (
	"time"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware wraps our service and logs stuff about our services.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

// Concrete implementation of the Logging Middleware.
type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

// Login logs stuff about our login process.
func (mw loggingMiddleware) Login(u string, p string) (user *User, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Login", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Login(u, p)
}
