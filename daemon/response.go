package daemon

// Response contains the exit status and any message from a Service.
type Response struct {
	// System exit code from the command. Note that this will only even be 0 or 1.
	ExitCode int

	// Error message (if any) from the command.
	Message string
}
