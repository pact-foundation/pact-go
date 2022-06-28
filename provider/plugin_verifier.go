package provider

import (
	"testing"
	"time"

	native "github.com/pact-foundation/pact-go/v2/internal/native"
)

// PluginVerifier is used to verify the provider side of a non-HTTP
// transport, provided by a plugin
type PluginVerifier struct {
	// ClientTimeout specifies how long to wait for Pact CLI to start
	// Can be increased to reduce likelihood of intermittent failure
	// Defaults to 10s
	ClientTimeout time.Duration
}

func (v *PluginVerifier) validateConfig() error {
	if v.ClientTimeout == 0 {
		v.ClientTimeout = 10 * time.Second
	}

	return nil
}

// VerifyProviderRaw reads the provided pact files and runs verification against
// a running Provider API, providing raw response from the Verification process.
//
// Order of events: BeforeEach, stateHandlers, requestFilter(pre <execute provider> post), AfterEach
func (v *PluginVerifier) verifyProviderRaw(request VerifyPluginRequest, writer outputWriter) error {
	err := v.validateConfig()

	// TODO: spin up HTTP server to handle states

	if err != nil {
		return err
	}

	native.Init()

	return request.Verify(writer)
}

// VerifyProvider accepts an instance of `*testing.T`
// running the provider verification with granular test reporting and
// automatic failure reporting for nice, simple tests.
func (v *PluginVerifier) VerifyProvider(t *testing.T, request VerifyPluginRequest) error {
	err := v.verifyProviderRaw(request, t)

	// TODO: granular test reporting
	// runTestCases(t, res)

	t.Run("Provider pact verification", func(t *testing.T) {
		if err != nil {
			t.Error(err)
		}
	})

	return err
}
