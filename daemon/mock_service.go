package daemon

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/pact-foundation/pact-go/utils"
)

// MockService is a wrapper for the Pact Mock Service.
type MockService struct {
	ServiceManager

	// Extra field for only MockService
	Port int
}

// NewService creates a new MockService with default settings.
func (m *MockService) NewService(args []string) (int, Service) {
	var port int
	if m.Port != 0 {
		port = m.Port
	} else {
		port, _ = utils.GetFreePort()
	}
	log.Println("[DEBUG] starting mock service on port:", port)

	m.Args = []string{
		"service",
		"--port",
		fmt.Sprintf("%d", port),
	}
	m.Args = append(m.Args, args...)

	m.Command = getMockServiceCommandPath()
	return port, m
}

func getMockServiceCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact", "bin", "pact-mock-service"))
}
