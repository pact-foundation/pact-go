package daemon

import "testing"

func TestVerificationService_NewService(t *testing.T) {
	s := &VerificationService{}
	svc := s.NewService([]string{"--foo bar"})

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[0] != "--foo bar" {
		t.Fatalf("Expected '--foo bar' argument to be passed")
	}
}
