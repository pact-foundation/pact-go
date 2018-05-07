package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

type AccessLevel struct {
	Role string `json:"role,omitempty"`
}

type User struct {
	ID     int           `json:"id,omitempty"`
	Name   string        `json:"name,omitempty"`
	Access []AccessLevel `json:"access,omitempty"`
}

// The actual Provider test itself
func TestMessageProvider_Success(t *testing.T) {
	pact := createPact()

	// Map test descriptions to message producer (handlers)
	// TODO: convert these all to types to ease readability
	functionMappings := dsl.MessageProviders{
		"some test case": func(m dsl.Message) (interface{}, error) {
			fmt.Println("Calling provider function that would produce a message")
			res := User{
				ID:   44,
				Name: "Baz",
				Access: []AccessLevel{
					{Role: "admin"},
					{Role: "admin"},
					{Role: "admin"}},
			}

			return res, nil
		},
	}

	// Verify the Provider with local Pact Files
	pact.VerifyMessageProvider(t, types.VerifyMessageRequest{
		PactURLs: []string{filepath.ToSlash(fmt.Sprintf("%s/pactgomessageconsumer-pactgomessageprovider.json", pactDir))},
	}, functionMappings)
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

// Setup the Pact client.
func createPact() dsl.Pact {
	// Create Pact connecting to local Daemon
	return dsl.Pact{
		Consumer:          "PactGoMessageConsumer",
		Provider:          "PactGoMessageProvider",
		LogDir:            logDir,
		LogLevel:          "DEBUG",
		PactFileWriteMode: "update",
	}
}
