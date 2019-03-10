package types

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishRequest_Validate(t *testing.T) {
	dir, _ := os.Getwd()
	testFile := fmt.Sprintf(filepath.Join(dir, "publish_test.go"))

	p := PublishRequest{}
	err := p.Validate()
	if err.Error() != "'PactURLs' is mandatory" {
		t.Fatalf("Expected a different error but got '%s'", err.Error())
	}

	p = PublishRequest{
		PactURLs: []string{testFile},
	}

	err = p.Validate()
	if err.Error() != "'PactBroker' is mandatory" {
		t.Fatalf("Expected a different error but got '%s'", err.Error())
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs:   []string{testFile},
	}

	err = p.Validate()
	if err.Error() != "'ConsumerVersion' is mandatory" {
		t.Fatalf("Expected a different error but got '%s'", err.Error())
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs: []string{
			testFile,
		},
		ConsumerVersion: "1.0.0",
		BrokerUsername:  "userwithoutpass",
	}

	err = p.Validate()
	if err.Error() != "both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given" {
		t.Fatalf("Expected a different error but got '%s'", err.Error())
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs: []string{
			testFile,
		},
		ConsumerVersion: "1.0.0",
		BrokerPassword:  "passwithoutuser",
	}

	err = p.Validate()
	if err.Error() != "both 'BrokerUsername' and 'BrokerPassword' must be supplied if one given" {
		t.Fatalf("Expected a different error but got '%s'", err.Error())
	}

	p = PublishRequest{
		PactURLs: []string{
			"aoeuaoeu",
		},
	}

	err = p.Validate()
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs: []string{
			testFile,
		},
		ConsumerVersion: "1.0.0",
	}

	err = p.Validate()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs: []string{
			testFile,
		},
		ConsumerVersion: "1.0.0",
	}

	err = p.Validate()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	p = PublishRequest{
		PactBroker: "http://foo.com",
		PactURLs: []string{
			testFile,
		},
		ConsumerVersion: "1.0.0",
		BrokerToken:     "1234",
	}

	err = p.Validate()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}
