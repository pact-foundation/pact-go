package provider

// Transport configures a way to connect to a given provider
type Transport struct {
	Scheme   string
	Protocol string
	Port     uint16
	Path     string
}
