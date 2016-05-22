package command

import (
	"github.com/mefellows/pact-go/daemon"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Creates a daemon for the Pact DSLs to communicate with",
	Long:  `Creates a daemon for the Pact DSLs to communicate with`,
	Run: func(cmd *cobra.Command, args []string) {
		daemon.StartDaemon()
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
}
