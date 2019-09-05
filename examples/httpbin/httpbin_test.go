package provider

import (
	"fmt"
	"path/filepath"
	"testing"
	"os"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/utils"
	"github.com/pact-foundation/pact-go/types"
)

// An external HTTPS provider
func TestPact_ExternalHttpsProvider(t *testing.T) {
	pact := createPact()

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
		Consumer:                 "consumer",
		Provider:                 "httpbinprovider",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
		LogLevel:                 "DEBUG",
	}
}