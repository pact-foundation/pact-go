// Package installer is responsible for finding, acquiring and addressing
// runtime dependencies for this package (e.g. Ruby standalone, Rust bindings etc.)
package installer

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	getter "github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"

	"crypto/md5"

	"github.com/spf13/afero"
)

// Installer is used to check the Pact Go installation is setup correctly, and can automatically install
// packages if required
type Installer struct {
	downloader downloader
	hasher     hasher
	config     configReadWriter
	os         string
	arch       string
	fs         afero.Fs
	libDir     string
	force      bool
}

type installerConfig func(*Installer) error

// NewInstaller creates a new initialised Installer
func NewInstaller(opts ...installerConfig) (*Installer, error) {
	i := &Installer{downloader: &defaultDownloader{}, fs: afero.NewOsFs(), hasher: &defaultHasher{}, config: &configuration{}}

	for _, opt := range opts {
		err := opt(i)
		if err != nil {
			log.Println("[ERROR] failure when configuring installer:", err)
			return nil, err
		}
	}

	if _, ok := supportedOSes[runtime.GOOS]; !ok {
		return nil, fmt.Errorf("%s is not a supported OS", runtime.GOOS)
	}
	i.os = supportedOSes[runtime.GOOS]

	if !strings.Contains(runtime.GOARCH, "64") {
		return nil, fmt.Errorf("%s is not a supported architecture, only 64 bit architectures are supported", runtime.GOARCH)
	}

	switch runtime.GOARCH {
	case "amd64":
		i.arch = x86_64
	case "arm64":
		if runtime.GOOS == "darwin" {
			i.arch = osx_aarch64
		} else {
			i.arch = aarch64
		}
	default:
		i.arch = x86_64
		log.Println("[WARN] amd64 architecture not detected, defaulting to x86_64. Behaviour may be undefined")
	}

	// Only perform a check if current OS is linux
	if runtime.GOOS == "linux" {
		err := i.checkMusl()
		if err != nil {
			log.Println("[DEBUG] unable to check for presence musl library due to error:", err)
		}
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
		if err := i.CheckPackageInstall(); err == nil {
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
	if err := i.CheckPackageInstall(); err != nil {
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

// CheckPackageInstall discovers any existing packages, and checks installation of a given binary using semver-compatible checks
func (i *Installer) CheckPackageInstall() error {
	for pkg, info := range packages {

		dst, _ := i.getLibDstForPackage(pkg)

		if _, err := i.fs.Stat(dst); err != nil {
			log.Println("[INFO] package", info.libName, "not found")
			return err
		} else {
			log.Println("[INFO] package", info.libName, "found")
		}

		lib, ok := i.config.readConfig().Libraries[pkg]
		if ok {
			if err := checkVersion(info.libName, lib.Version, info.semverRange); err != nil {
				return err
			}
			log.Println("[INFO] package", info.libName, "is correctly installed")
		} else {
			log.Println("[INFO] no package metadata information was found, run `pact-go install -f` to correct")
		}

		// This will only be populated during test when the ffi is loaded, but will actually test the FFI itself
		// It is helpful because it will prevent issues where the FFI is manually updated without using the `pact-go install` command
		if len(LibRegistry) == 0 {
			log.Println("[DEBUG] skip checking ffi version() call because FFI not loaded. This is expected when running the 'pact-go' command.")
		} else {
			lib, ok := LibRegistry[pkg]

			if ok {
				log.Println("[INFO] checking version", lib.Version(), "for lib", info.libName, "within semver range", info.semverRange)
				if err := checkVersion(info.libName, lib.Version(), info.semverRange); err != nil {
					return err
				}
			} else {
				log.Println("[DEBUG] unable to determine current version of package", pkg, "in LibRegistry", LibRegistry)
			}

			// Correct the configuration to reduce drift
			err := i.updateConfiguration(dst, pkg, info)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Download all dependencies, and update the pact-go configuration file
func (i *Installer) downloadDependencies() error {
	for pkg, pkgInfo := range packages {
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

		err = os.Chmod(dst, 0755)
		if err != nil {
			log.Println("[WARN] unable to set permissions on file", dst, "due to error:", err)
		}

		err = i.updateConfiguration(dst, pkg, pkgInfo)

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

			err = setOSXInstallName(dst)

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

	return fmt.Sprintf(downloadTemplate, pkg, pkgInfo.version, osToLibName[i.os], i.os, i.arch, osToExtension[i.os]), nil
}

func (i *Installer) getLibDstForPackage(pkg string) (string, error) {
	_, ok := packages[pkg]
	if !ok {
		return "", fmt.Errorf("unable to find package details for package: %s", pkg)
	}

	return path.Join(i.getLibDir(), osToLibName[i.os]) + "." + osToExtension[i.os], nil
}

// Write the metadata to reduce drift
func (i *Installer) updateConfiguration(dst string, pkg string, info packageInfo) error {
	// Get hash of file
	fmt.Println(i.hasher)
	hash, err := i.hasher.hash(dst)
	if err != nil {
		return err
	}

	// Read metadata
	c := i.config.readConfig()

	// Update config
	c.Libraries[pkg] = packageMetadata{
		LibName: info.libName,
		Version: info.version,
		Hash:    hash,
	}

	// Write metadata
	return i.config.writeConfig(c)
}

var setOSXInstallName = func(file string) error {
	cmd := exec.Command("install_name_tool", "-id", file, file)
	log.Println("[DEBUG] running command:", cmd)
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

// checkMusl checks if the OS uses musl library instead of glibc
func (i *Installer) checkMusl() error {
	lddPath, err := exec.LookPath("ldd")
	if err != nil {
		return fmt.Errorf("could not find ldd in environment path")
	}

	cmd := exec.Command(lddPath, "/bin/echo")
	out, err := cmd.CombinedOutput()

	if strings.Contains(string(out), "musl") {
		log.Println("[WARN] Usage of musl library is known to cause problems, prefer using glibc instead.")
	}

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

var osToLibName = map[string]string{
	windows: "pact_ffi",
	linux:   "libpact_ffi",
	osx:     "libpact_ffi",
}

type packageInfo struct {
	libName     string
	version     string
	semverRange string
}

const (
	FFIPackage     = "libpact_ffi"
	downloadEnvVar = "PACT_GO_LIB_DOWNLOAD_PATH"
	linux          = "linux"
	windows        = "windows"
	osx            = "osx"
	x86_64         = "x86_64"
	osx_aarch64    = "aarch64-apple-darwin"
	aarch64        = "aarch64"
)

var packages = map[string]packageInfo{
	FFIPackage: {
		libName:     "libpact_ffi",
		version:     "0.4.5",
		semverRange: ">= 0.4.0, < 1.0.0",
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

type packageMetadata struct {
	LibName string
	Version string
	Hash    string
}

type pactConfig struct {
	Libraries map[string]packageMetadata
}

type configReader interface {
	readConfig() pactConfig
}
type configWriter interface {
	writeConfig(pactConfig) error
}

type configReadWriter interface {
	configReader
	configWriter
}

type configuration struct{}

func getConfigPath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	return path.Join(user.HomeDir, ".pact", "pact-go.yml")
}

func (configuration) readConfig() pactConfig {
	pactConfigPath := getConfigPath()
	c := pactConfig{
		Libraries: map[string]packageMetadata{},
	}

	bytes, err := ioutil.ReadFile(pactConfigPath)
	if err != nil {
		log.Println("[DEBUG] error reading file", pactConfigPath, "error: ", err)
		return c
	}

	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		log.Println("[DEBUG] error unmarshalling YAML", pactConfigPath, "error: ", err)
	}
	return c
}

func (configuration) writeConfig(c pactConfig) error {
	log.Println("[DEBUG] writing config", c)
	pactConfigPath := getConfigPath()

	err := os.MkdirAll(filepath.Dir(pactConfigPath), 0755)
	if err != nil {
		log.Println("[DEBUG] error creating pact config directory")
		return err
	}

	bytes, err := yaml.Marshal(c)
	if err != nil {
		log.Println("[DEBUG] error marshalling YAML", pactConfigPath, "error: ", err)
		return err
	}
	log.Println("[DEBUG] writing yaml config to file", string(bytes))

	return ioutil.WriteFile(pactConfigPath, bytes, 0644)
}

type hasher interface {
	hash(src string) (string, error)
}

type defaultHasher struct{}

func (d *defaultHasher) hash(src string) (string, error) {
	log.Println("[DEBUG] obtaining hash for file", src)

	f, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
