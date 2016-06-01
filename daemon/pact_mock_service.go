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
	version := 2
	dir, _ := os.Getwd()
	dir = fmt.Sprintf(filepath.Join(dir, "../", "pacts"))
	logDir := fmt.Sprintf(filepath.Join(dir, "../", "logs"))
	log.Println("Starting mock service on port:", port)

	m.Args = []string{
		fmt.Sprintf("--port %d", port),
		fmt.Sprintf("--pact-specification-version %d", version),
		fmt.Sprintf("--pact-dir %s", dir),
		fmt.Sprintf("--log-dir %s", logDir),
		fmt.Sprintf("--ssl"),
	}
	m.Command = getCommandPath()
	return port, m
}

func getCommandPath() string {
	dir, _ := os.Getwd()
	return fmt.Sprintf(filepath.Join(dir, "pact-mock-service", "bin", "pact-mock-service"))
}
