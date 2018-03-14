package command

import (
	"os"
	"testing"
)

func TestInstallCommand(t *testing.T) {
	os.Args = []string{"install"}
	err := installCmd.Execute()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	installCmd.Run(nil, os.Args)
}
