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
		//nolint:wrapcheck // version.CheckVersion log.Fatals this error verbatim, so the
		// installer's own diagnostic ("darwin is not a supported OS", ...) is what users
		// read; a wrapper prefix here would only bury it.
		return err
	}

	//nolint:wrapcheck // as above: CheckInstallation's messages are the installer's
	// user-facing diagnostics, surfaced verbatim by version.CheckVersion.
	return i.CheckInstallation()
}
