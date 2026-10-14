package fakes_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fakes"
)

func fetchJSON(t *testing.T, method, target, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var doc map[string]any
	err = json.NewDecoder(res.Body).Decode(&doc)
	if err != nil {
		t.Fatal(err)
	}

	return res.StatusCode, doc
}

// link returns doc._links[rel].href.
func link(t *testing.T, doc map[string]any, rel string) string {
	t.Helper()
	links, _ := doc["_links"].(map[string]any)
	l, _ := links[rel].(map[string]any)
	href, _ := l["href"].(string)
	if href == "" {
		t.Fatalf("no %s link in %v", rel, doc)
	}

	return href
}

func TestBrokerServesPactsForVerification(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	b.AddPact("c_1", []byte(`{"consumer":{"name":"c_1"},"provider":{"name":"p"},"interactions":[]}`))

	_, index := fetchJSON(t, http.MethodGet, b.URL()+"/", "")
	forVerification := strings.ReplaceAll(link(t, index, "pb:provider-pacts-for-verification"), "{provider}", "p")

	status, list := fetchJSON(t, http.MethodPost, forVerification, `{"consumerVersionSelectors":[]}`)
	embedded, _ := list["_embedded"].(map[string]any)
	pacts, _ := embedded["pacts"].([]any)
	if status != http.StatusOK || len(pacts) != 1 {
		t.Fatalf("want one pact, got %d %v", status, list)
	}
	first, _ := pacts[0].(map[string]any)
	self := link(t, first, "self")

	_, pact := fetchJSON(t, http.MethodGet, self, "")
	consumer, _ := pact["consumer"].(map[string]any)
	if consumer["name"] != "c_1" {
		t.Fatalf("want the c_1 pact, got %v", pact)
	}
	if !strings.HasSuffix(link(t, pact, "pb:publish-verification-results"), "/consumer/c_1/verification-results") {
		t.Fatalf("want a publish link for c_1, got %v", pact)
	}
}

func TestBrokerRecordsPublications(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	b.AddPact("c_1", []byte(`{"interactions":[]}`))

	status, _ := fetchJSON(t, http.MethodPost, b.URL()+"/pacts/provider/p/consumer/c_1/verification-results",
		`{"success":true,"providerApplicationVersion":"1.2.3"}`)
	if status != http.StatusCreated {
		t.Fatalf("want 201, got %d", status)
	}

	got := b.Publications()
	if len(got) != 1 || got[0].Consumer != "c_1" || got[0].Body["success"] != true {
		t.Fatalf("got %+v", got)
	}
}

func TestBrokerRejectsUnknownPacts(t *testing.T) {
	b := fakes.NewBroker("p")
	t.Cleanup(b.Close)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, b.URL()+"/pacts/provider/p/consumer/nobody", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", res.StatusCode)
	}
}
