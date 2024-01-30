package consumer

import (
	"fmt"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

func TestHttpV4TypeSystem(t *testing.T) {

	p, err := NewV4Pact(MockHTTPProviderConfig{
		Consumer: "consumer",
		Provider: "provider",
	})
	assert.NoError(t, err)

	err = p.AddInteraction().
		Given("some state").
		UponReceiving("some scenario").
		WithRequest("GET", "/", func(b *V4RequestBuilder) {
			b.
				Header("Content-Type", S("application/json")).
				Header("Authorization", Like("Bearer 1234")).
				Query("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
				JSONBody(Map{
					"id":       Like(27),
					"name":     Like("billy"),
					"datetime": Like("2020-01-01'T'08:00:45"),
					"lastName": Like("billy"),
				})
		}).
		WillRespondWith(200, func(b *V4ResponseBuilder) {
			b.
				Header("Content-Type", Regex("application/json", "application\\/json")).
				JSONBody(Map{
					"datetime": Regex("2020-01-01", "[0-9\\-]+"),
					"name":     S("Billy"),
					"lastName": S("Sampson"),
					"itemsMin": ArrayMinLike("thereshouldbe3ofthese", 3),
				})

		}).
		ExecuteTest(t, func(msc MockServerConfig) error {
			// <- normally run the actually test here.

			return nil
		})
	assert.Error(t, err)

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/pact_plugin.proto", dir)

	err = p.AddInteraction().
		Given("some state").
		UponReceiving("some scenario").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.13",
		}).
		WithRequest("GET", "/").
		// WithRequest("GET", "/", func(b *V4InteractionWithPluginRequestBuilder) {
		// 	b.PluginContents("application/protobufs", "")
		// 	// TODO:
		// }).
		WillRespondWith(200, func(b *V4InteractionWithPluginResponseBuilder) {
			b.
				Header("Content-Type", S("application/protobufs")).
				PluginContents("application/protobufs", `
					{
						"pact:proto": "`+path+`",
						"pact:message-type": "InitPluginRequest"
					}
				`)
		}).
		ExecuteTest(t, func(msc MockServerConfig) error {
			// <- normally run the actually test here.

			return nil
		})
	assert.Error(t, err)

}

var Like = matchers.Like
var EachLike = matchers.EachLike
var Term = matchers.Term
var Regex = matchers.Regex
var HexValue = matchers.HexValue
var Identifier = matchers.Identifier
var IPAddress = matchers.IPAddress
var IPv6Address = matchers.IPv6Address
var Timestamp = matchers.Timestamp
var Date = matchers.Date
var Time = matchers.Time
var UUID = matchers.UUID

type S = matchers.String

var ArrayMinLike = matchers.ArrayMinLike

type Map = matchers.Map
