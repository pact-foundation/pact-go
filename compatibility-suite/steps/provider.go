package steps

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

// providerVersion is the provider version reported to the broker, as in the
// reference implementation.
const providerVersion = "1.2.3"

// stateChangeFailed is the verifier's error when a state setup call fails.
const stateChangeFailed = "One or more of the setup state change handlers has failed"

// verificationErrorKinds maps the suite's error names to the mismatch types
// that produce them, as the reference implementation does.
var verificationErrorKinds = map[string]string{
	"Response status did not match": "StatusMismatch",
	"Headers had differences":       "HeaderMismatch",
	"Body had differences":          "BodyMismatch",
}

func registerProvider(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^a provider is started that returns the response from interaction (\d+)$`, w.startProviderWith)
	sc.Step(`^a provider is started that returns the responses from interactions "([^"]*)"$`, w.startProviderWithList)
	sc.Step(`^a provider is started that returns the response from interaction (\d+), with the following changes:$`, w.startChangedProvider)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified$`, w.verifyPactFile)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified with a provider state "([^"]*)" defined$`, w.verifyPactFileInState)
	sc.Step(`^a Pact file for interaction (\d+) is to be verified from a Pact broker$`, w.verifyPactFromBroker)
	sc.Step(`^publishing of verification results is enabled$`, w.enablePublishing)
	sc.Step(`^a provider state callback is configured$`, w.configureStateCallback)
	sc.Step(`^a provider state callback is configured, but will return a failure$`, w.configureFailingStateCallback)
	sc.Step(`^a request filter is configured to make the following changes:$`, w.configureRequestFilter)
	sc.Step(`^the verification is run$`, w.runVerification)
	sc.Step(`^the verification will be successful$`, w.verificationPassed)
	sc.Step(`^the verification will NOT be successful$`, w.verificationFailed)
	sc.Step(`^the verification results will contain a "([^"]*)" error$`, w.verificationErrorIs)
	sc.Step(`^a verification result will NOT be published back$`, w.nothingPublished)
	sc.Step(`^a successful verification result will be published back for interaction \{(\d+)\}$`, w.publishedSuccess)
	sc.Step(`^a failed verification result will be published back for the interaction \{(\d+)\}$`, w.publishedFailure)
	sc.Step(`^the provider state callback will be called before the verification is run$`, w.stateSetupCalled)
	sc.Step(`^the provider state callback will receive a setup call with "([^"]*)" as the provider state parameter$`, w.stateSetupCalledFor)
	sc.Step(`^the provider state callback will be called after the verification is run$`, w.stateTeardownCalled)
	sc.Step(`^the provider state callback will receive a teardown call "([^"]*)" as the provider state parameter$`, w.stateTeardownCalledFor)
	sc.Step(`^the provider state callback will NOT receive a teardown call$`, w.stateTeardownNotCalled)
	sc.Step(`^a warning will be displayed that there was no provider state callback configured for provider state "([^"]*)"$`, w.noStateCallbackWarning)
	sc.Step(`^the request to the provider will contain the header "([^"]*)"$`, w.providerRequestsHaveHeader)
}

func (w *world) startProvider(interactions []model.Interaction) error {
	if w.provider != nil {
		return errors.New("a provider is already running")
	}
	w.provider = fakes.NewProvider(interactions)

	return nil
}

func (w *world) startProviderWith(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.startProvider([]model.Interaction{in})
}

func (w *world) startProviderWithList(list string) error {
	numbers, err := parseNumbers(list)
	if err != nil {
		return err
	}
	interactions := make([]model.Interaction, 0, len(numbers))
	for _, n := range numbers {
		in, err := w.interaction(n)
		if err != nil {
			return err
		}
		interactions = append(interactions, in)
	}

	return w.startProvider(interactions)
}

func (w *world) startChangedProvider(n int, table *godog.Table) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	err = applyResponseChanges(&in.Response, cells, w.fixtures, true)
	if err != nil {
		return err
	}

	return w.startProvider([]model.Interaction{in})
}

// pactFor renders the pact verifying interaction n, between consumer c_n and
// the provider, optionally in a provider state.
func (w *world) pactFor(n int, state string) ([]byte, error) {
	in, err := w.interaction(n)
	if err != nil {
		return nil, err
	}
	if state != "" {
		in.ProviderStates = []string{state}
	}
	w.states = append(w.states, in.ProviderStates...)

	return pactJSON(fmt.Sprintf("c_%d", n), providerName, []model.Interaction{in})
}

func (w *world) verifyPactFile(n int) error {
	return w.addPactFile(n, "")
}

func (w *world) verifyPactFileInState(n int, state string) error {
	return w.addPactFile(n, state)
}

func (w *world) addPactFile(n int, state string) error {
	pact, err := w.pactFor(n, state)
	if err != nil {
		return err
	}
	path := filepath.Join(w.dir, fmt.Sprintf("c_%d-%s.json", n, providerName))
	err = os.WriteFile(path, pact, 0o600)
	if err != nil {
		return fmt.Errorf("writing the pact file: %w", err)
	}
	w.pactFiles = append(w.pactFiles, path)

	return nil
}

func (w *world) verifyPactFromBroker(n int) error {
	pact, err := w.pactFor(n, "")
	if err != nil {
		return err
	}
	if w.broker == nil {
		w.broker = fakes.NewBroker(providerName)
	}
	w.broker.AddPact(fmt.Sprintf("c_%d", n), pact)

	return nil
}

func (w *world) enablePublishing() error {
	w.publish = true

	return nil
}

func (w *world) configureStateCallback() error {
	w.stateCalls = &stateRecorder{}

	return nil
}

func (w *world) configureFailingStateCallback() error {
	w.stateCalls = &stateRecorder{fail: true}

	return nil
}

func (w *world) configureRequestFilter(table *godog.Table) error {
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	for name, value := range cells {
		if name != headersCell {
			return fmt.Errorf("unsupported request filter change %q", name)
		}
		w.filterHeaders, err = parseHeaders(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *world) runVerification() error {
	if w.provider == nil {
		return errors.New("no provider has been started")
	}

	req := adapter.VerifyRequest{
		ProviderURL:     w.provider.URL(),
		PactFiles:       w.pactFiles,
		ProviderVersion: providerVersion,
		Publish:         w.publish,
		Headers:         w.filterHeaders,
	}
	if w.broker != nil {
		req.Broker = &adapter.BrokerSource{URL: w.broker.URL(), Provider: providerName}
	}
	if w.stateCalls != nil {
		// Interactions without a provider state are set up as state "".
		req.States = append([]string{""}, w.states...)
		req.StateCallback = w.stateCalls.call
	}

	result, err := w.adapter.Verifier.Verify(req)
	if err != nil {
		return fmt.Errorf("running the verification: %w", err)
	}
	w.result = &result

	return nil
}

func (w *world) verified() (adapter.VerifyResult, error) {
	if w.result == nil {
		return adapter.VerifyResult{}, errors.New("the verification has not been run")
	}

	return *w.result, nil
}

func (w *world) verificationPassed() error {
	result, err := w.verified()
	if err != nil {
		return err
	}
	if result.Err != nil {
		return fmt.Errorf("want the verification to pass, got %w: %+v", result.Err, result.Errors)
	}

	return nil
}

func (w *world) verificationFailed() error {
	result, err := w.verified()
	if err != nil {
		return err
	}
	if result.Err == nil {
		return errors.New("want the verification to fail, but it passed")
	}

	return nil
}

func (w *world) verificationErrorIs(name string) error {
	result, err := w.verified()
	if err != nil {
		return err
	}

	kind, isMismatch := verificationErrorKinds[name]
	message := name
	if name == "State change request failed" {
		message = stateChangeFailed
	}
	for _, e := range result.Errors {
		if isMismatch && slices.Contains(e.MismatchTypes, kind) || !isMismatch && e.Message == message {
			return nil
		}
	}

	return fmt.Errorf("want a %q error, got %+v", name, result.Errors)
}

func (w *world) nothingPublished() error {
	if w.broker == nil {
		return nil
	}
	if published := w.broker.Publications(); len(published) > 0 {
		return fmt.Errorf("want nothing published, got %+v", published)
	}

	return nil
}

func (w *world) publishedSuccess(n int) error {
	return w.publishedResult(n, true)
}

func (w *world) publishedFailure(n int) error {
	return w.publishedResult(n, false)
}

func (w *world) publishedResult(n int, success bool) error {
	if w.broker == nil {
		return errors.New("no broker was used")
	}
	consumer := fmt.Sprintf("c_%d", n)
	published := w.broker.Publications()
	for _, p := range published {
		got, ok := p.Body["success"].(bool)
		if p.Consumer == consumer && ok && got == success {
			return nil
		}
	}

	return fmt.Errorf("want a result with success=%t published for %s, got %+v", success, consumer, published)
}

// stateCall is one call to the provider state callback.
type stateCall struct {
	setup bool
	state string
}

// stateRecorder is a provider state callback that records its calls, and
// fails every call when fail is set.
type stateRecorder struct {
	mu    sync.Mutex
	fail  bool
	calls []stateCall
}

func (r *stateRecorder) call(setup bool, state string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, stateCall{setup: setup, state: state})
	if r.fail {
		return errors.New("the provider state callback was configured to fail")
	}

	return nil
}

// called reports whether the callback received a matching call.
func (w *world) called(match func(stateCall) bool) (bool, []stateCall, error) {
	if w.stateCalls == nil {
		return false, nil, errors.New("no provider state callback is configured")
	}
	w.stateCalls.mu.Lock()
	defer w.stateCalls.mu.Unlock()
	calls := slices.Clone(w.stateCalls.calls)

	return slices.ContainsFunc(calls, match), calls, nil
}

func (w *world) expectCall(what string, match func(stateCall) bool) error {
	found, calls, err := w.called(match)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("want %s, got calls %+v", what, calls)
	}

	return nil
}

func (w *world) stateSetupCalled() error {
	return w.expectCall("a setup call", func(c stateCall) bool { return c.setup })
}

func (w *world) stateSetupCalledFor(state string) error {
	return w.expectCall(fmt.Sprintf("a setup call for %q", state), func(c stateCall) bool { return c.setup && c.state == state })
}

func (w *world) stateTeardownCalled() error {
	return w.expectCall("a teardown call", func(c stateCall) bool { return !c.setup })
}

func (w *world) stateTeardownCalledFor(state string) error {
	return w.expectCall(fmt.Sprintf("a teardown call for %q", state), func(c stateCall) bool { return !c.setup && c.state == state })
}

func (w *world) stateTeardownNotCalled() error {
	found, calls, err := w.called(func(c stateCall) bool { return !c.setup })
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("want no teardown call, got %+v", calls)
	}

	return nil
}

// noStateCallbackWarning checks what the reference implementations check:
// that the verification ran with a provider state and no callback. The
// warning itself goes to the verifier's log, which neither they nor pact-go
// expose.
func (w *world) noStateCallbackWarning(state string) error {
	_, err := w.verified()
	if err != nil {
		return err
	}
	if w.stateCalls != nil {
		return errors.New("a provider state callback was configured")
	}
	if !slices.Contains(w.states, state) {
		return fmt.Errorf("no pact declared the provider state %q", state)
	}

	return nil
}

func (w *world) providerRequestsHaveHeader(header string) error {
	if w.provider == nil {
		return errors.New("no provider has been started")
	}
	name, value, ok := strings.Cut(header, ":")
	if !ok {
		return fmt.Errorf("header %q has no ':'", header)
	}
	name, value = strings.TrimSpace(name), strings.TrimSpace(value)

	requests := w.provider.Requests()
	if len(requests) == 0 {
		return errors.New("the provider received no requests")
	}
	for _, r := range requests {
		if r.Header.Get(name) != value {
			return fmt.Errorf("want %s: %s on every request, got %v on %s %s", name, value, r.Header, r.Method, r.Path)
		}
	}

	return nil
}
