// Package installer is responsible for finding, acquiring and addressing
// runtime dependencies for this package (e.g. Ruby standalone, Rust bindings etc.)
package installer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	getter "github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"

	"github.com/spf13/afero"
)

// Installer is used to check the Pact Go installation is setup correctly, and can automatically install
// packages if required
type Installer struct {
	downloader downloader
	os         string
	arch       string
	fs         afero.Fs
	libDir     string
	force      bool
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

// SetLibDir overrides the default library dir
func (i *Installer) SetLibDir(dir string) {
	i.libDir = dir
}

// Force installs over the top
func (i *Installer) Force(force bool) {
	i.force = force
}

// CheckInstallation checks installation of all of the required libraries
// and downloads if they aren't present
func (i *Installer) CheckInstallation() error {

	// Check if files exist
	// --> Check if existing installed files
	if !i.force {
		if err := i.checkPackageInstall(); err == nil {
			return nil
		}
	}

	// Download dependencies
	if err := i.downloadDependencies(); err != nil {
		return err
	}

	// Install dependencies
	if err := i.installDependencies(); err != nil {
		return err
	}

	// Double check files landed correctly (can't execute 'version' call here,
	// because of dependency on the native libs we're trying to download!)
	if err := i.checkPackageInstall(); err != nil {
		return fmt.Errorf("unable to verify downloaded/installed dependencies: %s", err)
	}

	return nil
}

func (i *Installer) getLibDir() string {
	if i.libDir != "" {
		return i.libDir
	}

	env := os.Getenv(downloadEnvVar)
	if env != "" {
		return env
	}

	return "/usr/local/lib"
}

// checkPackageInstall discovers any existing packages, and checks installation of a given binary using semver-compatible checks
func (i *Installer) checkPackageInstall() error {
	for pkg, info := range packages {

		dst, _ := i.getLibDstForPackage(pkg)

		if _, err := i.fs.Stat(dst); err != nil {
			log.Println("[INFO] package", info.libName, "not found")
			return err
		} else {
			log.Println("[INFO] package", info.libName, "found")
		}

		lib, ok := LibRegistry[pkg]

		if ok {
			log.Println("[INFO] checking version", lib.Version(), "for lib", info.libName, "within semver range", info.semverRange)
			if err := checkVersion(info.libName, lib.Version(), info.semverRange); err != nil {
				return err
			}
		} else {
			log.Println("[DEBUG] unable to determine current version of package", pkg, "this is probably because the package is currently being installed")
		}
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
	cmd := exec.Command("install_name_tool", "-id", fmt.Sprintf("%s.dylib", lib), file)
	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error setting install name on pact lib: %s", err)
	}

	log.Println("[DEBUG] output from command", stdoutStderr)

	return err
}

func checkVersion(lib, version, versionRange string) error {
	log.Println("[INFO] checking version", version, "of", lib, "against semver constraint", versionRange)

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
	libName     string
	version     string
	semverRange string
}

const (
	VerifierPackage   = "pact_verifier_ffi"
	MockServerPackage = "libpact_mock_server_ffi"
	downloadEnvVar    = "PACT_GO_LIB_DOWNLOAD_PATH"
	linux             = "linux"
	windows           = "windows"
	osx               = "osx"
	x86_64            = "x86_64"
)

var packages = map[string]packageInfo{
	VerifierPackage: {
		libName:     "libpact_verifier_ffi",
		version:     "0.0.4",
		semverRange: ">= 0.0.2, < 1.0.0",
	},
	MockServerPackage: {
		libName:     "libpact_mock_server_ffi",
		version:     "0.0.17",
		semverRange: ">= 0.0.15, < 1.0.0",
	},
}

type Versioner interface {
	Version() string
}

var LibRegistry = map[string]Versioner{}

type downloader interface {
	download(src string, dst string) error
}

type defaultDownloader struct{}

func (d *defaultDownloader) download(src string, dst string) error {
	log.Println("[INFO] downloading library from", src, "to", dst)

	return getter.GetFile(dst, src)
}
