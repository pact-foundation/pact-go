package provider

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
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

type AccessLevel struct {
	Role string `json:"role,omitempty"`
}

type User struct {
	ID     int           `json:"id,omitempty"`
	Name   string        `json:"name,omitempty"`
	Access []AccessLevel `json:"access,omitempty"`
}

var userHandlerWrapper = func(m dsl.Message) error {
	return userHandler(*m.Content.(*User))
}

var userHandler = func(u User) error {
	if u.ID == -1 {
		return errors.New("invalid object supplied, missing fields (id)")
	}

	// ... actually consume the message

	return nil
}

func TestMessageConsumer_Success(t *testing.T) {
	message := &dsl.Message{}
	message.
		Given("some state").
		ExpectsToReceive("some test case").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"id":   like(127),
			"name": "Baz",
			"access": eachLike(map[string]interface{}{
				"role": term("admin", "admin|controller|user"),
			}, 3),
		}).
		AsType(&User{})

	pact.VerifyMessageConsumer(t, message, userHandlerWrapper)
}
func TestMessageConsumer_Fail(t *testing.T) {
	t.Skip()
	message := &dsl.Message{}
	message.
		Given("some state").
		ExpectsToReceive("some test case").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"foo": "bar",
		})

	pact.VerifyMessageConsumer(t, message, func(m dsl.Message) error {
		t.Logf("[DEBUG] calling message handler func with arguments: %v \n", m)

		return errors.New("something bad happened and I couldn't parse the message")
	})
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer:          "PactGoMessageConsumer",
		Provider:          "PactGoMessageProvider",
		LogDir:            logDir,
		PactDir:           pactDir,
		LogLevel:          "DEBUG",
		PactFileWriteMode: "update",
	}
}
