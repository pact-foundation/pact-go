// Package install contains functions necessary for installing and checking
// if the necessary underlying Ruby tools have been properly installed
package install

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	goversion "github.com/hashicorp/go-version"
)

// Installer manages the underlying Ruby installation
type Installer struct {
}

const (
	mockServiceRange = ">= 2.6.4, < 3.0.0"
	verifierRange    = ">= 1.11.0, < 2.0.0"
	brokerRange      = ">= 1.14.0, < 2.0.0"
)

var versionMap = map[string]string{
	"pact-mock-service":      mockServiceRange,
	"pact-provider-verifier": verifierRange,
	"pact-broker":            brokerRange,
}

// CheckInstallation checks installation of all of the tools
func (i *Installer) CheckInstallation() error {

	for binary, versionRange := range versionMap {
		log.Println("[INFO] checking", binary, "within range", versionRange)

		version, err := i.GetVersionForBinary(binary)
		if err != nil {
			return err
		}

		if err = i.CheckVersion(binary, version); err != nil {
			return err
		}
	}

	return nil
}

// CheckVersion checks installation of a given binary using semver-compatible
// comparisions
func (i *Installer) CheckVersion(binary, version string) error {
	log.Println("[DEBUG] checking version for binary", binary, "version", version)
	v, err := goversion.NewVersion(version)
	if err != nil {
		log.Println("[DEBUG] err", err)
		return err
	}

	versionRange, ok := versionMap[binary]
	if !ok {
		return fmt.Errorf("unable to find version range for binary %s", binary)
	}

	log.Println("[DEBUG] checking if version", v, "within semver range", versionRange)
	constraints, err := goversion.NewConstraint(versionRange)
	if constraints.Check(v) {
		log.Println("[DEBUG]", v, "satisfies constraints", v, constraints)
		return nil
	}

	return fmt.Errorf("version %s does not match constraint %s", version, versionRange)
}

// InstallTools installs the CLI tools onto the host system
func (i *Installer) InstallTools() error {
	log.Println("[INFO] Installing tools")

	return nil
}

// GetVersionForBinary gets the version of a given Ruby binary
func (i *Installer) GetVersionForBinary(binary string) (version string, err error) {
	log.Println("[DEBUG] running binary", binary)

	cmd := exec.Command(binary, "version")
	content, err := cmd.Output()
	version = string(content)

	return strings.TrimSpace(version), err
}
