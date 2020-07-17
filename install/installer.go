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
	commander commander
}

const (
	mockServiceRange = ">= 3.5.0, < 4.0.0"
	verifierRange    = ">= 1.31.0, < 2.0.0"
	brokerRange      = ">= 1.22.3"
)

var versionMap = map[string]string{
	"pact-mock-service":      mockServiceRange,
	"pact-provider-verifier": verifierRange,
	"pact-broker":            brokerRange,
}

// NewInstaller creates a new initialised Installer
func NewInstaller() *Installer {
	return &Installer{commander: realCommander{}}
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
// comparisons
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

	return fmt.Errorf("version %s of %s does not match constraint %s", version, binary, versionRange)
}

// GetVersionForBinary gets the version of a given Ruby binary
func (i *Installer) GetVersionForBinary(binary string) (version string, err error) {
	log.Println("[DEBUG] running binary", binary)

	content, err := i.commander.Output(binary, "version")
	elements := strings.Split(strings.TrimSpace(string(content)), "\n")
	version = strings.TrimSpace(elements[len(elements)-1])

	return version, err
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
