package command

import "testing"

func init() {
	// os.Args = append(os.Args, "mock")
	// os.Args = append(os.Args, "--help")
}

func Test_PactMockServiceCommand(t *testing.T) {
	err := mockServiceCmd.Help()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// mockServiceCmd.Run(nil, os.Args)
}
