package v4

// TransportConfig is the address of a plugin-provided mock server
// started by StartTransport, passed through to the integration test.
type TransportConfig struct {
	Port    int
	Address string
}
