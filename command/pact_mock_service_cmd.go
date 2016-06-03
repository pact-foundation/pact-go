package command

import (
	"os"
	"os/signal"

	"github.com/mefellows/pact-go/daemon"
	"github.com/spf13/cobra"
)

var mockServiceCmd = &cobra.Command{
	Use:   "mock",
	Short: "Runs the Pact Mock Server",
	Long:  `Runs the Pact Mock Server on a randomly available port`,
	Run: func(cmd *cobra.Command, args []string) {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		// Start the service
		svcManager := &daemon.PactMockService{}
		svcManager.Setup()
		_, svc := svcManager.NewService([]string{})
		svc.Start()

		// Block until a signal is received.
		<-c
	},
}

func init() {
	RootCmd.AddCommand(mockServiceCmd)
}
