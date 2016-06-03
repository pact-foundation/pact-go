package command

import "testing"

func Test_PactMockServiceCommand(t *testing.T) {
	err := mockServiceCmd.Help()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}
