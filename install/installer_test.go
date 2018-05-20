package install

import (
	"errors"
	"reflect"
	"testing"
)

type testCommander struct {
	// Version to return
	version string

	// Error to return
	err error
}

// set version range to be the same for all binaries
func initVersionRange() {
	versionMap = map[string]string{
		"pact-mock-service":      ">= 1.0.0, < 2.0.0",
		"pact-provider-verifier": ">= 1.0.0, < 2.0.0",
		"pact-broker":            ">= 1.0.0, < 2.0.0",
	}
}

func (c testCommander) Output(command string, args ...string) ([]byte, error) {
	return []byte(c.version), c.err
}

func getInstaller(version string, err error) *Installer {
	initVersionRange()
	return &Installer{testCommander{version, err}}
}

func TestInstaller_NewInstaller(t *testing.T) {
	i := NewInstaller()

	if reflect.TypeOf(i).String() != "*install.Installer" {
		t.Fatal("want *install.Installer, got", reflect.TypeOf(i).String())
	}
}
func TestInstaller_CheckVersion(t *testing.T) {
	i := getInstaller("1.5.0", nil)
	err := i.CheckVersion("pact-mock-service", "1.5.0")

	if err != nil {
		t.Fatal("error:", err)
	}
}

func TestInstaller_CheckVersionFail(t *testing.T) {
	i := getInstaller("2.0.0", nil)
	err := i.CheckVersion("pact-mock-service", "2.0.0")

	if err == nil {
		t.Fatal("want error, got none")
	}
}

func TestInstaller_getVersionForBinary(t *testing.T) {
	version := "1.5.0"
	i := getInstaller(version, nil)
	v, err := i.GetVersionForBinary("pact-mock-service")

	if err != nil {
		t.Fatal("error:", err)
	}
	if v != version {
		t.Fatal("Want", version, "got", v)
	}
}

func TestInstaller_getVersionForBinaryError(t *testing.T) {
	i := getInstaller("", errors.New("test error"))
	_, err := i.GetVersionForBinary("pact-mock-service")

	if err == nil {
		t.Fatal("Want error, got nil")
	}
}

func TestInstaller_CheckInstallation(t *testing.T) {
	i := getInstaller("1.0.0", nil)
	err := i.CheckInstallation()

	if err != nil {
		t.Fatal("error:", err)
	}
}
func TestInstaller_CheckInstallationError(t *testing.T) {
	i := getInstaller("2.0.0", nil)
	err := i.CheckInstallation()

	if err == nil {
		t.Fatal("Want error, got nil")
	}
}
