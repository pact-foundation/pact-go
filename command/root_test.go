package command

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	os.Args = []string{"version"}
	err := RootCmd.Execute()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	Execute()
}

func TestRootCommand_LogLevel(t *testing.T) {
	res := captureOutput(func() {
		setLogLevel(true, "DEBUG")
		log.Println("[DEBUG] this should display")
	})

	if !strings.Contains(res, "[DEBUG] this should display") {
		t.Fatalf("Expected log message to contain '[DEBUG] this should display' but got '%s'", res)
	}

	res = captureOutput(func() {
		setLogLevel(true, "INFO")
		log.Print("[DEBUG] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}

	res = captureOutput(func() {
		setLogLevel(false, "INFO")
		log.Print("[DEBUG] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}
}

func captureOutput(action func()) string {
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	action()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stderr = rescueStderr

	return strings.TrimSpace(string(out))
}
