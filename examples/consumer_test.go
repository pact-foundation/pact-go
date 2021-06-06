// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	v3 "github.com/pact-foundation/pact-go/v3"
	. "github.com/pact-foundation/pact-go/v3/sugar"
	"github.com/stretchr/testify/assert"
)

func TestConsumerV2(t *testing.T) {
	v3.SetLogLevel("TRACE")

	mockProvider, err := v3.NewV2Pact(v3.MockHTTPProviderConfig{
		Consumer: "V2Consumer",
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
		WithRequest("POST", Regex("/foobar", `\/foo.*`)).
		WithHeader("Content-Type", S("application/json")).
		WithHeader("Authorization", Like("Bearer 1234")).
		WithQuery("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
		WithJSONBody(Map{
			"id":       Like(27),
			"name":     Like("billy"),
			"datetime": Like("2020-01-01'T'08:00:45"),
			"lastName": Like("billy"),
			// "equality": Equality("a thing"), // Add this in and watch me panic
		}).
		WillRespondWith(200).
		WithHeader("Content-Type", Regex("application/json", "application\\/json")).
		WithJSONBody(Map{
			"dateTime": Regex("2020-01-01", "[0-9\\-]+"),
			"name":     S("Billy"),
			"lastName": S("Sampson"),
			"itemsMin": ArrayMinLike("thereshouldbe3ofthese", 3),
			// "equality": Equality("a thing"), // Add this in and watch me panic
		})

	// Execute pact test
	err = mockProvider.ExecuteTest(test)
	assert.NoError(t, err)
}

func TestConsumerV2_Match(t *testing.T) {
	v3.SetLogLevel("TRACE")

	mockProvider, err := v3.NewV2Pact(v3.MockHTTPProviderConfig{
		Consumer: "V2ConsumerMatch",
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
		WithRequest("POST", Regex("/foobar", `\/foo.*`)).
		WithHeader("Content-Type", S("application/json")).
		WithHeader("Authorization", Like("Bearer 1234")).
		WithQuery("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
		WithBodyMatch(&User{}).
		WillRespondWith(200).
		WithHeader("Content-Type", Regex("application/json", "application\\/json")).
		WithBodyMatch(&User{})

	// Execute pact test
	err = mockProvider.ExecuteTest(test)
	assert.NoError(t, err)
}

func TestConsumerV3(t *testing.T) {
	v3.SetLogLevel("TRACE")

	mockProvider, err := v3.NewV3Pact(v3.MockHTTPProviderConfig{
		Consumer: "V3Consumer",
		Provider: "V3Provider",
		Host:     "127.0.0.1",
		TLS:      true,
	})
	assert.NoError(t, err)

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		Given(v3.ProviderStateV3{
			Name: "User foo exists",
			Parameters: map[string]interface{}{
				"id": "foo",
			},
		}).
		UponReceiving("A request to do a foo").
		WithRequest("POST", Regex("/foobar", `\/foo.*`)).
		WithHeader("Content-Type", S("application/json")).
		WithHeader("Authorization", Like("Bearer 1234")).
		WithQuery("baz", Regex("bar", "[a-z]+"), Regex("bat", "[a-z]+"), Regex("baz", "[a-z]+")).
		WithJSONBody(Map{
			"id":       Like(27),
			"name":     FromProviderState("${name}", "billy"),
			"lastName": Like("billy"),
			"datetime": DateTimeGenerated("2020-01-01T08:00:45", "yyyy-MM-dd'T'HH:mm:ss"),
		}).
		WillRespondWith(200).
		WithHeader("Content-Type", S("application/json")).
		WithJSONBody(Map{
			"dateTime":       Regex("2020-01-01", "[0-9\\-]+"),
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

	// Execute pact test
	err = mockProvider.ExecuteTest(test)
	assert.NoError(t, err)
}

func TestMessagePact(t *testing.T) {
	v3.SetLogLevel("TRACE")

	provider, err := NewMessagePactV3(MessageConfig{
		Consumer: "V3MessageConsumer",
		Provider: "V3MessageProvider", // must be different to the HTTP one, can't mix both interaction styles
	})
	assert.NoError(t, err)

	err = provider.AddMessage().
		Given(ProviderStateV3{
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

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=Billy"`
	LastName string `json:"lastName" pact:"example=Sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
}

// Pass in test case
var test = func(config MockServerConfig) error {
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
			RawQuery: "baz=bat&baz=foo&baz=something", // Default behaviour, test matching
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

// Message Pact - wrapped handler extracts the message
var userHandlerWrapper = func(m AsynchronousMessage) error {
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
