// Package installer is responsible for finding, acquiring and addressing
// runtime dependencies for this package (e.g. Ruby standalone, Rust bindings etc.)
package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	"github.com/ulikunitz/xz"
	"gopkg.in/yaml.v3"

	"github.com/spf13/afero"
)

// NativeLibPath returns the absolute path to the go package used to link to the native rust library.
func NativeLibPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	pactRoot := filepath.Dir(filepath.Dir(file))
	return filepath.Join(pactRoot, "internal", "native")
}

// Installer is used to check the Pact Go installation is setup correctly, and can automatically install
// packages if required.
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

// NewInstaller creates a new initialised Installer.
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
		i.arch = aarch64
	default:
		i.arch = x86_64
		log.Println("[WARN] amd64 architecture not detected, defaulting to x86_64. Behaviour may be undefined")
	}

	return i, nil
}

// SetLibDir overrides the default library dir.
func (i *Installer) SetLibDir(dir string) {
	i.libDir = dir
}

// Force installs over the top.
func (i *Installer) Force(force bool) {
	i.force = force
}

// CheckInstallation checks installation of all of the required libraries
// and downloads if they aren't present.
func (i *Installer) CheckInstallation() error {
	// Check if files exist
	// --> Check if existing installed files
	if !i.force {
		err := i.CheckPackageInstall()
		if err == nil {
			return nil
		}
	}

	// Download dependencies
	err := i.downloadDependencies()
	if err != nil {
		return err
	}

	// Install dependencies
	err = i.installDependencies()
	if err != nil {
		return err
	}

	// Double check files landed correctly (can't execute 'version' call here,
	// because of dependency on the native libs we're trying to download!)
	err = i.CheckPackageInstall()
	if err != nil {
		return fmt.Errorf("unable to verify downloaded/installed dependencies: %w", err)
	}

	return nil
}

// CheckPackageInstall discovers any existing packages, and checks installation of a given binary using semver-compatible checks.
func (i *Installer) CheckPackageInstall() error {
	for pkg, info := range packages {
		dst, _ := i.getLibDstForPackage(pkg)

		_, err := i.fs.Stat(dst)
		if err != nil {
			log.Println("[INFO] package", info.libName, "not found")
			return fmt.Errorf("checking for %s at %s: %w", info.libName, dst, err)
		}
		log.Println("[INFO] package", info.libName, "found")

		lib, ok := i.config.readConfig().Libraries[pkg]
		if ok {
			err := checkVersion(info.libName, lib.Version, info.semverRange)
			if err != nil {
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
			continue
		}

		if lib, ok := LibRegistry[pkg]; ok {
			log.Println("[INFO] checking version", lib.Version(), "for lib", info.libName, "within semver range", info.semverRange)
			err := checkVersion(info.libName, lib.Version(), info.semverRange)
			if err != nil {
				return err
			}
		} else {
			log.Println("[DEBUG] unable to determine current version of package", pkg, "in LibRegistry", LibRegistry)
		}

		// Correct the configuration to reduce drift
		err = i.updateConfiguration(dst, pkg, info)
		if err != nil {
			return err
		}
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

// Download all dependencies, and update the pact-go configuration file.
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

		err = i.downloader.download(src, dst, i.getLibFileName())
		if err != nil {
			return err
		}

		//nolint:gosec // G302: the documented default install path is the system-wide
		// /usr/local/lib, installed via `sudo pact-go install` (README.md); the library must
		// stay world-readable so the non-root process that later dlopen's it (go test, the
		// compiled binary) can load it. A 0600 mode here would leave the file readable only
		// by the root user that ran the installer.
		err = os.Chmod(dst, 0o755)
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
	if i.os == macos {
		for pkg, info := range packages {
			log.Println("[INFO] setting install_name on library", info.libName, "for macos")

			dst, err := i.getLibDstForPackage(pkg)
			if err != nil {
				return err
			}

			err = setMacOSInstallName(dst)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// returns src.
func (i *Installer) getDownloadURLForPackage(pkg string) (string, error) {
	pkgInfo, ok := packages[pkg]
	if !ok {
		return "", fmt.Errorf("unable to find package details for package: %s", pkg)
	}

	target := rustTarget(i.os, i.arch, i.os == linux && checkMusl())
	return fmt.Sprintf(downloadTemplate, pkgInfo.releaseName, pkgInfo.version, pkgInfo.releaseName, target, osToArchiveExtension[i.os]), nil
}

// rustTarget returns the Rust target triple that names a release archive.
func rustTarget(osName, arch string, musl bool) string {
	switch osName {
	case macos:
		return arch + "-apple-darwin"
	case windows:
		return arch + "-pc-windows-msvc"
	default:
		if musl {
			return arch + "-unknown-linux-musl"
		}
		return arch + "-unknown-linux-gnu"
	}
}

// getLibFileName returns the shared library's file name, both inside the
// release archive and at its install destination.
func (i *Installer) getLibFileName() string {
	return osToLibName[i.os] + "." + osToExtension[i.os]
}

func (i *Installer) getLibDstForPackage(pkg string) (string, error) {
	_, ok := packages[pkg]
	if !ok {
		return "", fmt.Errorf("unable to find package details for package: %s", pkg)
	}

	return path.Join(i.getLibDir(), i.getLibFileName()), nil
}

// Write the metadata to reduce drift.
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

var setMacOSInstallName = func(file string) error {
	//nolint:gosec // G204: file is the local install destination built from getLibDstForPackage
	// (getLibDir + internal osToLibName/osToExtension maps). getLibDir can be overridden via
	// SetLibDir or PACT_GO_LIB_DOWNLOAD_PATH, so file is not purely internal, but exec.Command
	// here takes a fixed binary name with a fixed argv (no shell), so there is no injection
	// vector regardless of what file's value is.
	cmd := exec.Command("install_name_tool", "-id", file, file)
	log.Println("[DEBUG] running command:", cmd)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting install name on pact lib: %w", err)
	}

	log.Println("[DEBUG] output from command", string(stdoutStderr))

	return nil
}

func checkVersion(lib, version, versionRange string) error {
	log.Println("[INFO] checking version", version, "of", lib, "against semver constraint", versionRange)

	v, err := goversion.NewVersion(version)
	if err != nil {
		return fmt.Errorf("parsing version %q of %s: %w", version, lib, err)
	}

	constraints, err := goversion.NewConstraint(versionRange)
	if err != nil {
		return fmt.Errorf("parsing semver constraint %q for %s: %w", versionRange, lib, err)
	}

	if constraints.Check(v) {
		log.Println("[DEBUG]", v, "satisfies constraints", v, constraints)
		return nil
	}

	return fmt.Errorf("version %s of %s does not match constraint %s", version, lib, versionRange)
}

// checkMusl checks if the OS uses musl library instead of glibc.
func checkMusl() bool {
	lddPath, err := exec.LookPath("ldd")
	if err != nil {
		return false
	}

	//nolint:gosec // G204: lddPath is resolved via exec.LookPath("ldd"), a fixed binary name;
	// the second argument is a hardcoded literal. Neither is user- or network-supplied input.
	cmd := exec.CommandContext(context.Background(), lddPath, "/bin/echo")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	if strings.Contains(string(out), "musl") {
		return true
	}

	return false
}

// download template structure: "https://github.com/pact-foundation/pact-reference/releases/download/RELEASE-vVERSION/RELEASE-TARGET.ARCHIVE_EXTENSION"
var downloadTemplate = "https://github.com/pact-foundation/pact-reference/releases/download/%s-v%s/%s-%s.%s"

var supportedOSes = map[string]string{
	"darwin": macos,
	windows:  windows,
	linux:    linux,
}

var osToExtension = map[string]string{
	windows: "dll",
	linux:   "so",
	macos:   "dylib",
}

var osToArchiveExtension = map[string]string{
	windows: "zip",
	linux:   "tar.xz",
	macos:   "tar.xz",
}

// osToLibName names the shared library. On Windows only pact_ffi.dll is
// extracted: the archive's pact_ffi.lib is a static library, and MinGW's ld
// would pick it over the DLL for -lpact_ffi.
var osToLibName = map[string]string{
	windows: "pact_ffi",
	linux:   ffiLibBaseName,
	macos:   ffiLibBaseName,
}

type packageInfo struct {
	libName string
	// releaseName prefixes both the pact-reference release tag and its archive names.
	releaseName string
	version     string
	semverRange string
}

const (
	// FFIPackage is the package key for the pact_ffi shared library, used
	// to look it up in packages and LibRegistry.
	FFIPackage     = "libpact_ffi"
	downloadEnvVar = "PACT_GO_LIB_DOWNLOAD_PATH"
	linux          = "linux"
	windows        = "windows"
	macos          = "macos"
	x86_64         = "x86_64"
	aarch64        = "aarch64"
	// ffiLibBaseName is the shared library file's base name on the OSes
	// that don't rename it (linux, macos); it happens to equal FFIPackage
	// today, but names a different thing - a file prefix, not a package key.
	ffiLibBaseName = "libpact_ffi"
)

var packages = map[string]packageInfo{
	FFIPackage: {
		libName:     ffiLibBaseName,
		releaseName: "pact_ffi",
		version:     "0.5.12",
		// Pin to the shipped FFI minor. The bindings in this release reference
		// symbols specific to this libpact_ffi minor, and past minors have both
		// added and removed symbols, so a wider range lets a mismatched-but-in-range
		// library already on disk pass CheckPackageInstall and skip reinstalling,
		// which then fails at link/load time with undefined FFI symbols. Bump this
		// in lockstep with version above.
		semverRange: ">= 0.5.12, < 0.6.0",
	},
}

// Versioner reports the version of an already-loaded native library.
type Versioner interface {
	Version() string
}

// LibRegistry holds the native libraries actually loaded into this
// process, keyed by package (e.g. FFIPackage), so CheckPackageInstall can
// verify the loaded version against the installed package metadata. It is
// only populated when a native library has been loaded, such as during
// tests; the internal/checker package registers native.MockServer here.
var LibRegistry = map[string]Versioner{}

type downloader interface {
	// download fetches the release archive at src, verifies it against the
	// checksum published at src + ".sha256", and installs the archive entry
	// named member at dst.
	download(src string, dst string, member string) error
}

type defaultDownloader struct{}

func (d *defaultDownloader) download(src string, dst string, member string) error {
	log.Println("[INFO] downloading library from", src, "to", dst)

	baseDir := filepath.Dir(dst)
	//nolint:gosec // G301: the documented default install path is the system-wide
	// /usr/local/lib, installed via `sudo pact-go install` (README.md); this directory
	// (already present by default, only actually created here for a custom --libDir) must
	// stay world-readable/-executable so a non-root process can stat and dlopen the library
	// inside it. A 0750 mode here would leave a freshly created directory inaccessible to
	// anyone but the root user that ran the installer.
	err := os.MkdirAll(baseDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create %s; %w", baseDir, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	want, err := fetchChecksum(ctx, src)
	if err != nil {
		return err
	}

	archive, err := os.CreateTemp("", "pact-ffi-archive-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary archive file; %w", err)
	}
	defer func() {
		_ = archive.Close()
		_ = os.Remove(archive.Name())
	}()

	got, size, err := fetchArchive(ctx, src, archive)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("checksum mismatch for %s: expected %x, got %x", src, want, got)
	}

	entry, err := openArchiveMember(archive, size, src, member)
	if err != nil {
		return err
	}
	defer func() {
		_ = entry.Close()
	}()

	return installFile(entry, dst)
}

func httpGet(ctx context.Context, src string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s; %w", src, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed http call to %s; %w", src, err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("failed http call to %s; unexpected status %s", src, resp.Status)
	}

	return resp, nil
}

// fetchChecksum reads the SHA-256 digest of src from src + ".sha256", whose
// content is `<hex digest> *<archive name>`.
func fetchChecksum(ctx context.Context, src string) ([]byte, error) {
	checksumURL := src + ".sha256"
	resp, err := httpGet(ctx, checksumURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxChecksumFileSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read checksum from %s; %w", checksumURL, err)
	}

	fields := strings.Fields(string(body))
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty checksum file at %s", checksumURL)
	}
	if len(fields) > 1 && strings.TrimPrefix(fields[1], "*") != path.Base(src) {
		return nil, fmt.Errorf("checksum file at %s names %s, expected %s", checksumURL, fields[1], path.Base(src))
	}

	sum, err := hex.DecodeString(fields[0])
	if err != nil || len(sum) != sha256.Size {
		return nil, fmt.Errorf("malformed SHA-256 digest %q in %s", fields[0], checksumURL)
	}

	return sum, nil
}

// fetchArchive writes the archive at src into dst, returning its SHA-256
// digest and size.
func fetchArchive(ctx context.Context, src string, dst io.Writer) ([]byte, int64, error) {
	resp, err := httpGet(ctx, src)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	h := sha256.New()
	size, err := io.Copy(io.MultiWriter(dst, h), io.LimitReader(resp.Body, maxArchiveSize+1))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download %s; %w", src, err)
	}
	if size > maxArchiveSize {
		return nil, 0, fmt.Errorf("archive at %s exceeds the %d byte safety limit; aborting", src, maxArchiveSize)
	}

	return h.Sum(nil), size, nil
}

// openArchiveMember opens the regular file named member inside the archive,
// choosing the format from the extension of src. The member's path inside the
// archive never reaches the filesystem, since the caller writes it to a fixed
// destination.
func openArchiveMember(archive io.ReaderAt, size int64, src string, member string) (io.ReadCloser, error) {
	switch {
	case strings.HasSuffix(src, ".zip"):
		return openZipMember(archive, size, member)
	case strings.HasSuffix(src, ".tar.xz"):
		return openTarXzMember(io.NewSectionReader(archive, 0, size), member)
	default:
		return nil, fmt.Errorf("unsupported archive format for %s", src)
	}
}

func openZipMember(archive io.ReaderAt, size int64, member string) (io.ReadCloser, error) {
	zr, err := zip.NewReader(archive, size)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip archive; %w", err)
	}

	for _, f := range zr.File {
		if f.Name != member || !f.Mode().IsRegular() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s in zip archive; %w", member, err)
		}
		return rc, nil
	}

	return nil, fmt.Errorf("%s not found in zip archive", member)
}

func openTarXzMember(archive io.Reader, member string) (io.ReadCloser, error) {
	xr, err := xz.NewReader(archive)
	if err != nil {
		return nil, fmt.Errorf("failed to create xz reader; %w", err)
	}

	tr := tar.NewReader(&cappedReader{r: xr, remaining: maxExpandedArchiveSize})
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s not found in tar archive", member)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar archive; %w", err)
		}

		// The archive nests its files under a `pact_ffi-<target>/` directory.
		if path.Base(hdr.Name) != member || hdr.Typeflag != tar.TypeReg {
			continue
		}
		if hdr.Size > maxDecompressedLibSize {
			return nil, fmt.Errorf("decompressed library exceeds the %d byte safety limit; aborting", maxDecompressedLibSize)
		}

		return io.NopCloser(tr), nil
	}
}

// cappedReader fails once it has read more than remaining bytes, bounding the
// work spent skipping over tar entries the installer does not extract.
type cappedReader struct {
	r         io.Reader
	remaining int64
}

func (c *cappedReader) Read(p []byte) (int, error) {
	if c.remaining <= 0 {
		return 0, fmt.Errorf("decompressed archive exceeds the %d byte safety limit; aborting", maxExpandedArchiveSize)
	}
	if int64(len(p)) > c.remaining {
		p = p[:c.remaining]
	}

	n, err := c.r.Read(p)
	c.remaining -= int64(n)
	//nolint:wrapcheck // io.Reader callers compare against io.EOF, which wrapping would hide.
	return n, err
}

// installFile writes src to dst through a temporary file in the same
// directory, so a failed or oversized extraction leaves no partial library at
// dst for CheckPackageInstall to find.
func installFile(src io.Reader, dst string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+"-*")
	if err != nil {
		return fmt.Errorf("failed to create output file; %w", err)
	}
	// Removing an already-renamed temporary file is a no-op error.
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	written, err := io.Copy(tmp, io.LimitReader(src, maxDecompressedLibSize+1))
	if err != nil {
		return fmt.Errorf("failed to extract library; %w", err)
	}
	if written > maxDecompressedLibSize {
		return fmt.Errorf("decompressed library exceeds the %d byte safety limit; aborting", maxDecompressedLibSize)
	}

	// Close before renaming: Windows refuses to move an open file.
	err = tmp.Close()
	if err != nil {
		return fmt.Errorf("failed to write library to %s; %w", tmp.Name(), err)
	}

	err = os.Rename(tmp.Name(), dst)
	if err != nil {
		return fmt.Errorf("failed to move library into place at %s; %w", dst, err)
	}

	return nil
}

// downloadTimeout bounds a single library download end to end, so a stalled
// connection fails the install instead of hanging it. The largest FFI archive
// is ~33MB, so this leaves room for a slow link without leaving
// `pact-go install` waiting forever.
const downloadTimeout = 10 * time.Minute

// maxChecksumFileSize bounds the downloaded `.sha256` file, which holds one
// digest line.
const maxChecksumFileSize = 1024

// maxArchiveSize bounds the downloaded FFI archive. The largest archive
// pact-reference publishes (pact_ffi-x86_64-pc-windows-msvc.zip) is ~33MB,
// because each archive also carries the static library.
// The limits below are vars so tests can lower them.
var maxArchiveSize int64 = 128 * 1024 * 1024

// maxDecompressedLibSize bounds the extracted shared library. The largest one
// pact-reference publishes (libpact_ffi.so for aarch64 musl) is ~16MB, so
// 150MB leaves ample room for growth while still bounding a decompression bomb.
var maxDecompressedLibSize int64 = 150 * 1024 * 1024

// maxExpandedArchiveSize bounds the decompressed tar stream read while
// searching for the shared library. The static library dominates it, at
// ~130MB for the largest target.
var maxExpandedArchiveSize int64 = 1024 * 1024 * 1024

type packageMetadata struct {
	LibName string `yaml:"libname"`
	Version string `yaml:"version"`
	Hash    string `yaml:"hash"`
}

type pactConfig struct {
	Libraries map[string]packageMetadata `yaml:"libraries"`
}

type configReader interface {
	readConfig() pactConfig
}
type configWriter interface {
	writeConfig(c pactConfig) error
}

type configReadWriter interface {
	configReader
	configWriter
}

type configuration struct{}

func getConfigPath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return path.Join(user.HomeDir, ".pact", "pact-go.yml")
}

func (configuration) readConfig() pactConfig {
	pactConfigPath := getConfigPath()
	c := pactConfig{
		Libraries: map[string]packageMetadata{},
	}

	//nolint:gosec // G304: pactConfigPath is derived from the current OS user's home directory
	// (getConfigPath), not user- or network-supplied input.
	bytes, err := os.ReadFile(pactConfigPath)
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

	err := os.MkdirAll(filepath.Dir(pactConfigPath), 0o750)
	if err != nil {
		log.Println("[DEBUG] error creating pact config directory")
		return fmt.Errorf("creating pact config directory %s: %w", filepath.Dir(pactConfigPath), err)
	}

	bytes, err := yaml.Marshal(c)
	if err != nil {
		log.Println("[DEBUG] error marshalling YAML", pactConfigPath, "error: ", err)
		return fmt.Errorf("marshalling pact config: %w", err)
	}
	log.Println("[DEBUG] writing yaml config to file", string(bytes))

	err = os.WriteFile(pactConfigPath, bytes, 0o600)
	if err != nil {
		return fmt.Errorf("writing pact config to %s: %w", pactConfigPath, err)
	}

	return nil
}

type hasher interface {
	hash(src string) (string, error)
}

type defaultHasher struct{}

func (d *defaultHasher) hash(src string) (string, error) {
	log.Println("[DEBUG] obtaining hash for file", src)

	//nolint:gosec // G304: src is the local install destination built from getLibDstForPackage
	// (getLibDir + internal osToLibName/osToExtension maps). getLibDir can be overridden via
	// SetLibDir or PACT_GO_LIB_DOWNLOAD_PATH, so src is not purely internal, but this reads
	// back the exact file the installer itself just downloaded to that same path — the caller
	// who sets libDir already controls the filesystem the installer runs against.
	f, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("opening %s to hash it: %w", src, err)
	}
	defer func() {
		_ = f.Close()
	}()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", fmt.Errorf("reading %s to hash it: %w", src, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
