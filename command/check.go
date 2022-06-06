package command

import (
	"log"
	"os"

	"github.com/pact-foundation/pact-go/v2/installer"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check required libraries",
	Long:  "Check the correct version of required libraries",
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

		if err = i.CheckPackageInstall(); err != nil {
			log.Println("[DEBUG] error from CheckPackageInstall:", err)
			log.Println("[ERROR] Your Pact library installation is out of date. Run `pact-go install` to correct")
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVarP(&libDir, "libDir", "d", "", "Target directory of the library installation")
	RootCmd.AddCommand(checkCmd)
}
