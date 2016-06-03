package daemon

import "testing"

func TestNewService(t *testing.T) {
	s := &PactMockService{}
	port, svc := s.NewService([]string{})

	if port <= 0 {
		t.Fatalf("Expected non-zero port but got: %d", port)
	}

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}
}
