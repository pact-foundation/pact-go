package types

// MockServer contains the RPC client interface to a Mock Server
type MockServer struct {
	Pid   int
	Port  int
	Error error
	Args  []string
}
