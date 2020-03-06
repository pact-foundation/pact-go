package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyRequestValidate(t *testing.T) {

	t.Run("local validation", func(t *testing.T) {
		tests := []struct {
			name    string
			request VerifyRequest
			err     bool
		}{
			{name: "valid parameters", request: VerifyRequest{
				BrokerURL:                  "http://localhost:1234",
				PactURLs:                   []string{"http://localhost:1234/path/to/pact"},
				BrokerUsername:             "abcd",
				BrokerPassword:             "1234",
				ProviderBaseURL:            "http://localhost:8080",
				ProviderStatesSetupURL:     "http://localhost:8080/setup",
				ProviderVersion:            "1.0.0",
				PublishVerificationResults: true,
				Verbose:                    true,
				CustomProviderHeaders: []string{
					"header: value",
				},
			}, err: false},
			{name: "no base URL provided", request: VerifyRequest{
				PactURLs: []string{"http://localhost:1234/path/to/pact"},
			}, err: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.Validate()
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}

	})

	t.Run("broker integration", func(t *testing.T) {
		tests := []struct {
			name    string
			request VerifyRequest
			err     bool
		}{
			{name: "url without version", request: VerifyRequest{
				PactURLs:        []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL: "http://localhost:8080",
				BrokerURL:       "http://localhost:1234",
			}, err: true},
			{name: "password without username", request: VerifyRequest{
				PactURLs:        []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL: "http://localhost:8080",
				BrokerURL:       "http://localhost:1234",
				ProviderVersion: "1.0.0",
				BrokerPassword:  "1234",
			}, err: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.Validate()
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
			request VerifyRequest
			err     bool
		}{
			{name: "no pacticipant", request: VerifyRequest{
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []ConsumerVersionSelector{ConsumerVersionSelector{}},
			}, err: false},
			{name: "pacticipant only", request: VerifyRequest{
				PactURLs:                 []string{"http://localhost:1234/path/to/pact"},
				ProviderBaseURL:          "http://localhost:8080",
				ConsumerVersionSelectors: []ConsumerVersionSelector{ConsumerVersionSelector{Pacticipant: "foo", Tag: "test"}},
			}, err: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.request.Validate()
				if tt.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}
