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
		})

	err := pact.VerifyMessageConsumer(message, func(m dsl.Message) error {
		t.Logf("[DEBUG] calling message handler func with arguments: %v \n", m.Content)

		body := m.Content.(map[string]interface{})

		_, ok := body["id"]

		if !ok {
			return errors.New("invalid object supplied, missing fields (id)")
		}

		return nil
	})

	if err != nil {
		t.Fatal("VerifyMessageConsumer failed:", err)
	}
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

	err := pact.VerifyMessageConsumer(message, func(m dsl.Message) error {
		t.Logf("[DEBUG] calling message handler func with arguments: %v \n", m)

		return errors.New("something bad happened and I couldn't parse the message")
	})

	if err != nil {
		t.Fatal("VerifyMessageConsumer failed:", err)
	}
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

// Setup the Pact client.
func createPact() dsl.Pact {
	// Create Pact connecting to local Daemon
	return dsl.Pact{
		Consumer: "billy",
		Provider: "bobby",
		LogDir:   logDir,
		// PactDir:  pactDir, // TODO: this seems to cause an issue "NoMethodError: undefined method `content' for #<Pact::Interaction:0x00007fc8f1a082e8>"
		PactDir:           "/tmp",
		LogLevel:          "DEBUG",
		PactFileWriteMode: "update",
	}
}
