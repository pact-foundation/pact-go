package consumer

import "github.com/pact-foundation/pact-go/v2/matchers"

// Response is the default implementation of the Response interface.
type Response struct {
	Status  int                 `json:"status"`
	Headers matchers.MapMatcher `json:"headers,omitempty"`
	Body    interface{}         `json:"body,omitempty"`
}
