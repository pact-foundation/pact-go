// Package types contains a number of structs common to the library.
package types

// CommandResponse contains the exit status and any message from running
// an external command / service.
type CommandResponse struct {
	// System exit code from the command. Note that this will only even be 0 or 1.
	ExitCode int

	// Error message (if any) from the command.
	Message string
}
