package version

import (
	"log"

	"github.com/pact-foundation/pact-go/v2/internal/checker"
)

// CheckVersion checks if the currently installed version is within semver range
// and will attempt to download the files to the default or configured directory if
// incorrect
func CheckVersion() {
	if err := checker.CheckInstall(); err != nil {
		log.Fatal("check version failed:", err)
	}

	log.Println("[DEBUG] version check completed")
}
