//go:build consumer
// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"
)

func TestConsumerV4(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV4Consumer",
		Provider: "V4Provider",
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
		WithRequest("POST", "/foobar", func(b *consumer.V4RequestBuilder) {
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
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
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
