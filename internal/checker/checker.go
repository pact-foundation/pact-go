package checker

import (
	"github.com/pact-foundation/pact-go/v2/installer"
	"github.com/pact-foundation/pact-go/v2/internal/native"
)

func CheckInstall() error {
	// initialised the lib registry
	installer.LibRegistry[installer.FFIPackage] = &native.MockServer{}
	installer.LibRegistry[installer.FFIPackage] = &native.Verifier{}

	i, err := installer.NewInstaller()
	if err != nil {
		return err
	}

	return i.CheckInstallation()
}
