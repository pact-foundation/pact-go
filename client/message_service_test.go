package client

import "testing"

func TestMessageService_NewService(t *testing.T) {
	s := &MessageService{}
	svc := s.NewService([]string{"--foo"})

	if svc == nil {
		t.Fatalf("Expected a non-nil object but got nil")
	}

	if s.Args[0] != "--foo" {
		t.Fatalf("Expected '--foo' argument to be passed")
	}
}
