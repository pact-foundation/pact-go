// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	v3 "github.com/pact-foundation/pact-go/v3"
)

type s = v3.String

func TestConsumerV2(t *testing.T) {
	v3.SetLogLevel("TRACE")

	mockProvider, err := v3.NewHTTPMockProviderV2(v3.MockHTTPProviderConfigV2{
		Consumer: "V2Consumer",
		Provider: "V2Provider",
		Host:     "127.0.0.1",
		Port:     8080,
		TLS:      true,
	})

	// Override default matching behaviour
	// mockProvider.SetMatchingConfig(v3.PactSerialisationOptionsV2{
	// QueryStringStyle: v3.AlwaysArray,
	// QueryStringStyle: v3.Array,
	// QueryStringStyle: v3.Default,
	// })

	// TODO: probably better than deferring to the execute test phase, but not sure
	if err != nil {
		t.Fatal(err)
	}

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to do a foo").
		WithRequest(v3.Request{
			Method:  "POST",
			Path:    v3.Regex("/foobar", `\/foo.*`),
			Headers: v3.MapMatcher{"Content-Type": s("application/json"), "Authorization": v3.Like("Bearer 1234")},
			Query: v3.QueryMatcher{
				"baz": []v3.Matcher{
					v3.Regex("bar", "[a-z]+"),
					v3.Regex("bat", "[a-z]+"),
					v3.Regex("baz", "[a-z]+"),
				},
			},
			// Body: v3.MapMatcher{
			// 	"id":       v3.Like(27),
			// 	"name":     v3.Like("billy"),
			// 	"datetime": v3.Like("2020-01-01'T'08:00:45"),
			// 	"lastName": v3.Like("billy"),
			// },
			Body: v3.MatchV2(&User{}),
		}).
		WillRespondWith(v3.Response{
			Status:  200,
			Headers: v3.MapMatcher{"Content-Type": v3.Regex("application/json", "application\\/json")},
			// Body:    v3.Match(&User{}),
			Body: v3.MapMatcher{
				"dateTime": v3.Regex("2020-01-01", "[0-9\\-]+"),
				"name":     s("FirstName"),
				"lastName": s("LastName"),
				"itemsMin": v3.ArrayMinLike("thereshouldbe3ofthese", 3),
				// Add any of these this to demonstrate adding a v3 matcher failing the build (not at the type system level unfortunately)
				// "id": v3.Integer(1),
				// "superstring": v3.Includes("foo"),
				// "accountBalance": v3.Decimal(123.76),
				// "itemsMinMax": v3.ArrayMinMaxLike(27, 3, 5),
				// "equality": v3.Equality("a thing"),
			},
		})

	// Execute pact test
	if err := mockProvider.ExecuteTest(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

func TestConsumerV3(t *testing.T) {
	v3.SetLogLevel("TRACE")

	mockProvider, err := v3.NewHTTPMockProviderV3(v3.MockHTTPProviderConfigV2{
		Consumer: "V3Consumer",
		Provider: "V3Provider",
		Host:     "127.0.0.1",
		Port:     8080,
		TLS:      true,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		// TODO: map this to given_with_param interface!
		Given(v3.ProviderStateV3{
			Name: "User foo exists",
			Parameters: map[string]interface{}{
				"id": "foo",
			},
		}).
		UponReceiving("A request to do a foo").
		WithRequest(v3.Request{
			Method:  "POST",
			Path:    v3.Regex("/foobar", `\/foo.*`),
			Headers: v3.MapMatcher{"Content-Type": s("application/json"), "Authorization": v3.Like("Bearer 1234")},
			Body: v3.MapMatcher{
				"id":       v3.Like(27),
				"name":     v3.FromProviderState("${name}", "billy"),
				"lastName": v3.Like("billy"),
				"datetime": v3.DateTimeGenerated("2020-01-01T08:00:45", "yyyy-MM-dd'T'HH:mm:ss"),
			},

			// Alternative use MatchV3
			// Body: v3.MatchV3(&User{}),
			// Body: v3.MatchV2(&User{}),
			Query: v3.QueryMatcher{
				"baz": []v3.Matcher{
					v3.Regex("bar", "[a-z]+"),
					v3.Regex("bat", "[a-z]+"),
					v3.Regex("baz", "[a-z]+"),
				},
			},
		}).
		WillRespondWith(v3.Response{
			Status:  200,
			Headers: v3.MapMatcher{"Content-Type": s("application/json")},
			// Body:    v3.MatchV3(&User{}),
			Body: v3.MapMatcher{
				"dateTime":       v3.Regex("2020-01-01", "[0-9\\-]+"),
				"name":           s("FirstName"),
				"lastName":       s("LastName"),
				"superstring":    v3.Includes("foo"),
				"id":             v3.Integer(12),
				"accountBalance": v3.Decimal(123.76),
				"itemsMinMax":    v3.ArrayMinMaxLike(27, 3, 5),
				"itemsMin":       v3.ArrayMinLike("thereshouldbe3ofthese", 3),
				"equality":       v3.Equality("a thing"),
				"arrayContaining": v3.ArrayContaining([]interface{}{
					v3.Like("string"),
					v3.Integer(1),
					v3.MapMatcher{
						"foo": v3.Like("bar"),
					},
				}),
			},
		})

	// Execute pact test
	if err := mockProvider.ExecuteTest(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

func TestMessagePact(t *testing.T) {
	provider, err := v3.NewMessagePactV3(v3.MessageConfig{
		Consumer:             "V3MessageConsumer",
		Provider:             "V3MessageProvider", // must be different to the HTTP one, can't mix both interaction styles
		SpecificationVersion: v3.V3,
	})

	if err != nil {
		t.Fatal(err)
	}

	message := provider.AddMessage()
	message.
		Given(v3.ProviderStateV3{
			Name: "User with id 127 exists",
			Parameters: map[string]interface{}{
				"id": 127,
			},
		}).
		ExpectsToReceive("a user event").
		WithMetadata(v3.MapMatcher{
			"Content-Type": s("application/json; charset=utf-8"),
		}).
		WithContent(v3.MapMatcher{
			"datetime": v3.Regex("2020-01-01", "[0-9\\-]+"),
			"name":     s("FirstName"),
			"lastName": s("LastName"),
			"id":       v3.Integer(12),
		}).
		AsType(&User{})

	provider.VerifyMessageConsumer(t, message, userHandlerWrapper)
}

type User struct {
	ID       int    `json:"id" pact:"example=27"`
	Name     string `json:"name" pact:"example=billy"`
	LastName string `json:"lastName" pact:"example=sampson"`
	Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
	// Date     string `json:"datetime" pact:"example=2020-01-01'T'08:00:45,regex=[0-9-]+,format=yyyy-MM-dd'T'HH:mm:ss,generator=datetime"`
	// Date     string `json:"datetime" pact:"example=20200101,regex=[0-9a-z-A-Z]+"`
}

// Pass in test case
var test = func(config v3.MockServerConfig) error {
	config.TLSConfig.InsecureSkipVerify = true
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config.TLSConfig,
		},
	}
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Host:   fmt.Sprintf("%s:%d", "localhost", config.Port),
			Scheme: "https",
			Path:   "/foobar",
			// RawQuery: "baz=foo&baz=foo&baz=foo", // TODO: Currently doesn't support matching rules being sent over the wire, so must have exact values
			RawQuery: "baz=bat&baz=foo&baz=something", // Default behaviour, test matching
			// RawQuery: "baz[]=bat&baz[]=foo&baz[]=something", // TODO: Rust v3 does not support this syntax
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
var userHandlerWrapper = func(m v3.Message) error {
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
