package model_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

func TestInteractionCloneIsDeep(t *testing.T) {
	original := model.Interaction{
		ProviderStates: []string{"a"},
		Request: model.Request{
			Query:   url.Values{"a": {"1"}},
			Headers: http.Header{"X": {"1"}},
			Body:    &model.Body{ContentType: "text/plain", Contents: []byte("hi")},
		},
		Response: model.Response{
			Headers: http.Header{"Y": {"1"}},
			Body:    &model.Body{ContentType: "text/plain", Contents: []byte("ok")},
		},
	}

	c := original.Clone()
	c.ProviderStates[0] = "b"
	c.Request.Query.Set("a", "2")
	c.Request.Headers.Set("X", "2")
	c.Request.Body.Contents[0] = 'H'
	c.Response.Headers.Set("Y", "2")
	c.Response.Body.Contents[0] = 'O'

	if original.ProviderStates[0] != "a" || original.Request.Query.Get("a") != "1" ||
		original.Request.Headers.Get("X") != "1" || string(original.Request.Body.Contents) != "hi" ||
		original.Response.Headers.Get("Y") != "1" || string(original.Response.Body.Contents) != "ok" {
		t.Fatalf("changing the clone changed the original: %+v", original)
	}
}

func TestCloneKeepsNilBodies(t *testing.T) {
	c := model.Interaction{}.Clone()
	if c.Request.Body != nil || c.Response.Body != nil {
		t.Fatalf("want nil bodies, got %+v", c)
	}
}

func TestParseQuery(t *testing.T) {
	cases := map[string]struct {
		raw  string
		want url.Values
	}{
		"absent":       {"", nil},
		"null":         {"null", nil},
		"V2 string":    {`"a=1&b=2"`, url.Values{"a": {"1"}, "b": {"2"}}},
		"V3 map":       {`{"a":["1"],"b":["2","3"]}`, url.Values{"a": {"1"}, "b": {"2", "3"}}},
		"empty string": {`""`, url.Values{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := model.ParseQuery(json.RawMessage(tc.raw))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseQueryRejectsOtherJSON(t *testing.T) {
	_, err := model.ParseQuery(json.RawMessage(`42`))
	if err == nil {
		t.Fatal("want an error for a numeric query")
	}
}
