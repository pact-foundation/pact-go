package types

import (
	"reflect"
	"testing"
)

func TestPactMessageRequest_Valid(t *testing.T) {
	s := &PactMessageRequest{
		Message: map[string]string{
			"user": "billy",
		},
		Consumer: "a",
		Provider: "b",
		PactDir:  "/path/to/pacts",
	}
	err := s.Validate()

	if err != nil {
		t.Fatal("want nil, got err:", err)
	}

	expected := []string{
		"update",
		`{"user":"billy"}`,
		"--consumer",
		"a",
		"--provider",
		"b",
		"--pact-dir",
		"/path/to/pacts",
		"--pact-specification-version",
		"3",
	}

	if !reflect.DeepEqual(s.Args, expected) {
		t.Fatalf("want %v got %v", expected, s.Args)
	}

}
