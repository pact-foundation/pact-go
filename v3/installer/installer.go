// Package installer is responsible for finding, acquiring and addressing
// runtime dependencies for this package (e.g. Ruby standalone, Rust bindings etc.)
package installer

import (
	"fmt"
	"log"
	"os/exec"
	"path"
	"runtime"
	"strings"

	getter "github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"

	// can't use these packages, because then the CLI installer wouldn't work - go won't run without it!
	// "github.com/pact-foundation/pact-go/v3/internal/native/verifier"
	// mockserver "github.com/pact-foundation/pact-go/v3/internal/native/mockserver"

	"github.com/spf13/afero"
)

// Installer manages the underlying Ruby installation
// (eventual) implementation requirements
// 1. Download OS specific artifacts if not pre-installed - DONE
// 1. Check the semver range of pre-installed artifacts - DONE
// 1. Enable global configuration (environment vars, config files, code options e.g. (`PACT_GO_SHARED_LIBRARY_PATH`))
// 1. Allow users to specify where they pre-install their artifacts (e.g. /usr/local/lib)

// Installer is used to check the Pact Go installation is setup correctly, and can automatically install
// packages if required
type Installer struct {
	downloader downloader
	os         string
	arch       string
	fs         afero.Fs
	libDir     string
}

type installerConfig func(*Installer) error

// NewInstaller creates a new initialised Installer
func NewInstaller(opts ...installerConfig) (*Installer, error) {
	i := &Installer{downloader: &defaultDownloader{}, fs: afero.NewOsFs()}

	for _, opt := range opts {
		opt(i)
	}

	if _, ok := supportedOSes[runtime.GOOS]; !ok {
		return nil, fmt.Errorf("%s is not a supported OS", runtime.GOOS)
	}
	i.os = supportedOSes[runtime.GOOS]

	if !strings.Contains(runtime.GOARCH, "64") {
		return nil, fmt.Errorf("%s is not a supported architecture, only 64 bit architectures are supported", runtime.GOARCH)
	}

	i.arch = x86_64
	if runtime.GOARCH != "amd64" {
		log.Println("[WARN] amd64 architecture not detected, behaviour may be undefined")
	}

	return i, nil
}

// CheckInstallation checks installation of all of the required libraries
// and downloads if they aren't present
func (i *Installer) CheckInstallation() error {

	// Check if files exist
	// --> Check versions of existing installed files
	if err := i.checkPackageInstall(); err == nil {
		return nil
	}

	// Check if override package path exists
	// -> if it does, copy files from existing location
	// --> Check versions of these files
	// --> copy files to lib dir

	// Download dependencies
	if err := i.downloadDependencies(); err != nil {
		return err
	}

	// Install dependencies
	if err := i.installDependencies(); err != nil {
		return err
	}

	if err := i.checkPackageInstall(); err != nil {
		return fmt.Errorf("unable to verify downloaded/installed dependencies: %s", err)
	}

	// --> Check if download is disabled (return error if downloads are disabled)
	// --> download files to lib dir

	return nil
}

func (i *Installer) getLibDir() string {
	if i.libDir == "" {
		return "/usr/local/lib"
	}

	return i.libDir
}

// checkPackageInstall discovers any existing packages, and checks installation of a given binary using semver-compatible checks
func (i *Installer) checkPackageInstall() error {
	for pkg, info := range packages {

		log.Println("[DEBUG] checking version for lib", info.libName, "semver range", info.semverRange)
		dst, _ := i.getLibDstForPackage(pkg)

		if _, err := i.fs.Stat(dst); err != nil {
			log.Println("[DEBUG] package", info.libName, "not found")
			return err
		}

		// if err := checkVersion(info.libName, info.testCommand(), info.semverRange); err != nil {
		// 	return err
		// }
	}

	return nil
}

// getVersionForBinary gets the version of a given Ruby binary
func (i *Installer) getVersionForBinary(binary string) (version string, err error) {
	log.Println("[DEBUG] running binary", binary)

	return "", nil
}

// TODO: checksums (they don't currently exist)
func (i *Installer) downloadDependencies() error {
	for pkg := range packages {
		src, err := i.getDownloadURLForPackage(pkg)

		if err != nil {
			return err
		}

		dst, err := i.getLibDstForPackage(pkg)

		if err != nil {
			return err
		}

		err = i.downloader.download(src, dst)

		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Installer) installDependencies() error {
	if i.os == osx {
		for pkg, info := range packages {
			log.Println("[INFO] setting install_name on library", info.libName, "for osx")

			dst, err := i.getLibDstForPackage(pkg)

			if err != nil {
				return err
			}

			err = setOSXInstallName(dst, info.libName)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// returns src
func (i *Installer) getDownloadURLForPackage(pkg string) (string, error) {
	pkgInfo, ok := packages[pkg]
	if !ok {
		return "", fmt.Errorf("unable to find package details for package: %s", pkg)
	}

	return fmt.Sprintf(downloadTemplate, pkg, pkgInfo.version, pkgInfo.libName, i.os, i.arch, osToExtension[i.os]), nil
}

func (i *Installer) getLibDstForPackage(pkg string) (string, error) {
	pkgInfo, ok := packages[pkg]
	if !ok {
		return "", fmt.Errorf("unable to find package details for package: %s", pkg)
	}

	return path.Join(i.getLibDir(), pkgInfo.libName) + "." + osToExtension[i.os], nil
}

var setOSXInstallName = func(file string, lib string) error {
	cmd := exec.Command("install_name_tool", "-id", fmt.Sprintf("/usr/local/lib/%s.dylib", lib), file)
	// cmd := exec.Command("install_name_tool", "-id", fmt.Sprintf("../../libs/%s.dylib", lib), file)
	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error setting install name on pact lib: %s", err)
	}

	log.Println("[DEBUG] output from command", stdoutStderr)

	return err
}

// download template structure: "https://github.com/pact-foundation/pact-reference/releases/download/PACKAGE-vVERSION/LIBNAME-OS-ARCH.EXTENSION.gz"
var downloadTemplate = "https://github.com/pact-foundation/pact-reference/releases/download/%s-v%s/%s-%s-%s.%s.gz"

var supportedOSes = map[string]string{
	"darwin": osx,
	windows:  windows,
	linux:    linux,
}

var osToExtension = map[string]string{
	windows: "dll",
	linux:   "so",
	osx:     "dylib",
}

type packageInfo struct {
	packageName string
	libName     string
	version     string
	semverRange string
	testCommand func() string
}

const (
	verifierPackage   = "pact_verifier_ffi"
	mockServerPackage = "libpact_mock_server_ffi"
	linux             = "linux"
	windows           = "windows"
	osx               = "osx"
	x86_64            = "x86_64"
)

var packages = map[string]packageInfo{
	verifierPackage: {
		libName:     "libpact_verifier_ffi",
		version:     "0.0.2",
		semverRange: ">= 0.0.2, < 1.0.0",
		// testCommand: func() string {
		// 	return (&verifier.Verifier{}).Version()
		// },
	},
	mockServerPackage: {
		libName:     "libpact_mock_server_ffi",
		version:     "0.0.15",
		semverRange: ">= 0.0.15, < 1.0.0",
		// testCommand: func() string {
		// 	return mockserver.Version()
		// },
	},
}

func checkVersion(lib, version, versionRange string) error {
	log.Println("[DEBUG] checking version", version, "of", lib, "against semver constraint", versionRange)

	v, err := goversion.NewVersion(version)
	if err != nil {
		return err
	}

	constraints, err := goversion.NewConstraint(versionRange)
	if err != nil {
		return err
	}

	if constraints.Check(v) {
		log.Println("[DEBUG]", v, "satisfies constraints", v, constraints)
		return nil
	}

	return fmt.Errorf("version %s of %s does not match constraint %s", version, lib, versionRange)
}

type downloader interface {
	download(src string, dst string) error
}

type defaultDownloader struct{}

func (d *defaultDownloader) download(src string, dst string) error {
	log.Println("[DEBUG] downloading library from", src, "to", dst)

	return getter.GetFile(dst, src)
}
