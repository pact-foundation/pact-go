// Package command contains the basic CLI commands to run Pact Go.
package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/spf13/cobra"
)

var verbose bool
var logLevel string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "pact-go",
	Short: "Pact Go makes it easier to work with Pact with Golang projects",
	Long: `Pact Go is a utility that wraps a number of external applications into
an idiomatic Golang interface and CLI, providing a mock service and DSL for
the consumer project, and interaction playback and verification for the
service provider project.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", true, "verbose output")
	RootCmd.PersistentFlags().StringVarP(&logLevel, "logLevel", "l", "INFO", "Set the logging level (DEBUG, INFO, ERROR)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setLogLevel(verbose bool, level string) {
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	if !verbose {
		log.SetOutput(ioutil.Discard)
	}
}
