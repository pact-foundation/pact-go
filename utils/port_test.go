package utils

import "testing"

func Test_GetFreePort(t *testing.T) {
	port, err := GetFreePort()

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if port <= 0 {
		t.Fatalf("Expected a port > 0 to be available, got %d", port)
	}
}
