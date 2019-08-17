package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

// An external HTTPS provider
func TestPact_GinProvider(t *testing.T) {

	pact := createPact()

	// Verify the Provider with local Pact Files
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:       "https://httpbin.org",
		PactURLs:              []string{filepath.ToSlash(fmt.Sprintf("%s/consumer-httpbin.json", pactDir))},
		CustomProviderHeaders: []string{"Authorization: Bearer SOME_TOKEN"},
	})

	if err != nil {
		t.Fatal(err)
	}
}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer:                 "jmarie",
		Provider:                 "loginprovider",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
		LogLevel:                 "DEBUG",
	}
}
