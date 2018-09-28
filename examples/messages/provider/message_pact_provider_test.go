package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/examples/messages/types"
)

var user *types.User

// The actual Provider test itself
func TestMessageProvider_Success(t *testing.T) {
	pact := createPact()

	// Map test descriptions to message producer (handlers)
	functionMappings := dsl.MessageHandlers{
		"a user": func(m dsl.Message) (interface{}, error) {
			if user != nil {
				return user, nil
			} else {
				return map[string]string{
					"message": "not found",
				}, nil
			}
		},
		"an order": func(m dsl.Message) (interface{}, error) {
			return types.Order{
				ID:   1,
				Item: "apple",
			}, nil
		},
	}

	stateMappings := dsl.StateHandlers{
		"user with id 127 exists": func(s dsl.State) error {
			user = &types.User{
				ID:   44,
				Name: "Baz",
				Access: []types.AccessLevel{
					{Role: "admin"},
					{Role: "admin"},
					{Role: "admin"}},
			}

			return nil
		},
		"no users": func(s dsl.State) error {
			user = nil

			return nil
		},
	}

	// Verify the Provider with local Pact Files
	pact.VerifyMessageProvider(t, dsl.VerifyMessageRequest{
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/pactgomessageconsumer-pactgomessageprovider.json", pactDir))},
		MessageHandlers: functionMappings,
		StateHandlers:   stateMappings,
	})
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer: "PactGoMessageConsumer",
		Provider: "PactGoMessageProvider",
		LogDir:   logDir,
		LogLevel: "DEBUG",
	}
}
