package daemon

import (
	"fmt"
	"path/filepath"

	"github.com/kardianos/osext"
)

// MockService is a wrapper for the Pact Mock Service.
type MockService struct {
	ServiceManager
}

// NewService creates a new MockService with default settings.
func (m *MockService) NewService(args []string) Service {
	m.Args = []string{
		"service",
	}
	m.Args = append(m.Args, args...)

	m.Command = getMockServiceCommandPath()
	return m
}

func getMockServiceCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact", "bin", "pact-mock-service"))
}
