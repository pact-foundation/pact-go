// Package install contains functions necessary for installing and checking
// if the necessary underlying shared libs have been properly installed
package installer

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

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
				name: "mock server - linux x86",
				pkg:  mockServerPackage,
				want: "https://github.com/pact-foundation/pact-reference/releases/download/libpact_mock_server_ffi-v0.0.14/libpact_mock_server_ffi-linux-x86_64.so.gz",
				test: Installer{
					os:   linux,
					arch: x86_64,
				},
			},
			{
				name: "mock server - osx x86",
				pkg:  mockServerPackage,
				want: "https://github.com/pact-foundation/pact-reference/releases/download/libpact_mock_server_ffi-v0.0.14/libpact_mock_server_ffi-osx-x86_64.dylib.gz",
				test: Installer{
					os:   osx,
					arch: x86_64,
				},
			},
			{
				name: "mock server - linux x86",
				pkg:  mockServerPackage,
				want: "https://github.com/pact-foundation/pact-reference/releases/download/libpact_mock_server_ffi-v0.0.14/libpact_mock_server_ffi-windows-x86_64.dll.gz",
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

		err := i.downloadDependencies()

		assert.NoError(t, err)
		assert.True(t, mock.called)
	})

	t.Run("checks if existing libraries are present", func(t *testing.T) {
		oldPackages := packages
		defer func() { packages = oldPackages }()

		packages = map[string]packageInfo{
			verifierPackage: {
				libName:     "libpact_verifier_ffi",
				version:     "0.0.2",
				semverRange: ">= 0.8.3, < 1.0.0",
				testCommand: func() string {
					return "0.8.0"
				},
			},
			mockServerPackage: {
				libName:     "libpact_mock_server_ffi",
				version:     "0.0.14",
				semverRange: ">= 0.0.14, < 1.0.0",
				testCommand: func() string {
					return "0.0.13"
				},
			},
		}

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
			os:         "osx",
		}

		for pkg := range packages {
			dst, _ := i.getLibDstForPackage(pkg)
			mockFs.Create(dst)
		}

		err := i.CheckInstallation()
		assert.NoError(t, err)
	})

}

func TestInstallerCheckPackageInstall(t *testing.T) {
	t.Run("downloads and install dependencies when existing libraries aren't present", func(t *testing.T) {
		mockFs := afero.NewMemMapFs()
		var i *Installer

		i = &Installer{
			fs: mockFs,
			downloader: &mockDownloader{
				callFunc: func() {
					for pkg := range packages {
						dst, _ := i.getLibDstForPackage(pkg)
						mockFs.Create(dst)
					}
				},
			},
			os: "osx",
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
