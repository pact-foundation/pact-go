package daemon

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kardianos/osext"
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
func (m *VerificationService) NewService(args []string) (int, Service) {
	log.Printf("[DEBUG] starting verification service with args: %v\n", args)

	m.Args = args
	m.Command = getVerifierCommandPath()
	return -1, m
}

func getVerifierCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact-provider-verifier", "bin", "pact-provider-verifier"))
}
