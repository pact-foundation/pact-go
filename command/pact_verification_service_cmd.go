package command

import (
	"log"

	"github.com/pact-foundation/pact-go/daemon"
	"github.com/spf13/cobra"
)

var verificationServiceCmd = &cobra.Command{
	Use:   "verify",
	Short: "Runs the Pact Provider Verifier",
	Long:  `Runs the Pact Provider Verifier`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("[DEBUG] starting pact verification service")
		setLogLevel(verbose, logLevel)

		svcManager := &daemon.VerificationService{}
		svcManager.Setup()
		_, svc := svcManager.NewService([]string{})
		command := svc.Start()
		command.Wait()
	},
}

func init() {
	RootCmd.AddCommand(verificationServiceCmd)
}
