// +build provider

package provider

import (
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
	"os"
	"path/filepath"
	"testing"
)

// An external HTTPS provider
func TestExample_ExternalHttpsProvider(t *testing.T) {
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
