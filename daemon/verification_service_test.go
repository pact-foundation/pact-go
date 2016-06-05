package daemon

import "testing"

func TestVerificationService_NewService(t *testing.T) {
	s := &VerificationService{}
	port, svc := s.NewService([]string{"--foo bar"})

	if port != -1 {
		t.Fatalf("Expected port to be -1 but got: %d", port)
	}

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[0] != "--foo bar" {
		t.Fatalf("Expected '--foo bar' argument to be passed")
	}
}
