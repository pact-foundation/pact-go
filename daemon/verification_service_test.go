package daemon

import (
	"fmt"
	"testing"
)

func TestVerificationService_NewService(t *testing.T) {
	s := &VerificationService{}
	svc := s.NewService([]string{"--foo bar"})

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[0] != "--foo bar" {
		t.Fatalf(fmt.Sprintf(`Expected "--foo bar" argument to be passed, got "%s"`, s.Args[0]))
	}
}
