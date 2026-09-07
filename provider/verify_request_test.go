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
					assert.Panics(t, func() {
						_ = tt.request.validate(handle)
					})
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
		enablePactUrlFunc := func(t *testing.T) {
			t.Helper()
			t.Setenv("PACT_URL", webhookURL)
		}
		tests := []struct {
			name         string
			setup        func(t *testing.T)
			request      *VerifyRequest
			expectedSize int
			expectedUrls []string
		}{
			{
				name:         "with env var and undefined request.PactURLs",
				setup:        enablePactUrlFunc,
				request:      &VerifyRequest{},
				expectedUrls: []string{webhookURL},
			},
			{
				name:         "with env var and configured PactURLS",
				setup:        enablePactUrlFunc,
				request:      &VerifyRequest{PactURLs: []string{verificationUrl}},
				expectedUrls: []string{verificationUrl, webhookURL},
			},
			{
				name:         "without env var and configured PactURLS",
				request:      &VerifyRequest{PactURLs: []string{verificationUrl}},
				expectedUrls: []string{verificationUrl},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.setup != nil {
					tt.setup(t)
				}
				addPactUrlsFromEnvironment(tt.request)
				assert.ElementsMatch(t, tt.request.PactURLs, tt.expectedUrls)
			})
		}
	})
}

func TestGetPort(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		expected int
	}{
		{name: "explicit port", rawURL: "http://localhost:8080", expected: 8080},
		{name: "explicit port on a non-http scheme", rawURL: "grpc://localhost:1234", expected: 1234},
		{name: "explicit port on an IPv6 host", rawURL: "http://[::1]:8080", expected: 8080},
		{name: "lowest valid port", rawURL: "http://localhost:0", expected: 0},
		{name: "highest valid port", rawURL: "http://localhost:65535", expected: 65535},
		{name: "http default", rawURL: "http://localhost", expected: 80},
		{name: "https default", rawURL: "https://localhost", expected: 443},
		{name: "port above the uint16 range", rawURL: "http://localhost:65536", expected: portOutOfRange},
		{name: "port far above the uint16 range", rawURL: "http://localhost:99999", expected: portOutOfRange},
		{name: "port beyond the int range", rawURL: "http://localhost:99999999999999999999", expected: portOutOfRange},
		{name: "scheme with no default port", rawURL: "tcp://localhost", expected: portUnknownScheme},
		{name: "host and port with no scheme", rawURL: "localhost:8080", expected: portUnknownScheme},
		{name: "unparseable URL", rawURL: "http://localhost:not-a-port", expected: portUnknownScheme},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getPort(tt.rawURL))
		})
	}
}

func TestVerifyRequestValidatePort(t *testing.T) {
	handle := native.NewVerifier("pact-go", command.Version)

	t.Run("port out of range", func(t *testing.T) {
		err := (&VerifyRequest{
			Provider:        "provider",
			ProviderBaseURL: "http://localhost:99999",
			PactFiles:       []string{"/tmp/doesnotexist.json"},
		}).validate(handle)

		assert.ErrorContains(t, err, "out of range")
	})

	t.Run("scheme with no default port", func(t *testing.T) {
		err := (&VerifyRequest{
			Provider:        "provider",
			ProviderBaseURL: "tcp://localhost",
			PactFiles:       []string{"/tmp/doesnotexist.json"},
		}).validate(handle)

		assert.ErrorContains(t, err, "unknown scheme")
	})
}
