package types

// PactMockServer contains the RPC client interface to a Mock Server
type PactMockServer struct {
	Pid    int
	Port   int
	Status int
	Args   []string
}
