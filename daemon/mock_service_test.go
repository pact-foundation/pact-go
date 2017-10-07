package daemon

import "testing"

func TestMockService_NewService(t *testing.T) {
	s := &MockService{}
	svc := s.NewService([]string{"--foo"})

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[1] != "--foo" {
		t.Fatalf("Expected '--foo' argument to be passed")
	}
}
