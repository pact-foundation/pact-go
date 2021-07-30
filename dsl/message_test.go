package dsl

import "testing"

func TestMessage_DSL(t *testing.T) {
	m := &Message{}
	m.Given("state string").
		ExpectsToReceive("description string").
		WithMetadata(MapMatcher{
			"content-type": String("application/json"),
		}).
		WithContent(map[string]interface{}{
			"foo": "bar",
		}).
		AsType(t)
}
