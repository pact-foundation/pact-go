package command

import (
	"log"
	"os"

	"github.com/pact-foundation/pact-go/v2/installer"

	"github.com/spf13/cobra"
)

var libDir string
var force bool
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install required libraries",
	Long:  "Install the correct version of required libraries",
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		// Run the installer
		i, err := installer.NewInstaller()

		if err != nil {
			log.Println("[ERROR] Your Pact library installation is out of date and we were unable to download a newer one for you:", err)
			os.Exit(1)
		}

		if libDir != "" {
			log.Println("[INFO] set lib dir target to", libDir)
			i.SetLibDir(libDir)
		}

		i.Force(force)

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
