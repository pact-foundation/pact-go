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
}

// NewService creates a new MockService with default settings.
func (m *MockService) NewService(args []string) (int, Service) {
	port, _ := utils.GetFreePort()
	log.Println("[DEBUG] starting mock service on port:", port)

	m.Args = []string{
		"--port",
		fmt.Sprintf("%d", port),
	}
	m.Args = append(m.Args, args...)

	m.Command = getMockServiceCommandPath()
	return port, m
}

func getMockServiceCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact-mock-service", "bin", "pact-mock-service"))
}
