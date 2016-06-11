package types

// PactListResponse contains a list of all running Servers.
type PactListResponse struct {
	// System exit code from the Publish task.
	Servers []*MockServer
}
