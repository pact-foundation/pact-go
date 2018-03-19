package client

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

	m.Cmd = getMockServiceCommandPath()
	return m
}

func getMockServiceCommandPath() string {
	return "pact-mock-service"
}
