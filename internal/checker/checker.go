// Package checker verifies that the native pact_ffi library required by
// this module is installed and loadable at the expected version.
package checker

import (
	"github.com/pact-foundation/pact-go/v2/installer"
	"github.com/pact-foundation/pact-go/v2/internal/native"
)

// CheckInstall verifies that the native pact_ffi library is installed and
// at a compatible version, loading it if necessary.
func CheckInstall() error {
	// initialised the lib registry. It just needs one of the main lib interfaces Version() here
	installer.LibRegistry[installer.FFIPackage] = &native.MockServer{}

	i, err := installer.NewInstaller()
	if err != nil {
		return err
	}

	return i.CheckInstallation()
}
