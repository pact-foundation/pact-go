package v2

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

type consumerAdapter struct{}

// Start builds a V2 pact, the oldest specification pact-go supports, and
// runs ExecuteTest in the background until the mock server is finished.
func (consumerAdapter) Start(pact adapter.ConsumerPact, interactions []model.Interaction) (adapter.MockServer, error) {
	for _, in := range interactions {
		err := expressible(in)
		if err != nil {
			return nil, err
		}
	}

	mock, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: pact.Consumer,
		Provider: pact.Provider,
		PactDir:  pact.PactDir,
		LogDir:   pact.PactDir,
	})
	if err != nil {
		return nil, fmt.Errorf("creating the pact: %w", err)
	}
	for _, in := range interactions {
		addInteraction(mock, in)
	}

	server := &mockServer{
		ready:    make(chan consumer.MockServerConfig, 1),
		done:     make(chan struct{}),
		finished: make(chan struct{}),
	}
	go server.run(mock)

	select {
	case server.config = <-server.ready:
		return server, nil
	case <-server.finished:
		return nil, fmt.Errorf("the mock server did not start: %w", server.outcome.Err)
	}
}

// expressible reports an interaction pact-go's public API cannot express.
func expressible(in model.Interaction) error {
	for _, body := range []*model.Body{in.Request.Body, in.Response.Body} {
		if body != nil && bytes.IndexByte(body.Contents, 0) >= 0 {
			return fmt.Errorf("interaction %d has a binary %s body, which pact-go can set only as "+
				"application/octet-stream (BinaryBody) or cut short at its first NUL byte (Body): %w",
				in.Number, body.ContentType, adapter.ErrUnsupported)
		}
	}

	return nil
}

func addInteraction(mock *consumer.V2HTTPMockProvider, in model.Interaction) {
	interaction := mock.AddInteraction()
	for _, state := range in.ProviderStates {
		interaction.Given(state)
	}

	interaction.UponReceiving(in.Description).
		WithRequest(consumer.Method(in.Request.Method), in.Request.Path, func(b *consumer.V2RequestBuilder) {
			for name, values := range in.Request.Query {
				b.Query(name, exactly(values)...)
			}
			for name, values := range in.Request.Headers {
				b.Header(name, exactly(values)...)
			}
			if in.Request.Body != nil {
				b.Body(in.Request.Body.ContentType, in.Request.Body.Contents)
			}
		}).
		WillRespondWith(in.Response.Status, func(b *consumer.V2ResponseBuilder) {
			for name, values := range in.Response.Headers {
				b.Header(name, exactly(values)...)
			}
			if in.Response.Body != nil {
				b.Body(in.Response.Body.ContentType, in.Response.Body.Contents)
			}
		})
}

// exactly wraps values as exact-match values for pact-go's builders.
func exactly(values []string) []matchers.Matcher {
	out := make([]matchers.Matcher, len(values))
	for i, v := range values {
		out[i] = matchers.String(v)
	}

	return out
}

// mockServer runs pact-go's ExecuteTest in a goroutine and holds its test
// function open until Finish, so steps can make requests in between.
type mockServer struct {
	config   consumer.MockServerConfig
	ready    chan consumer.MockServerConfig
	done     chan struct{}
	finished chan struct{}
	stop     sync.Once
	outcome  adapter.Outcome
}

func (s *mockServer) URL() string {
	return "http://" + net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))
}

func (s *mockServer) Mismatches() ([]model.Mismatch, error) {
	select {
	case <-s.finished:
		return s.outcome.Mismatches, nil
	default:
		return readMismatches(s.config.Port)
	}
}

func (s *mockServer) Finish() adapter.Outcome {
	s.stop.Do(func() { close(s.done) })
	<-s.finished

	return s.outcome
}

func (s *mockServer) run(mock *consumer.V2HTTPMockProvider) {
	defer close(s.finished)

	var (
		mismatches []model.Mismatch
		readErr    error
		testErr    error
	)
	sandbox("consumer", func(t *testing.T) {
		t.Helper()

		testErr = mock.ExecuteTest(t, func(config consumer.MockServerConfig) error {
			s.ready <- config
			<-s.done
			// Read before the test function returns: pact-go shuts the mock
			// server down afterwards.
			mismatches, readErr = readMismatches(config.Port)

			return nil
		})
	})

	s.outcome = adapter.Outcome{Err: testErr, Mismatches: mismatches}
	if readErr != nil {
		s.outcome.Err = errors.Join(testErr, readErr)
	}
}
