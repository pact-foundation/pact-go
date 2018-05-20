package client

import (
	"log"
	"os"
)

// VerificationService is a wrapper for the Pact Provider Verifier Service.
type VerificationService struct {
	ServiceManager
}

// NewService creates a new VerificationService with default settings.
// Arguments allowed:
//
// 		--provider-base-url
// 		--pact-urls
// 		--provider-states-url
// 		--provider-states-setup-url
// 		--broker-username
// 		--broker-password
//    --publish-verification-results
//    --provider-app-version
//    --custom-provider-headers
func (v *VerificationService) NewService(args []string) Service {
	log.Printf("[DEBUG] starting verification service with args: %v\n", args)

	v.Args = args
	v.Cmd = getVerifierCommandPath()
	v.Env = append(os.Environ(), `PACT_INTERACTION_RERUN_COMMAND="To re-run this specific test, set the following environment variables and run your test again: PACT_DESCRIPTION=\"<PACT_DESCRIPTION>\" PACT_PROVIDER_STATE=\"<PACT_PROVIDER_STATE>\""`)

	return v
}

func getVerifierCommandPath() string {
	return "pact-provider-verifier"
}
