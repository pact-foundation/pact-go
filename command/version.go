package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "v1.7.0"
var cliToolsVersion = "1.82.3"
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pact Go",
	Long:  `All software has versions. This is Pact Go's`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Pact Go CLI %s, using CLI tools version %s", version, cliToolsVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
