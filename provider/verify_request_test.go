package provider

import (
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/command"
	"github.com/pact-foundation/pact-go/v2/internal/native"
	"github.com/stretchr/testify/assert"
)

func TestVerifyRequestValidate(t *testing.T) {
	handle := native.NewVerifier("pact-go", command.Version)

	t.Run("local validation", func(t *testing.T) {
		tests := []struct {
			name    string
			request *VerifyRequest
			err     bool
			panic   bool
		}{
			{name: "valid parameters", request: &VerifyRequest{
				PactURLs:               []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:        "http://localhost:8080",
				ProviderStatesSetupURL: "http://localhost:8080/setup",
				ProviderVersion:        "1.0.0",
			}, err: false},
			{name: "no base URL provided", request: &VerifyRequest{
				PactURLs: []string{"http://localhost:1234/path/to/pact"},
			}, err: true, panic: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.panic {
					assert.Panics(t, (func() {
						tt.request.validate(handle)
					}))
				} else {
					err := tt.request.validate(handle)
					if tt.err {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				}
			})
		}

	})

	t.Run("broker integration", func(t *testing.T) {
		handle := native.NewVerifier("pact-go", command.Version)

		tests := []struct {
			name    string
			request *VerifyRequest
			err     bool
		}{
			{name: "url without version", request: &VerifyRequest{
				PactURLs:        []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL: "http://localhost:8080",
				BrokerURL:       "http://localhost:1234",
			}, err: true},
			{name: "broker url without name/version", request: &VerifyRequest{
				BrokerURL:       "http://localhost:1234",
				ProviderBaseURL: "http://localhost:8080",
				ProviderVersion: "1.0.0",
				BrokerPassword:  "1234",
			}, err: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.validate(handle)
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("consumer version selectors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *VerifyRequest
			err     bool
		}{
			{name: "no pacticipant", request: &VerifyRequest{
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []Selector{&ConsumerVersionSelector{}},
			}, err: false},
			{name: "pacticipant only", request: &VerifyRequest{
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []Selector{&ConsumerVersionSelector{Consumer: "foo", Tag: "test"}},
			}, err: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.validate(handle)
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

func TestVerifyRequest(t *testing.T) {
	t.Run("#addPactUrlsFromEnvironment", func(t *testing.T) {
		const webhookURL, verificationUrl = "pact_changed_webhook_url", "http://localhost:1234/path/to/pact"
		enablePactUrlFunc := func() func() {
			const pactUrl = "PACT_URL"
			os.Setenv(pactUrl, webhookURL)
			return func() {
				defer os.Unsetenv(pactUrl)
			}
		}
		tests := []struct {
			name         string
			setup        func() (teardown func())
			request      *VerifyRequest
			expectedSize int
			expectedUrls []string
		}{
			{name: "with env var and undefined request.PactURLs",
				setup:        enablePactUrlFunc,
				request:      &VerifyRequest{},
				expectedUrls: []string{webhookURL},
			},
			{name: "with env var and configured PactURLS",
				setup:        enablePactUrlFunc,
				request:      &VerifyRequest{PactURLs: []string{verificationUrl}},
				expectedUrls: []string{verificationUrl, webhookURL},
			},
			{name: "without env var and configured PactURLS",
				setup:        func() func() { return func() {} },
				request:      &VerifyRequest{PactURLs: []string{verificationUrl}},
				expectedUrls: []string{verificationUrl},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				teardown := tt.setup()
				defer teardown()
				addPactUrlsFromEnvironment(tt.request)
				assert.ElementsMatch(t, tt.request.PactURLs, tt.expectedUrls)
			})
		}
	})
}
