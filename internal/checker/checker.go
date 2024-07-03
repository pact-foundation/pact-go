package checker

import (
	"github.com/pact-foundation/pact-go/v2/installer"
	"github.com/pact-foundation/pact-go/v2/internal/native"
)

func CheckInstall(libDir string) error {
	// initialised the lib registry. It just needs one of the main lib interfaces Version() here
	installer.LibRegistry[installer.FFIPackage] = &native.MockServer{}

	i, err := installer.NewInstaller()
	if err != nil {
		return err
	}
	if libDir != "" {
		i.SetLibDir(libDir)
	}
	return i.CheckInstallation()
}
