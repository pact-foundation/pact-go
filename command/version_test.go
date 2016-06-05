package command

import (
	"os"
	"testing"
)

func Test_VersionCommand(t *testing.T) {
	os.Args = []string{"version"}
	err := versionCmd.Execute()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	versionCmd.Run(nil, os.Args)
}
