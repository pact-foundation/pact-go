package provider

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/examples/messages/types"
)

var like = dsl.Like
var eachLike = dsl.EachLike
var term = dsl.Term

type s = dsl.String
type request = dsl.Request

var commonHeaders = dsl.MapMatcher{
	"Content-Type": s("application/json; charset=utf-8"),
}

var pact = createPact()

func TestMessageConsumer_UserExists(t *testing.T) {
	message := pact.AddMessage()
	message.
		Given("user with id 127 exists").
		ExpectsToReceive("a user").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"id":   like(127),
			"name": "Baz",
			"access": eachLike(map[string]interface{}{
				"role": term("admin", "admin|controller|user"),
			}, 3),
		}).
		AsType(&types.User{})

	pact.VerifyMessageConsumer(t, message, userHandlerWrapper)
}

func TestMessageConsumer_Order(t *testing.T) {
	message := pact.AddMessage()
	message.
		Given("an order exists").
		ExpectsToReceive("an order").
		WithMetadata(commonHeaders).
		WithContent(dsl.Match(types.Order{})).
		AsType(&types.Order{})

	pact.VerifyMessageConsumer(t, message, orderHandlerWrapper)
}

func TestMessageConsumer_Fail(t *testing.T) {
	t.Skip()
	message := pact.AddMessage()
	message.
		Given("no users").
		ExpectsToReceive("a user").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"foo": "bar",
		})

	pact.VerifyMessageConsumer(t, message, func(m dsl.Message) error {
		t.Logf("[DEBUG] calling message handler func with arguments: %v \n", m)

		return errors.New("something bad happened and I couldn't parse the message")
	})
}

var userHandlerWrapper = func(m dsl.Message) error {
	return userHandler(*m.Content.(*types.User))
}

var orderHandlerWrapper = func(m dsl.Message) error {
	return orderHandler(*m.Content.(*types.Order))
}

var userHandler = func(u types.User) error {
	if u.ID == 0 {
		return errors.New("invalid object supplied, missing fields (id)")
	}

	// ... actually consume the message

	return nil
}

var orderHandler = func(o types.Order) error {
	if o.ID == 0 {
		return errors.New("expected order, missing fields (id)")
	}

	// ... actually consume the message

	return nil
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer: "PactGoMessageConsumer",
		Provider: "PactGoMessageProvider",
		LogDir:   logDir,
		PactDir:  pactDir,
		LogLevel: "DEBUG",
	}
}
