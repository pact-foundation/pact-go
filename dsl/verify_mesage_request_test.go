package dsl

import "testing"

func TestVerifyMessageRequest_Valid(t *testing.T) {
	r := VerifyMessageRequest{
		PactURLs: []string{
			"http://localhost:1234",
		},
		BrokerPassword:             "user",
		BrokerUsername:             "pass",
		ProviderVersion:            "1.0.0",
		PublishVerificationResults: true,
	}

	err := r.Validate()

	if err != nil {
		t.Fatal("want nil, got err: ", err)
	}
}
func TestVerifyMessageRequest_Invalid(t *testing.T) {
	r := VerifyMessageRequest{}

	err := r.Validate()

	if err == nil {
		t.Fatal("want error, got nil")
	}
}
