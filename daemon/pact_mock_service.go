package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
	"github.com/pact-foundation/pact-go/utils"
)

// PactMockService is a wrapper for the Pact Mock Service.
type PactMockService struct {
	ServiceManager
}

// NewService creates a new PactMockService with default settings.
func (m *PactMockService) NewService(args []string) (int, Service) {
	port, _ := utils.GetFreePort()
	version := 2
	dir, _ := os.Getwd()
	logDir := fmt.Sprintf(filepath.Join(dir, "logs"))
	dir = fmt.Sprintf(filepath.Join(dir, "pacts"))
	log.Println("[DEBUG] starting mock service on port:", port)

	m.Args = []string{
		fmt.Sprintf("--port %d", port),
		fmt.Sprintf("--pact-specification-version %d", version),
		fmt.Sprintf("--pact-dir %s", dir),
		fmt.Sprintf("--log %s/pact.log", logDir),
	}
	m.Args = append(m.Args, args...)
	m.Command = getMockServiceCommandPath()
	return port, m
}

func getMockServiceCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact-mock-service", "bin", "pact-mock-service"))
}
