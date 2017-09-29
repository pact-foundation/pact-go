package utils

import (
	"net"
	"testing"
)

func Test_GetFreePort(t *testing.T) {
	port, err := GetFreePort()

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if port <= 0 {
		t.Fatalf("Expected a port > 0 to be available, got %d", port)
	}
}

func Test_FindPortInRange(t *testing.T) {
	cases := []struct {
		description string
		s           string
		port        int
		errorMsg    string
	}{
		{
			description: "single value",
			s:           "6667",
			port:        6667,
			errorMsg:    "",
		},
		{
			description: "csv",
			s:           "6667,6668,6669",
			port:        6667,
			errorMsg:    "",
		},
		{
			description: "range",
			s:           "6668-6669",
			port:        6668,
			errorMsg:    "",
		},
		{
			description: "invalid single",
			s:           "abc",
			port:        0,
			errorMsg:    `strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			description: "invalid lower range",
			s:           "abc-123",
			port:        0,
			errorMsg:    `strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			description: "invalid upper range",
			s:           "123-abc",
			port:        0,
			errorMsg:    `strconv.Atoi: parsing "abc": invalid syntax`,
		},
		{
			description: "invalid range",
			s:           "8888-7777",
			port:        0,
			errorMsg:    "invalid range passed",
		},
		{
			description: "double range",
			s:           "6668-6669,7000-7001",
			port:        0,
			errorMsg:    "invalid range passed",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			p, err := FindPortInRange(c.s)
			if err != nil && err.Error() != c.errorMsg {
				t.Fatalf("unexpected error %s", err.Error())
			} else if err == nil && c.errorMsg != "" {
				t.Fatalf("expected error %s", c.errorMsg)
			}
			if p != c.port {
				t.Fatalf("Expected port to be %d got %d", c.port, p)
			}
		})
	}
}

// Need to differentiate from above cases because this one requires
// us to use a port. Because of this the values must remain the same
func Test_FindPortInRangeWithUsedPorts(t *testing.T) {
	cases := []struct {
		description string
		s           string
		port        int
		errorMsg    string
	}{
		{
			description: "all ports used csv",
			s:           "6667",
			port:        0,
			errorMsg:    "all passed ports are unusable",
		},
		{
			description: "all ports used range",
			s:           "6667-6667",
			port:        0,
			errorMsg:    "all passed ports are unusable",
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			s := "localhost:6667"
			addr, err := net.ResolveTCPAddr("tcp", s)
			if err != nil {
				t.Fatalf("Could not resolve address %s in test", s)
			}

			l, err := net.ListenTCP("tcp", addr)
			if err != nil {
				t.Fatalf("Could not bind to port in test", s)
			}
			defer l.Close()
			p, err := FindPortInRange(c.s)
			if err != nil && err.Error() != c.errorMsg {
				t.Fatalf("unexpected error %s", err.Error())
			} else if err == nil && c.errorMsg != "" {
				t.Fatalf("expected error %s", c.errorMsg)
			}
			if p != c.port {
				t.Fatalf("Expected port to be %d got %d", c.port, p)
			}
		})
	}
}

func Test_checkPort(t *testing.T) {
	// Most cases tested above just have this one to test
	err := checkPort(-100)
	if err == nil {
		t.Fatalf("Expected error got none")
	}
}
