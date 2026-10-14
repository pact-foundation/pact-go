package v2

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

var basicJSON = []byte(`{"one": "a", "two": "b"}`)

func getBasic() model.Interaction {
	return model.Interaction{
		Number: 1, Description: "get basic",
		Request:  model.Request{Method: http.MethodGet, Path: "/basic"},
		Response: model.Response{Status: http.StatusOK, Body: &model.Body{ContentType: "application/json", Contents: basicJSON}},
	}
}

func putBasic() model.Interaction {
	return model.Interaction{
		Number: 2, Description: "put basic",
		Request:  model.Request{Method: http.MethodPut, Path: "/basic", Body: &model.Body{ContentType: "application/json", Contents: basicJSON}},
		Response: model.Response{Status: http.StatusOK},
	}
}

func startConsumer(t *testing.T, interactions ...model.Interaction) (adapter.MockServer, string) {
	t.Helper()
	dir := t.TempDir()
	server, err := consumerAdapter{}.Start(adapter.ConsumerPact{Consumer: "c", Provider: "p", PactDir: dir}, interactions)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { server.Finish() })

	return server, dir
}

func send(t *testing.T, method, target, contentType, body string) int {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()

	return res.StatusCode
}

func hasDetail(mismatches []model.Mismatch, kind, location string) bool {
	for _, m := range mismatches {
		for _, d := range m.Details {
			if d.Type == kind && d.Location == location {
				return true
			}
		}
	}

	return false
}

func TestConsumerPassesAndWritesPact(t *testing.T) {
	server, dir := startConsumer(t, getBasic())
	status := send(t, http.MethodGet, server.URL()+"/basic", "", "")
	if status != http.StatusOK {
		t.Fatalf("want 200, got %d", status)
	}

	outcome := server.Finish()
	if outcome.Err != nil {
		t.Fatalf("want the test to pass, got %v", outcome.Err)
	}
	_, err := os.Stat(filepath.Join(dir, "c-p.json"))
	if err != nil {
		t.Fatalf("want a pact file: %v", err)
	}
}

func TestConsumerReportsMismatchDetailsWhileRunningAndAfter(t *testing.T) {
	server, dir := startConsumer(t, putBasic())
	status := send(t, http.MethodPut, server.URL()+"/basic", "application/json", `{"one": "a", "two": "c"}`)
	if status != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", status)
	}

	running, err := server.Mismatches()
	if err != nil {
		t.Fatal(err)
	}
	if !hasDetail(running, "BodyMismatch", "$.two") {
		t.Fatalf("want a BodyMismatch at $.two while running, got %+v", running)
	}

	outcome := server.Finish()
	if outcome.Err == nil {
		t.Fatal("want the test to fail")
	}
	if !hasDetail(outcome.Mismatches, "BodyMismatch", "$.two") {
		t.Fatalf("want a BodyMismatch at $.two after finishing, got %+v", outcome.Mismatches)
	}
	_, err = os.Stat(filepath.Join(dir, "c-p.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("want no pact file, got %v", err)
	}
}

func TestConsumerReportsMissingRequests(t *testing.T) {
	server, _ := startConsumer(t, getBasic())
	outcome := server.Finish()
	if outcome.Err == nil {
		t.Fatal("want the test to fail")
	}
	if len(outcome.Mismatches) != 1 || outcome.Mismatches[0].Type != "missing-request" ||
		outcome.Mismatches[0].Path != "/basic" {
		t.Fatalf("want one missing-request for /basic, got %+v", outcome.Mismatches)
	}
}

func TestConsumerFinishIsRepeatable(t *testing.T) {
	server, _ := startConsumer(t, getBasic())
	first := server.Finish()
	second := server.Finish()
	if first.Err == nil || !errors.Is(second.Err, first.Err) {
		t.Fatalf("want the same failure twice, got %v and %v", first.Err, second.Err)
	}
}

func TestConsumerRejectsBinaryBodies(t *testing.T) {
	in := putBasic()
	in.Request.Body = &model.Body{ContentType: "image/jpeg", Contents: []byte{0xff, 0xd8, 0x00, 0x10}}

	_, err := consumerAdapter{}.Start(adapter.ConsumerPact{Consumer: "c", Provider: "p", PactDir: t.TempDir()},
		[]model.Interaction{in})
	if !errors.Is(err, adapter.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}
