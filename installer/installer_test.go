package installer

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ulikunitz/xz"
)

func TestNativeLibPath(t *testing.T) {
	lib := NativeLibPath()

	libFilePath := filepath.Join(lib, "lib.go")
	//nolint:gosec // G304: libFilePath is derived from NativeLibPath(), a fixed repo-relative
	// path, not user- or network-supplied input.
	file, err := os.ReadFile(libFilePath)
	require.NoError(t, err)
	assert.Contains(t, string(file), "-lpact_ffi")
}

// 1. Be able to specify the path of the binary in advance
// 2. Check if the correct versions of the libs are present???
// 3. Download the appropriate libs
// 4. Disable the check

func TestInstallerDownloader(t *testing.T) {
	t.Run("generates correct download URLs", func(t *testing.T) {
		tests := []struct {
			name string
			pkg  string
			want string
			test Installer
		}{
			{
				name: "ffi lib - linux x86_64",
				pkg:  FFIPackage,
				want: func() string {
					if checkMusl() {
						return fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-x86_64-unknown-linux-musl.tar.xz", packages[FFIPackage].version)
					}
					return fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-x86_64-unknown-linux-gnu.tar.xz", packages[FFIPackage].version)
				}(),
				test: Installer{
					os:   linux,
					arch: x86_64,
				},
			},
			{
				name: "ffi lib - linux aarch64",
				pkg:  FFIPackage,
				want: func() string {
					if checkMusl() {
						return fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-aarch64-unknown-linux-musl.tar.xz", packages[FFIPackage].version)
					}
					return fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-aarch64-unknown-linux-gnu.tar.xz", packages[FFIPackage].version)
				}(),
				test: Installer{
					os:   linux,
					arch: aarch64,
				},
			},
			{
				name: "ffi lib - macos x86_64",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-x86_64-apple-darwin.tar.xz", packages[FFIPackage].version),
				test: Installer{
					os:   macos,
					arch: x86_64,
				},
			},
			{
				name: "ffi lib - macos arm64",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-aarch64-apple-darwin.tar.xz", packages[FFIPackage].version),
				test: Installer{
					os:   macos,
					arch: aarch64,
				},
			},
			{
				name: "ffi lib - windows aarch64",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-aarch64-pc-windows-msvc.zip", packages[FFIPackage].version),
				test: Installer{
					os:   windows,
					arch: aarch64,
				},
			},
			{
				name: "ffi lib - windows x86_64",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/pact_ffi-v%s/pact_ffi-x86_64-pc-windows-msvc.zip", packages[FFIPackage].version),
				test: Installer{
					os:   windows,
					arch: x86_64,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				src, err := tt.test.getDownloadURLForPackage(tt.pkg)
				require.NoError(t, err)
				assert.Equal(t, tt.want, src)
			})
		}
	})

	t.Run("downloads the files when libs are not installed", func(t *testing.T) {
		mock := &mockDownloader{}

		i, _ := NewInstaller(func(i *Installer) error {
			i.downloader = mock

			return nil
		})

		_ = i.downloadDependencies() // This will actually error on the "chmod" if the file doesn't exist

		assert.True(t, mock.called)
	})

	t.Run("checks if existing libraries are present", func(_ *testing.T) {
		oldPackages := packages
		defer func() { packages = oldPackages }()

		packages = map[string]packageInfo{
			FFIPackage: {
				libName:     "libpact_ffi",
				version:     "0.0.0",
				semverRange: ">= 0.0.0, < 1.0.0",
			},
		}

		// TODO:
	})

	t.Run("errors if installed versions are out of date", func(_ *testing.T) {
	})

	t.Run("errors if installed versions are out of date", func(_ *testing.T) {
	})
}

func TestInstallerCheckInstallation(t *testing.T) {
	t.Run("returns an error when existing libraries aren't present", func(t *testing.T) {
		i := &Installer{
			fs:         afero.NewMemMapFs(),
			downloader: &mockDownloader{},
			hasher:     &mockHasher{},
			config:     &mockConfiguration{},
			os:         "macos",
		}
		err := i.CheckInstallation()

		assert.Error(t, err)
	})

	t.Run("returns nil when existing libraries are present", func(t *testing.T) {
		mockFs := afero.NewMemMapFs()
		i := &Installer{
			fs:         mockFs,
			downloader: &mockDownloader{},
			hasher:     &mockHasher{},
			config:     &mockConfiguration{},
			os:         "macos",
		}

		for pkg := range packages {
			dst, _ := i.getLibDstForPackage(pkg)
			_, _ = mockFs.Create(dst)
		}

		err := i.CheckInstallation()
		assert.NoError(t, err)
	})
}

func TestInstallerCheckPackageInstall(t *testing.T) {
	t.Run("downloads and install dependencies when existing libraries aren't present", func(t *testing.T) {
		defer restoreMacOSInstallName()()
		mockFs := afero.NewMemMapFs()

		var i *Installer

		i = &Installer{
			fs: mockFs,
			downloader: &mockDownloader{
				callFunc: func() {
					for pkg := range packages {
						dst, _ := i.getLibDstForPackage(pkg)
						_, _ = mockFs.Create(dst)
					}
				},
			},
			hasher: &mockHasher{},
			config: &mockConfiguration{},
			os:     "macos",
		}

		err := i.CheckInstallation()
		assert.NoError(t, err)
	})
}

type mockDownloader struct {
	src      string
	dst      string
	member   string
	called   bool
	callFunc func()
}

func (m *mockDownloader) download(src, dst, member string) error {
	m.src = src
	m.dst = dst
	m.member = member
	m.called = true
	if m.callFunc != nil {
		m.callFunc()
	}

	return nil
}

type mockHasher struct{}

func (m *mockHasher) hash(_ string) (string, error) {
	return "1234", nil
}

type mockConfiguration struct{}

func (m *mockConfiguration) readConfig() pactConfig {
	return pactConfig{
		Libraries: make(map[string]packageMetadata),
	}
}

func (m *mockConfiguration) writeConfig(pactConfig) error {
	return nil
}

func restoreMacOSInstallName() func() {
	old := setMacOSInstallName
	setMacOSInstallName = func(string) error {
		return nil
	}

	return func() {
		setMacOSInstallName = old
	}
}

func TestUpdateConfiguration(_ *testing.T) {
}

// TestPackagesVersionSatisfiesOwnSemverRange guards against version/semverRange
// drift: the FFI version compiled into each release must fall inside the semver
// range used to accept an already-installed library. If they drift apart, a
// fresh install of the shipped version can fail its own post-install check, or a
// stale, mismatched library can pass the check and then break at link/load time.
func TestPackagesVersionSatisfiesOwnSemverRange(t *testing.T) {
	for pkg, info := range packages {
		err := checkVersion(info.libName, info.version, info.semverRange)
		assert.NoErrorf(t, err, "package %q: compiled-in version %q must satisfy its own semverRange %q", pkg, info.version, info.semverRange)
	}
}

type archiveEntry struct {
	name    string
	body    []byte
	symlink bool
}

func tarXz(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	xw, err := xz.NewWriter(&buf)
	require.NoError(t, err)
	tw := tar.NewWriter(xw)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: tar.TypeReg}
		if e.symlink {
			hdr = &tar.Header{Name: e.name, Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink}
		}
		require.NoError(t, tw.WriteHeader(hdr))
		if !e.symlink {
			_, err = tw.Write(e.body)
			require.NoError(t, err)
		}
	}
	require.NoError(t, tw.Close())
	require.NoError(t, xw.Close())

	return buf.Bytes()
}

func zipped(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, err := zw.Create(e.name)
		require.NoError(t, err)
		_, err = w.Write(e.body)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())

	return buf.Bytes()
}

// serveRelease serves archive under name, with a `.sha256` file in the
// format pact-reference publishes, including its trailing blank line.
func serveRelease(t *testing.T, name string, archive []byte, checksum string) string {
	t.Helper()

	if checksum == "" {
		sum := sha256.Sum256(archive)
		checksum = hex.EncodeToString(sum[:])
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/" + name:
			_, _ = w.Write(archive)
		case "/" + name + ".sha256":
			_, _ = fmt.Fprintf(w, "%s *%s\n\n", checksum, name)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return server.URL + "/" + name
}

func assertNotExist(t *testing.T, dst string) {
	t.Helper()

	_, err := os.Stat(dst)
	assert.True(t, os.IsNotExist(err), "expected %s to be absent, got %v", dst, err)
}

func TestDefaultDownloader_Download(t *testing.T) {
	originalLib, originalArchive, originalExpanded := maxDecompressedLibSize, maxArchiveSize, maxExpandedArchiveSize
	defer func() {
		maxDecompressedLibSize, maxArchiveSize, maxExpandedArchiveSize = originalLib, originalArchive, originalExpanded
	}()
	maxDecompressedLibSize = 1024
	maxArchiveSize = 64 * 1024
	maxExpandedArchiveSize = 16 * 1024

	lib := bytes.Repeat([]byte{0x7f}, int(maxDecompressedLibSize))

	t.Run("extracts the library from a tar.xz archive", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", tarXz(t,
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/README.md", body: []byte("readme")},
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.so", body: lib},
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.a", body: []byte("static")},
		), "")

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		require.NoError(t, (&defaultDownloader{}).download(src, dst, "libpact_ffi.so"))

		//nolint:gosec // G304: dst is a test temp path.
		got, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, lib, got)
	})

	t.Run("extracts the library from the root of a zip archive", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-pc-windows-msvc.zip", zipped(t,
			archiveEntry{name: "pact_ffi.lib", body: []byte("static")},
			archiveEntry{name: "pact_ffi.dll", body: lib},
			archiveEntry{name: "pact_ffi.dll.lib", body: []byte("import")},
		), "")

		dst := filepath.Join(t.TempDir(), "pact_ffi.dll")
		require.NoError(t, (&defaultDownloader{}).download(src, dst, "pact_ffi.dll"))

		//nolint:gosec // G304: dst is a test temp path.
		got, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, lib, got)
	})

	t.Run("rejects an archive that does not match its checksum", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", tarXz(t,
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.so", body: lib},
		), strings.Repeat("00", sha256.Size))

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		err := (&defaultDownloader{}).download(src, dst, "libpact_ffi.so")

		require.ErrorContains(t, err, "checksum mismatch")
		assertNotExist(t, dst)
	})

	t.Run("fails when the checksum file is missing", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", nil, "")

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		err := (&defaultDownloader{}).download(src+"-missing", dst, "libpact_ffi.so")

		require.ErrorContains(t, err, "404")
		assertNotExist(t, dst)
	})

	t.Run("ignores a symlink with the library's name", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", tarXz(t,
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.so", symlink: true},
		), "")

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		err := (&defaultDownloader{}).download(src, dst, "libpact_ffi.so")

		require.ErrorContains(t, err, "libpact_ffi.so not found in tar archive")
		assertNotExist(t, dst)
	})

	t.Run("rejects a library beyond the size limit", func(t *testing.T) {
		big := bytes.Repeat([]byte{0x7f}, int(maxDecompressedLibSize)+1)
		for name, archive := range map[string][]byte{
			"pact_ffi-x86_64-unknown-linux-gnu.tar.xz": tarXz(t, archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.so", body: big}),
			"pact_ffi-x86_64-pc-windows-msvc.zip":      zipped(t, archiveEntry{name: "libpact_ffi.so", body: big}),
		} {
			t.Run(name, func(t *testing.T) {
				src := serveRelease(t, name, archive, "")

				dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
				err := (&defaultDownloader{}).download(src, dst, "libpact_ffi.so")

				require.ErrorContains(t, err, "exceeds the 1024 byte safety limit")
				assertNotExist(t, dst)
			})
		}
	})

	t.Run("rejects a tar stream beyond the expanded size limit", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", tarXz(t,
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.a", body: make([]byte, maxExpandedArchiveSize)},
			archiveEntry{name: "pact_ffi-x86_64-unknown-linux-gnu/libpact_ffi.so", body: lib},
		), "")

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		err := (&defaultDownloader{}).download(src, dst, "libpact_ffi.so")

		require.ErrorContains(t, err, "decompressed archive exceeds the 16384 byte safety limit")
		assertNotExist(t, dst)
	})

	t.Run("rejects an archive beyond the download size limit", func(t *testing.T) {
		src := serveRelease(t, "pact_ffi-x86_64-unknown-linux-gnu.tar.xz", make([]byte, maxArchiveSize+1), "")

		dst := filepath.Join(t.TempDir(), "libpact_ffi.so")
		err := (&defaultDownloader{}).download(src, dst, "libpact_ffi.so")

		require.ErrorContains(t, err, "exceeds the 65536 byte safety limit")
		assertNotExist(t, dst)
	})
}
