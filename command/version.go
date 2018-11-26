package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cliToolsVersion = "1.52.0"
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pact Go",
	Long:  `All software has versions. This is Pact Go's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pact Go CLI v1.0.0-beta.3, using CLI tools version", cliToolsVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
