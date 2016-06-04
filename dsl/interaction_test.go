package dsl

import "testing"

func TestInteraction_NewInteraction(t *testing.T) {
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})

	if i.State != "Some state" {
		t.Fatalf("Expected 'Some state' but got '%s'", i.State)
	}
	if i.Description != "Some name for the test" {
		t.Fatalf("Expected 'Some name for the test' but got '%s'", i.Description)
	}
}
