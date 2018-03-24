package install

import (
	"fmt"
	"testing"
)

// TODO: mock out the file system
func TestInstaller_CheckVersion(t *testing.T) {
	i := Installer{}
	err := i.CheckVersion("pact-mock-service", "2.7.3")

	if err != nil {
		t.Fatal("error:", err)
	}
}

func TestInstaller_CheckVersionFail(t *testing.T) {
	i := Installer{}
	err := i.CheckVersion("pact-mock-service", "3.7.3")

	if err == nil {
		t.Fatal("want error, got none")
	}
}

func TestInstaller_getVersionForBinary(t *testing.T) {
	i := Installer{}
	v, err := i.GetVersionForBinary("pact-mock-service")

	if err != nil {
		t.Fatal("error:", err)
	}

	fmt.Println("version: ", v)
}

func TestInstaller_CheckInstallation(t *testing.T) {
	i := Installer{}
	err := i.CheckInstallation()

	if err != nil {
		t.Fatal("error:", err)
	}
}
