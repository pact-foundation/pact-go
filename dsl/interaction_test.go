package dsl

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInteraction_NewInteraction(t *testing.T) {
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	if i.State != "Some state" {
		t.Fatalf("Expected 'Some state' but got '%s'", i.State)
	}
	if i.Description != "Some name for the test" {
		t.Fatalf("Expected 'Some name for the test' but got '%s'", i.Description)
	}
}

func TestInteraction_WithRequest(t *testing.T) {
	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Body: "somestring",
		})

	content, ok := i.Request.Body.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// structured string should be changed to an interface{}
	i = (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Body: map[string]string{
				"foo": "bar",
				"baz": "bat",
			},
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Request.Body.(map[string]string); !ok {
		t.Fatal("Expected response to be of type 'map[string]string', but got", reflect.TypeOf(i.Request.Body))
	}
}

func TestInteraction_WillRespondWith(t *testing.T) {
	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{
			Body: "somestring",
		})

	content, ok := i.Response.Body.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// structured string should be changed to an interface{}
	i = (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{
			Body: map[string]string{
				"foo": "bar",
				"baz": "bat",
			},
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Response.Body.(map[string]string); !ok {
		t.Fatal("Expected response to be of type 'map[string]string', but got", reflect.TypeOf(i.Response.Body))
	}
}

func TestInteraction_isStringLikeObject(t *testing.T) {
	testCases := map[string]bool{
		"somestring":    false,
		"":              false,
		`{"foo":"bar"}`: true,
	}

	for testCase, want := range testCases {
		if isJsonFormattedObject(testCase) != want {
			t.Fatal("want", want, "got", !want, "for test case", testCase)
		}
	}
}
