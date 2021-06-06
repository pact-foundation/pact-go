package command

import (
	"log"
	"os"

	"github.com/pact-foundation/pact-go/installer"

	"github.com/spf13/cobra"
)

var libDir string
var force bool
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Check required tools",
	Long:  "Checks versions of required Pact CLI tools for used by the library",
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		// Run the installer
		i, err := installer.NewInstaller()

		if libDir != "" {
			log.Println("[INFO] set lib dir target to", libDir)
			i.SetLibDir(libDir)
		}

		i.Force(force)

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
	installCmd.Flags().BoolVarP(&force, "force", "f", false, "Force a new installation")
	installCmd.Flags().StringVarP(&libDir, "libDir", "d", "", "Target directory to install the library")
	RootCmd.AddCommand(installCmd)
}
