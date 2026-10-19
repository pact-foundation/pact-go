// Package steps defines the compatibility suite's step definitions. It
// drives the API under test only through an adapter.Adapter.
package steps

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// world is the state of one scenario.
type world struct {
	adapter  adapter.Adapter
	fixtures fixtures.Loader
	dir      string

	interactions []model.Interaction

	// consumer scenarios
	mock         adapter.MockServer
	pactDir      string
	response     *http.Response
	responseBody []byte
	outcome      *adapter.Outcome
	pact         *pactFile

	// provider scenarios
	provider      *fakes.Provider
	broker        *fakes.Broker
	pactFiles     []string
	states        []string
	publish       bool
	stateCalls    *stateRecorder
	filterHeaders http.Header
	result        *adapter.VerifyResult
}

// Initialize returns the godog scenario initializer registering every step
// definition, driving the API under test through a.
func Initialize(a adapter.Adapter, fixturesDir string) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &world{adapter: a, fixtures: fixtures.Loader{Dir: fixturesDir}}

		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			dir, err := os.MkdirTemp("", "compatibility-suite-")
			if err != nil {
				return ctx, fmt.Errorf("creating the scenario directory: %w", err)
			}
			w.dir = dir

			return ctx, nil
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			w.close()

			return ctx, nil
		})

		sc.Step(`^the following HTTP interactions have been defined:$`, w.defineInteractions)
		registerConsumer(sc, w)
		registerProvider(sc, w)
	}
}

// close releases everything the scenario started, whether or not it passed.
func (w *world) close() {
	if w.mock != nil {
		w.mock.Finish()
	}
	if w.provider != nil {
		w.provider.Close()
	}
	if w.broker != nil {
		w.broker.Close()
	}
	if w.dir != "" {
		_ = os.RemoveAll(w.dir)
	}
}

func (w *world) defineInteractions(table *godog.Table) error {
	interactions, err := interactionsFromTable(rows(table), w.fixtures)
	if err != nil {
		return err
	}
	w.interactions = interactions

	return nil
}

// interaction returns a copy of the table's interaction n (1-based).
func (w *world) interaction(n int) (model.Interaction, error) {
	if n < 1 || n > len(w.interactions) {
		return model.Interaction{}, fmt.Errorf("there is no interaction %d (%d defined)", n, len(w.interactions))
	}

	return w.interactions[n-1].Clone(), nil
}
