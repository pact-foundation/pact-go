package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mefellows/pact-go/utils"
)

// PactMockService is a wrapper for the Pact Mock Service.
type PactMockService struct {
	ServiceManager
}

// NewService creates a new PactMockService with default settings.
func (m *PactMockService) NewService() (int, Service) {
	port, _ := utils.GetFreePort()
	log.Println("Starting mock service on port:", port)

	m.Args = []string{
		fmt.Sprintf("--port %d", port),
	}
	m.Command = getCommandPath()
	return port, m
}

func getCommandPath() string {
	dir, _ := os.Getwd()
	return fmt.Sprintf(filepath.Join(dir, "pact-mock-service", "bin", "pact-mock-service"))
}
