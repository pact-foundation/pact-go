package checker

import (
	"github.com/pact-foundation/pact-go/v2/installer"
	"github.com/pact-foundation/pact-go/v2/internal/native/mockserver"
	"github.com/pact-foundation/pact-go/v2/internal/native/verifier"
)

func CheckInstall() error {
	// initialised the lib registry
	installer.LibRegistry[installer.MockServerPackage] = &mockserver.MockServer{}
	installer.LibRegistry[installer.VerifierPackage] = &verifier.Verifier{}

	i, err := installer.NewInstaller()
	if err != nil {
		return err
	}

	return i.CheckInstallation()
}
