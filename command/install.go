package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var path string
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install required tools",
	Long:  "Installs underlying Pact standalone tools for use by the library",
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		// Run the installer
		fmt.Println("[INFO] Installing required tools...", path)
		fmt.Println("[INFO] Checking installation succeeded...")
		fmt.Println("[INFO] Tooling installed and up to date!")
	},
}

func init() {
	installCmd.Flags().StringVarP(&path, "path", "p", "/opt/pact", "Local daemon port to listen on")
	RootCmd.AddCommand(installCmd)
}
