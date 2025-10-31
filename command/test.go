package command

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/pact-foundation/pact-go/v2/installer"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:                "test",
	Short:              "Run go test with Pact library path configured",
	Long:               "Wrapper for 'go test' that sets PACT_LD_LIBRARY_PATH to the configured library directory",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		// Get the configured library directory
		libDir := installer.GetConfigLibDir()
		if libDir == "" {
			log.Println("[WARN] No library directory configured. Run 'pact-go install -d <path>' to set one.")
			// check PACT_LD_LIBRARY_PATH env var
			if envLibDir := os.Getenv("PACT_LD_LIBRARY_PATH"); envLibDir != "" {
				libDir = envLibDir
				log.Printf("[INFO] Using PACT_LD_LIBRARY_PATH from environment: %s", libDir)
			} else {
				log.Println("[ERROR] No library directory configured and PACT_LD_LIBRARY_PATH is not set. Cannot run tests.")
				os.Exit(1)
			}
		}

		// Set the PACT_LD_LIBRARY_PATH environment variable
		err := os.Setenv("PACT_LD_LIBRARY_PATH", libDir)
		if err != nil {
			log.Printf("[ERROR] Failed to set PACT_LD_LIBRARY_PATH: %v", err)
			os.Exit(1)
		}

		log.Printf("[INFO] Running go test with PACT_LD_LIBRARY_PATH=%s", libDir)

		// Prepare the go test command
		goTestArgs := append([]string{"test"}, args...)
		goCmd := exec.Command("go", goTestArgs...)

		// Pass through environment variables
		goCmd.Env = os.Environ()

		// Set up stdout, stderr, and stdin
		goCmd.Stdout = os.Stdout
		goCmd.Stderr = os.Stderr
		goCmd.Stdin = os.Stdin

		// Execute the command and exit with the same exit code
		if err := goCmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				// Exit with the same code as go test
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					os.Exit(status.ExitStatus())
				}
			}
			log.Printf("[ERROR] Failed to run go test: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(testCmd)
}
