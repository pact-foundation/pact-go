package command

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var mockServiceCmd = &cobra.Command{
	Use:   "mock",
	Short: "Runs the Pact Mock Servver",
	Long:  `Runs the Pact Mock Servver`,
	Run: func(cmd *cobra.Command, args []string) {
		runMockService()
		services.Wait()
	},
}

func init() {
	RootCmd.AddCommand(mockServiceCmd)
}

var services = sync.WaitGroup{}

func runMockService() {
	dir, _ := os.Getwd()
	cmdName := fmt.Sprintf(filepath.Join(dir, "pact-mock-service", "bin", "pact-mock-service"))
	// cmdName := fmt.Sprintf(filepath.Join(dir, "pact-provider-verifier", "bin", "pact-provider-verifier"))
	fmt.Println(cmdName)

	cmdArgs := []string{}

	services.Add(1)
	go func() {
		cmd := exec.Command(cmdName, cmdArgs...)
		// process, _ := os.StartProcess(cmdName, cmdArgs, &os.ProcAttr{})
		// process.Wait()

		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
			os.Exit(1)
		}

		cmdReaderErr, err := cmd.StderrPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("error: | %s\n", scanner.Text())
			}
		}()

		scanner2 := bufio.NewScanner(cmdReaderErr)
		go func() {
			for scanner2.Scan() {
				fmt.Printf("mock-service:  %s\n", scanner2.Text())
			}
		}()

		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
			os.Exit(1)
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
			os.Exit(1)
		}

		// TODO: Register mock service in a registry somewhere.....
	}()
}
