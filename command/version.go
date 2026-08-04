package command

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is the Pact Go version. It is normally left as "dev" here and
// resolved automatically:
//   - Official release binaries (built via `goreleaser`) have this
//     overridden at build time via -ldflags (see .goreleaser.yml).
//   - Builds via `go install github.com/pact-foundation/pact-go/v2@vX.Y.Z`,
//     or when pact-go is imported as a module dependency, resolve it from
//     the Go module's build info at runtime, so no source change is needed.
//
// There is no version to bump for a release - see RELEASING.md.
var (
	Version    = "dev"
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Pact Go",
		Long:  `All software has versions. This is Pact Go's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pact Go CLI %s", Version)
		},
	}
)

func init() {
	if Version == "dev" {
		if v := moduleVersion(); v != "" {
			Version = v
		}
	}
	RootCmd.AddCommand(versionCmd)
}

// moduleVersion resolves the pact-go version from Go module build info,
// covering both `go install .../pact-go/v2@vX.Y.Z` (module is main) and
// pact-go being imported as a dependency of another module.
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	if isResolvedVersion(info.Main.Version) {
		return info.Main.Version
	}

	for _, dep := range info.Deps {
		if dep.Path == "github.com/pact-foundation/pact-go/v2" && isResolvedVersion(dep.Version) {
			return dep.Version
		}
	}

	return ""
}

func isResolvedVersion(v string) bool {
	return v != "" && v != "(devel)"
}
