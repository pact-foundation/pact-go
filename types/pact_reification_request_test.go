package types

import (
	"reflect"
	"testing"
)

func TestPactReificationRequest_Valid(t *testing.T) {
	s := &PactReificationRequest{
		Message: map[string]string{
			"user": "billy",
		},
	}
	err := s.Validate()

	if err != nil {
		t.Fatal("want nil, got err:", err)
	}

	expected := []string{
		"reify",
		`{"user":"billy"}`,
	}

	if !reflect.DeepEqual(s.Args, expected) {
		t.Fatalf("want %v got %v", expected, s.Args)
	}

}
