package steps

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/fixtures"
	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

var fx = fixtures.Loader{Dir: "../pact-compatibility-suite/fixtures"}

func TestInteractionsFromConsumerTable(t *testing.T) {
	got, err := interactionsFromTable([][]string{
		{"No", "method", "path", "query", "headers", "body", "response", "response content", "response body"},
		{"1", "GET", "/basic", "", "", "", "200", "application/json", "file: basic.json"},
		{"2", "GET", "/with_params", "a=1&b=2", "", "", "200", "", ""},
		{"3", "GET", "/with_headers", "", "'X-TEST: Compatibility'", "", "200", "", ""},
		{"8", "PUT", "/form", "", "", "file: form-post-body.xml", "200", "", ""},
	}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("want 4 interactions, got %d", len(got))
	}
	first := got[0]
	if first.Number != 1 || first.Description != "Interaction 0" || first.Request.Method != http.MethodGet ||
		first.Request.Path != "/basic" || first.Response.Status != 200 ||
		first.Response.Body == nil || first.Response.Body.ContentType != "application/json" || first.Request.Body != nil {
		t.Fatalf("interaction 1: %+v", first)
	}
	if got[1].Request.Query.Get("a") != "1" || got[1].Request.Query.Get("b") != "2" {
		t.Fatalf("interaction 2 query: %v", got[1].Request.Query)
	}
	if got[2].Request.Headers.Get("X-Test") != "Compatibility" {
		t.Fatalf("interaction 3 headers: %v", got[2].Request.Headers)
	}
	if got[3].Request.Body.ContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("interaction 4 body: %+v", got[3].Request.Body)
	}
}

func TestInteractionsFromProviderTableReadsResponseHeaders(t *testing.T) {
	got, err := interactionsFromTable([][]string{
		{"No", "method", "path", "query", "headers", "body", "response", "response headers", "response content", "response body"},
		{"5", "GET", "/basic", "", "", "", "200", "'X-TEST: Something'", "application/json", "file: basic.json"},
	}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Response.Headers.Get("X-Test") != "Something" {
		t.Fatalf("response headers: %v", got[0].Response.Headers)
	}
}

func TestRequestChangesKeepContentTypeForText(t *testing.T) {
	req := model.Request{Body: &model.Body{ContentType: "application/x-www-form-urlencoded", Contents: []byte("a=1")}}
	err := applyRequestChanges(&req, map[string]string{"body": "a=1&b=2&c=33&d=4"}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if req.Body.ContentType != "application/x-www-form-urlencoded" || string(req.Body.Contents) != "a=1&b=2&c=33&d=4" {
		t.Fatalf("body: %+v", req.Body)
	}
}

func TestRequestChangesReplaceHeadersAndQuery(t *testing.T) {
	req := model.Request{Headers: http.Header{"X-Test": {"Compatibility"}}}
	err := applyRequestChanges(&req, map[string]string{"headers": "'X-OTHER: Something'", "query": "a=1&c=3", "method": "HEAD", "path": "/path"}, fx)
	if err != nil {
		t.Fatal(err)
	}
	if req.Headers.Get("X-Test") != "" || req.Headers.Get("X-Other") != "Something" {
		t.Fatalf("headers: %v", req.Headers)
	}
	if req.Query.Get("c") != "3" || req.Method != http.MethodHead || req.Path != "/path" {
		t.Fatalf("request: %+v", req)
	}
}

func TestRequestChangesRejectUnknownColumn(t *testing.T) {
	req := model.Request{}
	err := applyRequestChanges(&req, map[string]string{"colour": "red"}, fx)
	if err == nil || !strings.Contains(err.Error(), "colour") {
		t.Fatalf("want an error naming the column, got %v", err)
	}
}

func TestResponseChangesMergeHeaders(t *testing.T) {
	res := model.Response{Status: 200, Headers: http.Header{"X-Test": {"Something"}}}
	err := applyResponseChanges(&res, map[string]string{"response headers": "'X-TEST: Compatibility'", "response": "400"}, fx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.Headers.Values("X-Test"), []string{"Something", "Compatibility"}) || res.Status != 400 {
		t.Fatalf("response: %+v", res)
	}
}

func TestResponseChangesRejectUnknownColumn(t *testing.T) {
	res := model.Response{}
	err := applyResponseChanges(&res, map[string]string{"colour": "red"}, fx, true)
	if err == nil || !strings.Contains(err.Error(), "colour") {
		t.Fatalf("want an error naming the column, got %v", err)
	}
}

func TestChangesNeedsOneRowOfValues(t *testing.T) {
	_, err := changes([][]string{{"body"}})
	if err == nil {
		t.Fatal("want an error for a table without values")
	}
	got, err := changes([][]string{{" body "}, {" Hello Mars! "}})
	if err != nil || got["body"] != "Hello Mars!" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestParseHeaders(t *testing.T) {
	got, err := parseHeaders("'A: 1', 'B: 2'")
	if err != nil {
		t.Fatal(err)
	}
	if got.Get("A") != "1" || got.Get("B") != "2" {
		t.Fatalf("got %v", got)
	}
	_, err = parseHeaders("'no colon'")
	if err == nil {
		t.Fatal("want an error for a header without ':'")
	}
}

func TestParseNumbers(t *testing.T) {
	got, err := parseNumbers("1, 2")
	if err != nil || !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got %v %v", got, err)
	}
	_, err = parseNumbers("1, x")
	if err == nil {
		t.Fatal("want an error for a non-number")
	}
}

func TestOrdinal(t *testing.T) {
	for word, want := range map[string]int{"first": 0, "second": 1, "third": 2} {
		got, err := ordinal(word)
		if err != nil || got != want {
			t.Errorf("%s: got %d %v", word, got, err)
		}
	}
	_, err := ordinal("fourth")
	if err == nil {
		t.Fatal("want an error for an unknown ordinal")
	}
}
