package steps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/cucumber/godog"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

const (
	consumerName = "compatibility-suite-c"
	providerName = "p"
)

func registerConsumer(sc *godog.ScenarioContext, w *world) {
	sc.Step(`^the mock server is started with interaction (\d+)$`, w.startMockServerWith)
	sc.Step(`^the mock server is started with interactions "([^"]*)"$`, w.startMockServerWithList)
	sc.Step(`^request (\d+) is made to the mock server$`, w.makeRequest)
	sc.Step(`^request (\d+) is made to the mock server with the following changes:$`, w.makeChangedRequest)
	sc.Step(`^a (\d+) (?:success|error) response is returned$`, w.responseStatusIs)
	sc.Step(`^the payload will contain the "([^"]*)" JSON document$`, w.payloadIsJSONDocument)
	sc.Step(`^the content type will be set as "([^"]*)"$`, w.responseContentTypeIs)
	sc.Step(`^the pact test is done$`, w.finishPactTest)
	sc.Step(`^the mock server status will be OK$`, w.mockServerOK)
	sc.Step(`^the mock server status will NOT be OK$`, w.mockServerNotOK)
	sc.Step(`^the mock server status will be mismatches$`, w.mockServerHasMismatches)
	sc.Step(`^the mock server will write out a Pact file for the interactions? when done$`, w.pactFileWritten)
	sc.Step(`^the mock server will NOT write out a Pact file for the interactions? when done$`, w.pactFileNotWritten)
	sc.Step(`^the pact file will contain \{(\d+)\} interactions?$`, w.pactInteractionCount)
	sc.Step(`^the \{(\w+)\} interaction request will be for a "([^"]*)"$`, w.pactRequestMethodIs)
	sc.Step(`^the \{(\w+)\} interaction request query parameters will be "([^"]*)"$`, w.pactRequestQueryIs)
	sc.Step(`^the \{(\w+)\} interaction request will contain the header "([^"]*)" with value "([^"]*)"$`, w.pactRequestHeaderIs)
	sc.Step(`^the \{(\w+)\} interaction request content type will be "([^"]*)"$`, w.pactRequestContentTypeIs)
	sc.Step(`^the \{(\w+)\} interaction request will contain the "([^"]*)" document$`, w.pactRequestDocumentIs)
	sc.Step(`^the \{(\w+)\} interaction response will contain the "([^"]*)" document$`, w.pactResponseDocumentIs)
	sc.Step(`^the mock server status will be an expected but not received error for interaction \{(\d+)\}$`, w.missingRequestFor)
	sc.Step(`^the mock server status will be an unexpected "([^"]*)" request received error for interaction \{(\d+)\}$`, w.unexpectedRequestFor)
	sc.Step(`^the mock server status will be an unexpected "([^"]*)" request received error for path "([^"]*)"$`, w.unexpectedRequestAt)
	sc.Step(`^the mismatches will contain a "([^"]*)" mismatch with error "([^"]*)"$`, w.mismatchWithError)
	sc.Step(`^the mismatches will contain a "([^"]*)" mismatch with path "([^"]*)" with error "([^"]*)"$`, w.mismatchAtPathWithError)
}

func (w *world) startMockServerWith(n int) error {
	return w.startMockServer([]int{n})
}

func (w *world) startMockServerWithList(list string) error {
	numbers, err := parseNumbers(list)
	if err != nil {
		return err
	}

	return w.startMockServer(numbers)
}

func (w *world) startMockServer(numbers []int) error {
	selected := make([]model.Interaction, 0, len(numbers))
	for _, n := range numbers {
		in, err := w.interaction(n)
		if err != nil {
			return err
		}
		selected = append(selected, in)
	}

	w.pactDir = filepath.Join(w.dir, "pacts")
	mock, err := w.adapter.Consumer.Start(adapter.ConsumerPact{
		Consumer: consumerName, Provider: providerName, PactDir: w.pactDir,
	}, selected)
	if err != nil {
		return fmt.Errorf("starting the mock server: %w", err)
	}
	w.mock = mock

	return nil
}

func (w *world) makeRequest(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.send(in.Request)
}

func (w *world) makeChangedRequest(n int, table *godog.Table) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}
	cells, err := changes(rows(table))
	if err != nil {
		return err
	}
	err = applyRequestChanges(&in.Request, cells, w.fixtures)
	if err != nil {
		return err
	}

	return w.send(in.Request)
}

func (w *world) send(req model.Request) error {
	if w.mock == nil {
		return errors.New("the mock server has not been started")
	}
	//nolint:bodyclose // sendRequest has already read and closed the body.
	res, body, err := sendRequest(w.mock.URL(), req)
	if err != nil {
		return err
	}
	w.response, w.responseBody = res, body

	return nil
}

func (w *world) responseStatusIs(status int) error {
	if w.response == nil {
		return errors.New("no request has been made")
	}
	if w.response.StatusCode != status {
		return fmt.Errorf("want status %d, got %d: %s", status, w.response.StatusCode, w.responseBody)
	}

	return nil
}

func (w *world) payloadIsJSONDocument(name string) error {
	want, err := w.fixtures.Read(name + ".json")
	if err != nil {
		return fmt.Errorf("reading the %q document: %w", name, err)
	}

	return jsonEqual(w.responseBody, want)
}

func (w *world) responseContentTypeIs(contentType string) error {
	if w.response == nil {
		return errors.New("no request has been made")
	}
	got := w.response.Header.Get("Content-Type")
	if got != contentType {
		return fmt.Errorf("want content type %q, got %q", contentType, got)
	}

	return nil
}

func (w *world) finishPactTest() error {
	if w.mock == nil {
		return errors.New("the mock server has not been started")
	}
	outcome := w.mock.Finish()
	w.outcome = &outcome

	return nil
}

func (w *world) finished() (adapter.Outcome, error) {
	if w.outcome == nil {
		return adapter.Outcome{}, errors.New("the pact test is not done")
	}

	return *w.outcome, nil
}

func (w *world) mockServerOK() error {
	outcome, err := w.finished()
	if err != nil {
		return err
	}
	if outcome.Err != nil {
		return fmt.Errorf("want the pact test to pass, but it failed: %w", outcome.Err)
	}

	return nil
}

func (w *world) mockServerNotOK() error {
	outcome, err := w.finished()
	if err != nil {
		return err
	}
	if outcome.Err == nil {
		return errors.New("want the pact test to fail, but it passed")
	}

	return nil
}

func (w *world) mockServerHasMismatches() error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	if len(mismatches) == 0 {
		return errors.New("want mismatches, got none")
	}

	return nil
}

// mismatches are the mock server's mismatches: final once the pact test is
// done, and as recorded so far before then.
func (w *world) mismatches() ([]model.Mismatch, error) {
	if w.outcome != nil {
		return w.outcome.Mismatches, nil
	}
	if w.mock == nil {
		return nil, errors.New("the mock server has not been started")
	}

	mismatches, err := w.mock.Mismatches()
	if err != nil {
		return nil, fmt.Errorf("reading mismatches: %w", err)
	}

	return mismatches, nil
}

func (w *world) pactFilePath() string {
	return filepath.Join(w.pactDir, consumerName+"-"+providerName+".json")
}

func (w *world) pactFileWritten() error {
	_, err := w.finished()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(w.pactFilePath())
	if err != nil {
		return fmt.Errorf("want a pact file: %w", err)
	}

	var pact pactFile
	err = json.Unmarshal(data, &pact)
	if err != nil {
		return fmt.Errorf("parsing the pact file: %w", err)
	}
	w.pact = &pact

	return nil
}

func (w *world) pactFileNotWritten() error {
	_, err := w.finished()
	if err != nil {
		return err
	}

	_, err = os.Stat(w.pactFilePath())
	if err == nil {
		return fmt.Errorf("want no pact file, but %s was written", w.pactFilePath())
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("checking for a pact file: %w", err)
	}

	return nil
}

func (w *world) pactInteraction(word string) (pactInteraction, error) {
	if w.pact == nil {
		return pactInteraction{}, errors.New("no pact file has been read")
	}
	i, err := ordinal(word)
	if err != nil {
		return pactInteraction{}, err
	}
	if i >= len(w.pact.Interactions) {
		return pactInteraction{}, fmt.Errorf("the pact file has %d interactions, not a %s", len(w.pact.Interactions), word)
	}

	return w.pact.Interactions[i], nil
}

func (w *world) pactInteractionCount(n int) error {
	if w.pact == nil {
		return errors.New("no pact file has been read")
	}
	if len(w.pact.Interactions) != n {
		return fmt.Errorf("want %d interactions in the pact file, got %d", n, len(w.pact.Interactions))
	}

	return nil
}

func (w *world) pactRequestMethodIs(word, method string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	if !strings.EqualFold(in.Request.Method, method) {
		return fmt.Errorf("want a %s request, got %s", method, in.Request.Method)
	}

	return nil
}

func (w *world) pactRequestQueryIs(word, query string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	got, err := in.Request.query()
	if err != nil {
		return err
	}
	want, err := parseQueryCell(query)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return fmt.Errorf("want query %v, got %v", want, got)
	}

	return nil
}

func (w *world) pactRequestHeaderIs(word, name, value string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}
	got, ok := in.Request.header(name)
	if !ok || got != value {
		return fmt.Errorf("want header %s: %s, got %q (present: %v)", name, value, got, ok)
	}

	return nil
}

func (w *world) pactRequestContentTypeIs(word, contentType string) error {
	return w.pactRequestHeaderIs(word, "Content-Type", contentType)
}

func (w *world) pactRequestDocumentIs(word, name string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}

	return w.documentIs(in.Request.Body, name)
}

func (w *world) pactResponseDocumentIs(word, name string) error {
	in, err := w.pactInteraction(word)
	if err != nil {
		return err
	}

	return w.documentIs(in.Response.Body, name)
}

// documentIs checks a pact file body against a fixture: JSON fixtures as
// JSON, anything else as text.
func (w *world) documentIs(body json.RawMessage, name string) error {
	want, err := w.fixtures.File(name, "")
	if err != nil {
		return fmt.Errorf("reading the %q document: %w", name, err)
	}
	if strings.HasSuffix(name, ".json") {
		return jsonEqual(body, want.Contents)
	}

	var got string
	err = json.Unmarshal(body, &got)
	if err != nil {
		return fmt.Errorf("want a text body, got %s", body)
	}
	if got != string(want.Contents) {
		return fmt.Errorf("want body %q, got %q", want.Contents, got)
	}

	return nil
}

func (w *world) missingRequestFor(n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.findMismatch("missing-request", in.Request.Method, in.Request.Path, func(model.Mismatch) bool { return true })
}

func (w *world) unexpectedRequestFor(method string, n int) error {
	in, err := w.interaction(n)
	if err != nil {
		return err
	}

	return w.findMismatch("request-not-found", method, in.Request.Path, func(m model.Mismatch) bool {
		return sameQuery(m.Query, in.Request.Query)
	})
}

func (w *world) unexpectedRequestAt(method, path string) error {
	return w.findMismatch("request-not-found", method, path, func(model.Mismatch) bool { return true })
}

func (w *world) findMismatch(kind, method, path string, also func(model.Mismatch) bool) error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	for _, m := range mismatches {
		if m.Type == kind && m.Method == method && m.Path == path && also(m) {
			return nil
		}
	}

	return fmt.Errorf("want a %s mismatch for %s %s, got %+v", kind, method, path, mismatches)
}

func sameQuery(a, b map[string][]string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	return reflect.DeepEqual(a, b)
}

func (w *world) mismatchWithError(kind, message string) error {
	return w.findDetail(kind, nil, message)
}

func (w *world) mismatchAtPathWithError(kind, location, message string) error {
	return w.findDetail(kind, &location, message)
}

// findDetail looks for a request mismatch detail of kind ("query",
// "header", "body" or "body-content-type") whose message contains message,
// and, when location is set, at that query parameter, header or body path.
func (w *world) findDetail(kind string, location *string, message string) error {
	mismatches, err := w.mismatches()
	if err != nil {
		return err
	}
	for _, m := range mismatches {
		if m.Type != "request-mismatch" {
			continue
		}
		for _, d := range m.Details {
			if detailIs(d, kind) && (location == nil || d.Location == *location) && strings.Contains(d.Message, message) {
				return nil
			}
		}
	}

	return fmt.Errorf("want a %s mismatch containing %q, got %+v", kind, message, mismatches)
}

// detailIs matches mismatch types the way the reference implementation does:
// by prefix, so "body" matches BodyMismatch, with "body-content-type"
// meaning BodyTypeMismatch.
func detailIs(d model.MismatchDetail, kind string) bool {
	if kind == "body-content-type" {
		return d.Type == "BodyTypeMismatch"
	}

	return strings.HasPrefix(strings.ToLower(d.Type), kind)
}
