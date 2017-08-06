package daemon

import "testing"

func TestMockService_NewService(t *testing.T) {
	s := &MockService{}
	port, svc := s.NewService([]string{"--foo"})

	if port <= 0 {
		t.Fatalf("Expected non-zero port but got: %d", port)
	}

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[3] != "--foo" {
		t.Fatalf("Expected '--foo' argument to be passed")
	}
}
