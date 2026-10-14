package v2

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestParseMismatches(t *testing.T) {
	data := []byte(`[
	  {"type":"request-mismatch","method":"PUT","path":"/basic","mismatches":[
	    {"type":"BodyMismatch","path":"$.two","expected":"\"b\"","actual":"\"c\"","mismatch":"Expected 'c' (String) to be equal to 'b' (String)"},
	    {"type":"QueryMismatch","parameter":"b","mismatch":"Expected query parameter 'b' but was missing"},
	    {"type":"HeaderMismatch","key":"X-TEST","mismatch":"Expected a header 'X-TEST' but was missing"}]},
	  {"type":"request-not-found","method":"GET","path":"/with_params","request":{"method":"GET","path":"/with_params","query":{"a":["1"],"b":["2"]}}},
	  {"type":"missing-request","method":"GET","path":"/basic","request":{"method":"GET","path":"/basic","query":"x=1"}}
	]`)

	got, err := parseMismatches(data)
	if err != nil {
		t.Fatal(err)
	}

	want := []model.Mismatch{
		{Type: "request-mismatch", Method: "PUT", Path: "/basic", Details: []model.MismatchDetail{
			{Type: "BodyMismatch", Location: "$.two", Message: "Expected 'c' (String) to be equal to 'b' (String)"},
			{Type: "QueryMismatch", Location: "b", Message: "Expected query parameter 'b' but was missing"},
			{Type: "HeaderMismatch", Location: "X-TEST", Message: "Expected a header 'X-TEST' but was missing"},
		}},
		{Type: "request-not-found", Method: "GET", Path: "/with_params", Query: url.Values{"a": {"1"}, "b": {"2"}}},
		{Type: "missing-request", Method: "GET", Path: "/basic", Query: url.Values{"x": {"1"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %#v\nwant %#v", got, want)
	}
}

func TestParseMismatchesRejectsMalformedJSON(t *testing.T) {
	_, err := parseMismatches([]byte(`{not json`))
	if err == nil {
		t.Fatal("want an error for malformed JSON")
	}
}
