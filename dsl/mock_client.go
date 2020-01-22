package dsl

import (
	"github.com/pact-foundation/pact-go/types"
)

// Mock Client for testing the DSL package
type mockClient struct {
	VerifyProviderResponse   []types.ProviderVerifierResponse
	VerifyProviderError      error
	Servers                  []*types.MockServer
	StopServerResponse       *types.MockServer
	StopServerError          error
	RemoveAllServersResponse []*types.MockServer
	MockServer               *types.MockServer
	ReifyMessageResponse     *types.ReificationResponse
	ReifyMessageError        error
	UpdateMessagePactError   error
	PublishPactsError        error
}

func newMockClient() *mockClient {
	return &mockClient{
		MockServer: &types.MockServer{
			Pid:  0,
			Port: 0,
		},
		ReifyMessageResponse: &types.ReificationResponse{
			Response: map[string]string{
				"foo": "bar",
			},
		},
	}
}

// StartServer starts a remote Pact Mock Server.
func (p *mockClient) StartServer(args []string, port int) *types.MockServer {
	return p.MockServer
}

// ListServers lists all known Mock Servers
func (p *mockClient) ListServers() []*types.MockServer {
	return p.Servers
}

// StopServer stops a remote Pact Mock Server.
func (p *mockClient) StopServer(server *types.MockServer) (*types.MockServer, error) {
	return p.StopServerResponse, p.StopServerError
}

// RemoveAllServers stops all remote Pact Mock Servers.
func (p *mockClient) RemoveAllServers(server *types.MockServer) []*types.MockServer {
	return p.RemoveAllServersResponse
}

// VerifyProvider runs the verification process against a running Provider.
func (p *mockClient) VerifyProvider(request types.VerifyRequest) ([]types.ProviderVerifierResponse, error) {
	return p.VerifyProviderResponse, p.VerifyProviderError
}

// UpdateMessagePact adds a pact message to a contract file
func (p *mockClient) UpdateMessagePact(request types.PactMessageRequest) error {
	return p.UpdateMessagePactError
}

// ReifyMessage takes a structured object, potentially containing nested Matchers
// and returns an object with just the example (generated) content
// The object may be a simple JSON primitive e.g. string or number or a complex object
func (p *mockClient) ReifyMessage(request *types.PactReificationRequest) (res *types.ReificationResponse, err error) {
	return p.ReifyMessageResponse, p.ReifyMessageError
}

// PublishPacts publishes pacts to a broker
func (p *mockClient) PublishPacts(request types.PublishRequest) error {
	return p.PublishPactsError
}
