package provider

import (
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
		}{
			{name: "valid parameters", request: &VerifyRequest{
				handle:                 handle,
				PactURLs:               []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:        "http://localhost:8080",
				ProviderStatesSetupURL: "http://localhost:8080/setup",
				ProviderVersion:        "1.0.0",
			}, err: false},
			{name: "no base URL provided", request: &VerifyRequest{
				handle:   handle,
				PactURLs: []string{"http://localhost:1234/path/to/pact"},
			}, err: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.validate()
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
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
				handle:          handle,
				PactURLs:        []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL: "http://localhost:8080",
				BrokerURL:       "http://localhost:1234",
			}, err: true},
			{name: "broker url without name/version", request: &VerifyRequest{
				handle:          handle,
				BrokerURL:       "http://localhost:1234",
				ProviderVersion: "1.0.0",
				BrokerPassword:  "1234",
			}, err: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.validate()
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
				handle:                   handle,
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []Selector{&ConsumerVersionSelector{}},
			}, err: false},
			{name: "pacticipant only", request: &VerifyRequest{
				handle:                   handle,
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []Selector{&ConsumerVersionSelector{Consumer: "foo", Tag: "test"}},
			}, err: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.validate()
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
