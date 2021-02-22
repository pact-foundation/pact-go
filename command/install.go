package command

import (
	"log"
	"os"

	"github.com/pact-foundation/pact-go/v3/installer"

	"github.com/spf13/cobra"
)

var path string
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Check required tools",
	Long:  "Checks versions of required Pact CLI tools for used by the library",
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		// Run the installer
		i, err := installer.NewInstaller()
		if err != nil {
			log.Println("[ERROR] Your Pact library installation is out of date and we were unable to download a newer one for you:", err)
			os.Exit(1)
		}

		if err = i.CheckInstallation(); err != nil {
			log.Println("[ERROR] Your Pact library installation is out of date and we were unable to download a newer one for you:", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
