package steps

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/adapter"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestPactJSONWritesV2Interaction(t *testing.T) {
	in := model.Interaction{
		Number: 2, Description: "Interaction 1", ProviderStates: []string{"state one"},
		Request:  model.Request{Method: "GET", Path: "/with_params", Query: url.Values{"a": {"1"}, "b": {"2"}}},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "application/json", Contents: []byte(`{"one": "a"}`)}},
	}
	data, err := pactJSON("c_2", "p", []model.Interaction{in})
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Consumer struct {
			Name string `json:"name"`
		} `json:"consumer"`
		Provider struct {
			Name string `json:"name"`
		} `json:"provider"`
		Interactions []struct {
			Description   string `json:"description"`
			ProviderState string `json:"providerState"`
			Request       struct {
				Method string `json:"method"`
				Path   string `json:"path"`
				Query  string `json:"query"`
			} `json:"request"`
			Response struct {
				Status  int               `json:"status"`
				Headers map[string]string `json:"headers"`
				Body    map[string]string `json:"body"`
			} `json:"response"`
		} `json:"interactions"`
		Metadata struct {
			PactSpecification struct {
				Version string `json:"version"`
			} `json:"pactSpecification"`
		} `json:"metadata"`
	}
	err = json.Unmarshal(data, &doc)
	if err != nil {
		t.Fatal(err)
	}
	got := doc.Interactions[0]
	if doc.Consumer.Name != "c_2" || doc.Provider.Name != "p" || doc.Metadata.PactSpecification.Version != "2.0.0" ||
		got.Description != "Interaction 1" || got.ProviderState != "state one" ||
		got.Request.Method != http.MethodGet || got.Request.Path != "/with_params" || got.Request.Query != "a=1&b=2" ||
		got.Response.Status != 200 || got.Response.Headers["Content-Type"] != "application/json" || got.Response.Body["one"] != "a" {
		t.Fatalf("got %s", data)
	}
}

func TestPactJSONWritesTextBodiesAsStrings(t *testing.T) {
	in := model.Interaction{
		Request:  model.Request{Method: "GET", Path: "/form"},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "application/x-www-form-urlencoded", Contents: []byte("a=1&b=2")}},
	}
	data, err := pactJSON("c", "p", []model.Interaction{in})
	if err != nil {
		t.Fatal(err)
	}
	var pact pactFile
	err = json.Unmarshal(data, &pact)
	if err != nil {
		t.Fatal(err)
	}
	var body string
	err = json.Unmarshal(pact.Interactions[0].Response.Body, &body)
	if err != nil || body != "a=1&b=2" {
		t.Fatalf("want a string body, got %s", pact.Interactions[0].Response.Body)
	}
}

func TestPactJSONRejectsBinaryBodies(t *testing.T) {
	in := model.Interaction{
		Request:  model.Request{Method: "GET", Path: "/bin"},
		Response: model.Response{Status: 200, Body: &model.Body{ContentType: "image/jpeg", Contents: []byte{0xff, 0xd8, 0x00}}},
	}
	_, err := pactJSON("c", "p", []model.Interaction{in})
	if !errors.Is(err, adapter.ErrUnsupported) {
		t.Fatalf("want ErrUnsupported, got %v", err)
	}
}

func TestPactPartHeadersAndQuery(t *testing.T) {
	var v2, v3 pactPart
	err := json.Unmarshal([]byte(`{"headers":{"x-test":"Compatibility"},"query":"a=1&b=2"}`), &v2)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(`{"headers":{"X-Test":["Compatibility"]},"query":{"a":["1"],"b":["2"]}}`), &v3)
	if err != nil {
		t.Fatal(err)
	}
	for name, part := range map[string]pactPart{"V2": v2, "V3": v3} {
		value, ok := part.header("X-TEST")
		if !ok || value != "Compatibility" {
			t.Errorf("%s header: %q %v", name, value, ok)
		}
		query, err := part.query()
		if err != nil || query.Get("a") != "1" || query.Get("b") != "2" {
			t.Errorf("%s query: %v %v", name, query, err)
		}
	}
	_, ok := v2.header("Missing")
	if ok {
		t.Fatal("want no value for a missing header")
	}
}

func TestAddPactBodyKeepsAContentTypeHeaderWhateverItsCase(t *testing.T) {
	part := map[string]any{}
	body := &model.Body{ContentType: "application/json", Contents: []byte(`{}`)}
	err := addPactBody(part, http.Header{"content-type": {"text/plain"}}, body)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := part["headers"].(map[string]string)
	if len(got) != 1 || got["content-type"] != "text/plain" {
		t.Fatalf("want only the content-type given, got %v", got)
	}
}
