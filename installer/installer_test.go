// Package install contains functions necessary for installing and checking
// if the necessary underlying shared libs have been properly installed
package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNativeLibPath(t *testing.T) {
	lib := NativeLibPath()

	libFilePath := filepath.Join(lib, "lib.go")
	file, err := os.ReadFile(libFilePath)
	assert.NoError(t, err)
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
				name: "ffi lib - linux x86",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v%s/libpact_ffi-linux-x86_64.so.gz", packages[FFIPackage].version),
				test: Installer{
					os:   linux,
					arch: x86_64,
				},
			},
			{
				name: "ffi lib - osx x86",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v%s/libpact_ffi-osx-x86_64.dylib.gz", packages[FFIPackage].version),
				test: Installer{
					os:   osx,
					arch: x86_64,
				},
			},
			{
				name: "ffi lib - osx arm64",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v%s/libpact_ffi-osx-aarch64-apple-darwin.dylib.gz", packages[FFIPackage].version),
				test: Installer{
					os:   osx,
					arch: osx_aarch64,
				},
			},
			{
				name: "ffi lib - windows x86",
				pkg:  FFIPackage,
				want: fmt.Sprintf("https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v%s/pact_ffi-windows-x86_64.dll.gz", packages[FFIPackage].version),
				test: Installer{
					os:   windows,
					arch: x86_64,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				src, err := tt.test.getDownloadURLForPackage(tt.pkg)
				assert.NoError(t, err)
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

	t.Run("checks if existing libraries are present", func(t *testing.T) {
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

	t.Run("errors if installed versions are out of date", func(t *testing.T) {

	})

	t.Run("errors if installed versions are out of date", func(t *testing.T) {

	})
}

func TestInstallerCheckInstallation(t *testing.T) {
	t.Run("returns an error when existing libraries aren't present", func(t *testing.T) {
		i := &Installer{
			fs:         afero.NewMemMapFs(),
			downloader: &mockDownloader{},
			hasher:     &mockHasher{},
			config:     &mockConfiguration{},
			os:         "osx",
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
			os:         "osx",
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
		defer restoreOSXInstallName()()
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
			os:     "osx",
		}

		err := i.CheckInstallation()
		assert.NoError(t, err)
	})
}

type mockDownloader struct {
	src      string
	dst      string
	called   bool
	callFunc func()
}

func (m *mockDownloader) download(src, dst string) error {
	m.src = src
	m.dst = dst
	m.called = true
	if m.callFunc != nil {
		m.callFunc()
	}

	return nil
}

type mockHasher struct {
}

func (m *mockHasher) hash(src string) (string, error) {
	return "1234", nil
}

type mockConfiguration struct {
}

func (m *mockConfiguration) readConfig() pactConfig {
	return pactConfig{
		Libraries: make(map[string]packageMetadata),
	}
}

func (m *mockConfiguration) writeConfig(pactConfig) error {
	return nil
}

func restoreOSXInstallName() func() {
	old := setOSXInstallName
	setOSXInstallName = func(string) error {
		return nil
	}

	return func() {
		setOSXInstallName = old
	}
}

func TestUpdateConfiguration(t *testing.T) {

}
