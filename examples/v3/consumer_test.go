// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
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

	// Create Pact connecting to local Daemon
	mockProvider, err := v3.NewHTTPMockProviderV2(v3.MockHTTPProviderConfigV2{
		Consumer:             "MyConsumer",
		Provider:             "MyProvider",
		Host:                 "127.0.0.1",
		Port:                 8080,
		SpecificationVersion: v3.V2,
		TLS:                  true,
	})

	// TODO: probably better than deferring to the execute test phase, but not sure
	if err != nil {
		t.Fatal(err)
	}

	// Set up our expected interactions.
	mockProvider.
		AddInteraction().
		// Given("User foo exists").
		UponReceiving("A request to do a foo").
		WithRequest(v3.Request{
			Method:  "POST",
			Path:    s("/foobar"),
			Headers: v3.MapMatcher{"Content-Type": s("application/json"), "Authorization": s("Bearer 1234")},
			Body: v3.MapMatcher{
				"name": s("billy"),
			},
		}).
		WillRespondWith(v3.Response{
			Status:  200,
			Headers: v3.MapMatcher{"Content-Type": s("application/json")},
			// Body:    v3.Match(&User{}),
			Body: v3.MapMatcher{
				"dateTime": v3.Regex("2020-01-01", "[0-9\\-]+"),
				"name":     s("FirstName"),
				"lastName": s("LastName"),
			},
		})

	// Execute pact test
	if err := mockProvider.ExecuteTest(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

func TestConsumerV3(t *testing.T) {
	v3.SetLogLevel("TRACE")

	// Create Pact connecting to local Daemon
	mockProvider, err := v3.NewHTTPMockProviderV3(v3.MockHTTPProviderConfigV2{
		Consumer:             "MyConsumer",
		Provider:             "MyProvider",
		Host:                 "127.0.0.1",
		Port:                 8080,
		SpecificationVersion: v3.V2,
		TLS:                  true,
	})

	if err != nil {
		t.Fatal(err)
	}

	// Execute pact test
	if err := mockProvider.ExecuteTest(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}

type User struct {
	Name     string `json:"name" pact:"example=billy"`
	LastName string `json:"lastName" pact:"example=sampson"`
	Date     string `json:"datetime" pact:"example=20200101,regex=[0-9a-z-A-Z]+"`
}

// Pass in test case
var test = func(config v3.MockServerConfig) error {
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
		},
		Body:   ioutil.NopCloser(strings.NewReader(`{"name":"billy"}`)),
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
