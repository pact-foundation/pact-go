package dsl

import (
	"errors"

	"github.com/pact-foundation/pact-go/types"
)

// Mock Client for testing the DSL package
type mockClient struct {
}

// StartServer starts a remote Pact Mock Server.
func (p *mockClient) StartServer(args []string, port int) *types.MockServer {
	return &types.MockServer{
		Pid:  0,
		Port: 0,
	}
}

// ListServers lists all known Mock Servers
func (p *mockClient) ListServers() []*types.MockServer {
	var servers []*types.MockServer

	return servers
}

// StopServer stops a remote Pact Mock Server.
func (p *mockClient) StopServer(server *types.MockServer) (*types.MockServer, error) {
	return nil, errors.New("failed stopping server")
}

// RemoveAllServers stops all remote Pact Mock Servers.
func (p *mockClient) RemoveAllServers(server *types.MockServer) *[]types.MockServer {
	return nil
}

// VerifyProvider runs the verification process against a running Provider.
func (p *mockClient) VerifyProvider(request types.VerifyRequest) (types.ProviderVerifierResponse, error) {
	return types.ProviderVerifierResponse{}, nil
}

// UpdateMessagePact adds a pact message to a contract file
func (p *mockClient) UpdateMessagePact(request types.PactMessageRequest) error {
	return nil
}

// ReifyMessage takes a structured object, potentially containing nested Matchers
// and returns an object with just the example (generated) content
// The object may be a simple JSON primitive e.g. string or number or a complex object
func (p *mockClient) ReifyMessage(request *types.PactReificationRequest) (res *types.ReificationResponse, err error) {
	return &types.ReificationResponse{
		Response: map[string]string{
			"foo": "bar",
		},
	}, nil
}
