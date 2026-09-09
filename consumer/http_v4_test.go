package consumer

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpV4TypeSystem(t *testing.T) {
	p, err := NewV4Pact(MockHTTPProviderConfig{
		Consumer: "consumer",
		Provider: "provider",
	})
	require.NoError(t, err)

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
		ExecuteTest(t, func(_ MockServerConfig) error {
			// <- normally run the actually test here.

			return nil
		})
	require.Error(t, err)

	dir, _ := os.Getwd()
	path := strings.ReplaceAll(dir, "\\", "/") + "/pact_plugin.proto"

	err = p.AddInteraction().
		Given("some state").
		UponReceiving("some scenario").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.5.4",
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
		ExecuteTest(t, func(_ MockServerConfig) error {
			// <- normally run the actually test here.

			return nil
		})
	assert.Error(t, err)
}

func TestV4HTTPAddExternalReference(t *testing.T) {
	p, err := NewV4Pact(MockHTTPProviderConfig{
		Consumer: "consumer",
		Provider: "provider",
	})
	require.NoError(t, err)

	err = p.AddInteraction().
		UponReceiving("a request with an external reference").
		AddExternalReference("Jira", "TICKET-123", "https://jira.example.com/browse/TICKET-123").
		WithRequest("GET", "/", func(_ *V4RequestBuilder) {}).
		WillRespondWith(200, func(_ *V4ResponseBuilder) {}).
		ExecuteTest(t, func(msc MockServerConfig) error {
			url := "http://" + net.JoinHostPort(msc.Host, strconv.Itoa(msc.Port)) + "/"
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("building the mock server request: %w", err)
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("calling the mock server: %w", err)
			}
			defer func() { _ = res.Body.Close() }()
			return nil
		})
	assert.NoError(t, err)
}

var (
	Like        = matchers.Like
	EachLike    = matchers.EachLike
	Term        = matchers.Term
	Regex       = matchers.Regex
	HexValue    = matchers.HexValue
	Identifier  = matchers.Identifier
	IPAddress   = matchers.IPAddress
	IPv6Address = matchers.IPv6Address
	Timestamp   = matchers.Timestamp
	Date        = matchers.Date
	Time        = matchers.Time
	UUID        = matchers.UUID
)

type S = matchers.String

var ArrayMinLike = matchers.ArrayMinLike

type Map = matchers.Map
