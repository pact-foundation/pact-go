package daemon

import "testing"

func TestNewService(t *testing.T) {
	s := &PactMockService{}
	port, svc := s.NewService()

	if port <= 0 {
		t.Fatalf("Expected non-zero port but got: %d", port)
	}

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if len(svc.Args) != 1 {
		t.Fatalf("Expected 1 argument (--port) but got: %d", len(svc.Args))
	}
}
