package types

import (
	"testing"
)

func TestVerifyRequest_ValidRequest(t *testing.T) {
	r := VerifyRequest{
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
	}

	err := r.Validate()

	if err != nil {
		t.Fatal("want nil, got err: ", err)
	}
}

func TestVerifyRequest_NoBaseURL(t *testing.T) {
	r := VerifyRequest{
		PactURLs: []string{"http://localhost:1234/path/to/pact"},
	}

	err := r.Validate()

	if err == nil {
		t.Fatal("want err, got nil")
	}
}

func TestVerifyRequest_BrokerUsernameWithoutPassword(t *testing.T) {
	r := VerifyRequest{
		PactURLs:        []string{"http://localhost:1234/path/to/pact"},
		ProviderBaseURL: "http://localhost:8080",
		BrokerURL:       "http://localhost:1234",
		ProviderVersion: "1.0.0.",
		BrokerPassword:  "1234",
	}

	err := r.Validate()

	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestVerifyRequest_BrokerURLWithoutVersion(t *testing.T) {
	r := VerifyRequest{
		PactURLs:        []string{"http://localhost:1234/path/to/pact"},
		ProviderBaseURL: "http://localhost:8080",
		BrokerURL:       "http://localhost:1234",
		BrokerPassword:  "1234",
	}

	err := r.Validate()

	if err == nil {
		t.Fatal("want error, got nil")
	}
}
