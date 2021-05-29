package v3

// import "testing"

// type t struct {
// 	ID int
// }

// func TestMessage_DSL(t *testing.T) {
// 	m := &Message{}
// 	m.Given(ProviderStateV3{
// 		Name: "some state",
// 	}).
// 		ExpectsToReceive("description string").
// 		WithMetadata(MapMatcher{
// 			"content-type": String("application/json"),
// 		}).
// 		WithContent(map[string]interface{}{
// 			"foo": "bar",
// 		}).
// 		AsType(t)
// }
