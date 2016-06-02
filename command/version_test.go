package command

import (
	"os"
	"testing"
)

func init() {
	// Set CLI flags to simulate real
	// os.Args = append(os.Args, "version")
	os.Args = []string{"version"}
}

func Test_VersionCommand(t *testing.T) {
	err := versionCmd.Execute()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	versionCmd.Run(nil, os.Args)
}
