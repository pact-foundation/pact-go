package command

import (
	"github.com/pact-foundation/pact-go/daemon"
	"github.com/spf13/cobra"
)

var port int
var daemonCmdInstance *daemon.Daemon
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Creates a daemon for the Pact DSLs to communicate with",
	Long:  `Creates a daemon for the Pact DSLs to communicate with`,
	Run: func(cmd *cobra.Command, args []string) {
		setLogLevel(verbose, logLevel)

		mock := &daemon.MockService{}
		mock.Setup()
		verifier := &daemon.VerificationService{}
		verifier.Setup()
		daemonCmdInstance = daemon.NewDaemon(mock, verifier)
		daemonCmdInstance.StartDaemon(port)
	},
}

func init() {
	daemonCmd.Flags().IntVarP(&port, "port", "p", 6666, "Local daemon port")
	RootCmd.AddCommand(daemonCmd)
}
