package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v2.4.2"
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pact Go",
	Long:  `All software has versions. This is Pact Go's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pact Go CLI %s", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
