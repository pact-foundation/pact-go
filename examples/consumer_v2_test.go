//go:build consumer
// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

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
var ArrayMinLike = matchers.ArrayMinLike

type Map = matchers.MapMatcher

func TestConsumerV2(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV2Consumer",
		Provider: "V2Provider",
		Host:     "127.0.0.1",
		TLS:      true,
	})

	assert.NoError(t, err)

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to do a foo").
		WithRequestPathMatcher("POST", Regex("/foobar", `\/foo.*`), func(b *consumer.V2RequestBuilder) {
			b.
				Header("Content-Type", S("application/json")).
				Header("Authorization", Like("Bearer 1234")).
				Query("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
				JSONBody(Map{
					"id":       Like(27),
					"name":     Like("billy"),
					"datetime": Like("2020-01-01'T'08:00:45"),
					"lastName": Like("billy"),
					// "equality": Equality("a thing"), // Add this in and watch me panic
				})
		}).
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.Header("Content-Type", Regex("application/json", "application\\/json"))
			b.JSONBody(Map{
				"datetime": Regex("2020-01-01", "[0-9\\-]+"),
				"name":     S("Billy"),
				"lastName": S("Sampson"),
				"itemsMin": ArrayMinLike("thereshouldbe3ofthese", 3),
				// "equality": Equality("a thing"), // Add this in and watch me panic
			})
		}).
		ExecuteTest(t, test)
	assert.NoError(t, err)
}

func TestConsumerV2_Match(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV2ConsumerMatch",
		Provider: "V2ProviderMatch",
		Host:     "127.0.0.1",
		TLS:      true,
	})

	assert.NoError(t, err)

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to do a foo").
		WithRequest("POST", "/foobar", func(b *consumer.V2RequestBuilder) {
			b.Header("Content-Type", S("application/json"))
			b.Header("Authorization", Like("Bearer 1234"))
			b.Query("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+"))
			b.BodyMatch(&User{})

		}).
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.Header("Content-Type", Regex("application/json", "application\\/json"))
			b.BodyMatch(&User{})
		}).
		ExecuteTest(t, test)
	assert.NoError(t, err)
}

func TestConsumerV2AllInOne(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "PactGoV2ConsumerAllInOne",
		Provider: "V2Provider",
		Host:     "127.0.0.1",
		TLS:      true,
	})

	assert.NoError(t, err)

	// Set up our expected interactions.
	err = mockProvider.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to do a foo").
		WithCompleteRequest(consumer.Request{
			Method: "POST",
			Path:   Regex("/foobar", `\/foo.*`),
			Query: Map{
				"baz": Regex("bat", "[a-zA-Z]+"),
			},
			Headers: commonHeaders,
			Body: Map{
				"id":       Like(27),
				"name":     Like("billy"),
				"datetime": Like("2020-01-01'T'08:00:45"),
				"lastName": Like("billy"),
				// "equality": Equality("a thing"), // Add this in and watch me panic
			},
		}).
		WithCompleteResponse(consumer.Response{
			Headers: commonHeaders,
			Status:  200,
			Body: Map{
				"datetime": Regex("2020-01-01", "[0-9\\-]+"),
				"name":     S("Billy"),
				"lastName": S("Sampson"),
				"itemsMin": ArrayMinLike("thereshouldbe3ofthese", 3),
				// "equality": Equality("a thing"), // Add this in and watch me panic
			},
		}).
		ExecuteTest(t, legacyTest)
	assert.NoError(t, err)
}

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=Billy"`
	LastName string `json:"lastName" pact:"example=Sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
}

// Pass in test case

var test = func() func(config consumer.MockServerConfig) error {
	return rawTest("baz=bat&baz=foo&baz=something")
}()

var rawTest = func(query string) func(config consumer.MockServerConfig) error {

	return func(config consumer.MockServerConfig) error {

		config.TLSConfig.InsecureSkipVerify = true
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: config.TLSConfig,
			},
		}
		req := &http.Request{
			Method: "POST",
			URL: &url.URL{
				Host:     fmt.Sprintf("%s:%d", "localhost", config.Port),
				Scheme:   "https",
				Path:     "/foobar",
				RawQuery: query,
			},
			Body:   ioutil.NopCloser(strings.NewReader(`{"id": 27, "name":"billy", "lastName":"sampson", "datetime":"2021-01-01T08:00:45"}`)),
			Header: make(http.Header),
		}

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 1234")

		_, err := client.Do(req)

		return err
	}
}

var commonHeaders = Map{
	"Content-Type": S("application/json"),
}

var legacyTest = func() func(config consumer.MockServerConfig) error {
	return rawTest("baz=bat")
}()
