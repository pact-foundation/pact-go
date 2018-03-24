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
			Body: `{
			"foo": "bar",
			"baz": "bat"
			}`,
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Request.Body.(map[string]interface{}); !ok {
		t.Fatalf("Expected response to be of type 'map[string]string'")
	}

	if !reflect.DeepEqual(i.Request.Body, expect) {
		t.Fatalf("Expected response object body '%v' to match '%v'", i.Request.Body, expect)
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
			Body: `{
				"foo": "bar",
				"baz": "bat"
			}`,
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Response.Body.(map[string]interface{}); !ok {
		t.Fatalf("Expected response to be of type 'map[string]string'")
	}

	if !reflect.DeepEqual(i.Response.Body, expect) {
		t.Fatalf("Expected response object body '%v' to match '%v'", i.Response.Body, expect)
	}
}

func TestInteraction_toObject(t *testing.T) {
	// unstructured string should not be changed
	res := toObject("somestring")
	content, ok := res.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// errors should return a string repro of original interface{}
	res = toObject("")
	content, ok = res.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "" {
		t.Fatalf("Expected '' but got '%s'", content)
	}
}
