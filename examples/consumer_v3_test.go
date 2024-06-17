//go:build consumer
// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"errors"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	message "github.com/pact-foundation/pact-go/v2/message/v3"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"
)

var Decimal = matchers.Decimal
var Integer = matchers.Integer
var Equality = matchers.Equality
var Includes = matchers.Includes
var FromProviderState = matchers.FromProviderState
var EachKeyLike = matchers.EachKeyLike
var ArrayContaining = matchers.ArrayContaining
var ArrayMinMaxLike = matchers.ArrayMinMaxLike
var ArrayMaxLike = matchers.ArrayMaxLike
var DateGenerated = matchers.DateGenerated
var TimeGenerated = matchers.TimeGenerated
var DateTimeGenerated = matchers.DateTimeGenerated

func TestConsumerV3(t *testing.T) {
	log.SetLogLevel("INFO")

	mockProvider, err := consumer.NewV3Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV3Consumer",
		Provider: "V3Provider",
		Host:     "127.0.0.1",
		TLS:      true,
	})
	assert.NoError(t, err)

	// Set up our expected interactions.
	err = mockProvider.
		AddInteraction().
		Given("state 1").
		GivenWithParameter(models.ProviderState{
			Name: "User foo exists",
			Parameters: map[string]interface{}{
				"id": "foo",
			},
		}).
		UponReceiving("A request to do a foo").
		WithRequest("POST", "/foobar", func(b *consumer.V3RequestBuilder) {
			b.
				Header("Content-Type", S("application/json")).
				Header("Authorization", Like("Bearer 1234")).
				Query("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
				JSONBody(Map{
					"id":       Like(27),
					"name":     FromProviderState("${name}", "billy"),
					"lastName": Like("billy"),
					"datetime": DateTimeGenerated("2020-01-01T08:00:45", "yyyy-MM-dd'T'HH:mm:ss"),
				})

		}).
		WillRespondWith(200, func(b *consumer.V3ResponseBuilder) {
			b.
				Header("Content-Type", S("application/json")).
				JSONBody(Map{
					"datetime":       Regex("2020-01-01", "[0-9\\-]+"),
					"name":           S("Billy"),
					"lastName":       S("Sampson"),
					"superstring":    Includes("foo"),
					"id":             Integer(12),
					"accountBalance": Decimal(123.76),
					"itemsMinMax":    ArrayMinMaxLike(27, 3, 5),
					"itemsMin":       ArrayMinLike("thereshouldbe3ofthese", 3),
					"equality":       Equality("a thing"),
					"arrayContaining": ArrayContaining([]interface{}{
						Like("string"),
						Integer(1),
						Map{
							"foo": Like("bar"),
						},
					}),
				})
		}).
		ExecuteTest(t, test)
	assert.NoError(t, err)
}
func TestMessagePact(t *testing.T) {
	log.SetLogLevel("INFO")

	provider, err := message.NewMessagePact(message.Config{
		Consumer: "PactGoV3MessageConsumer",
		Provider: "V3MessageProvider", // must be different to the HTTP one, can't mix both interaction styles
	})
	assert.NoError(t, err)

	err = provider.AddMessage().
		GivenWithParameter(models.ProviderState{
			Name: "User with id 127 exists",
			Parameters: map[string]interface{}{
				"id": 127,
			},
		}).
		ExpectsToReceive("a user event").
		WithMetadata(map[string]string{
			"Content-Type": "application/json",
		}).
		WithJSONContent(Map{
			"datetime": Regex("2020-01-01", "[0-9\\-]+"),
			"name":     S("Billy"),
			"lastName": S("Sampson"),
			"id":       Integer(12),
		}).
		AsType(&User{}).
		ConsumedBy(userHandlerWrapper).
		Verify(t)

	assert.NoError(t, err)
}

// Message Pact - wrapped handler extracts the message
var userHandlerWrapper = func(m message.MessageContents) error {
	return userHandler(*m.Content.(*User))
}

// Message Pact - actual handler
var userHandler = func(u User) error {
	if u.ID == 0 {
		return errors.New("invalid object supplied, missing fields (id)")
	}

	// ... actually consume the message

	return nil
}
