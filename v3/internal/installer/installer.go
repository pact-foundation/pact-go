// Package installer is responsible for finding, acquiring and addressing
// runtime dependencies for this package (e.g. Ruby standalone, Rust bindings etc.)
package installer

import (
	"log"
	"os/exec"
)

// Installer manages the underlying Ruby installation
type Installer struct {
	commander commander
}

const (
	mockServiceRange = ">= 0.0.11, < 1.0.0"
	verifierRange    = ">= 0.8.3, < 1.0.0"
)

var versionMap = map[string]string{
	"libpact_mock_server_ffi": mockServiceRange,
	"pact_verifier_cli":       verifierRange,
}

// NewInstaller creates a new initialised Installer
func NewInstaller() *Installer {
	return &Installer{commander: realCommander{}}
}

// CheckInstallation checks installation of all of the tools
func (i *Installer) CheckInstallation() error {

	return nil
}

// CheckVersion checks installation of a given binary using semver-compatible
// comparisions
func (i *Installer) CheckVersion(binary, version string) error {
	log.Println("[DEBUG] checking version for binary", binary, "version", version)

	return nil
}

// GetVersionForBinary gets the version of a given Ruby binary
func (i *Installer) GetVersionForBinary(binary string) (version string, err error) {
	log.Println("[DEBUG] running binary", binary)

	return "", nil
}

// commander wraps the exec package, allowing us
// properly test the file system
type commander interface {
	Output(command string, args ...string) ([]byte, error)
}

type realCommander struct{}

func (c realCommander) Output(command string, args ...string) ([]byte, error) {
	return exec.Command(command, args...).CombinedOutput()
}

// Requirements

// 1. Enable global configuration (environment vars, config files, code options)
// 1. Allow users to specify where they pre-install their artifacts (e.g. /usr/local/pact/libs)
// 1. Download OS specific artifacts if not pre-installed
// 1. Check the semver range of pre-installed artifacts (how to do this with FFI?)
// 1.
