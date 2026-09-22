// Package log configures the leveled logging used throughout Pact Go.
// It filters the standard library's log output by level (TRACE, DEBUG,
// INFO, WARN, ERROR), defaulting to INFO, or to the PACT_LOG_LEVEL or
// LOG_LEVEL environment variable if set.
package log

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

var (
	logFilter       *logutils.LevelFilter
	defaultLogLevel = "INFO"
)

const (
	logLevelTrace logutils.LogLevel = "TRACE"
	logLevelDebug logutils.LogLevel = "DEBUG"
	logLevelInfo  logutils.LogLevel = "INFO"
	logLevelWarn  logutils.LogLevel = "WARN"
	logLevelError logutils.LogLevel = "ERROR"
)

func init() {
	pactLogLevel := os.Getenv("PACT_LOG_LEVEL")
	logLevel := os.Getenv("LOG_LEVEL")

	level := defaultLogLevel
	if pactLogLevel != "" {
		level = pactLogLevel
	} else if logLevel != "" {
		level = logLevel
	}

	if logFilter == nil {
		logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError},
			MinLevel: logutils.LogLevel(level),
			Writer:   os.Stderr,
		}
		log.SetOutput(logFilter)
		log.Println("[DEBUG] initialised logging")
	}
}

// TODO: use the unified logging method to the FFI

// SetLogLevel sets the default log level for the Pact framework.
func SetLogLevel(level logutils.LogLevel) error {
	switch level {
	case logLevelTrace, logLevelDebug, logLevelError, logLevelInfo, logLevelWarn:
		logFilter.SetMinLevel(level)
		return nil
	default:
		return fmt.Errorf(`invalid logLevel '%s'. Please specify one of "TRACE", "DEBUG", "INFO", "WARN", "ERROR"`, level)
	}
}

// LogLevel gets the current log level for the Pact framework.
//
//nolint:revive // renaming would break the public API
func LogLevel() logutils.LogLevel {
	if logFilter != nil {
		return logFilter.MinLevel
	}

	return logutils.LogLevel(defaultLogLevel)
}

// PactCrash panics with err wrapped in a crash report asking the caller to
// open a bug report against Pact Go, for use where Pact Go cannot recover.
func PactCrash(err error) {
	log.Panicf(crashMessage, err.Error())
}

var crashMessage = `!!!!!!!!! PACT CRASHED !!!!!!!!!

%s

This is almost certainly a bug in Pact Go. It would be great if you could
open a bug report at: https://github.com/pact-foundation/pact-go/issues
so that we can fix it.

There is additional debugging information above. If you open a bug report, 
please rerun with SetLogLevel('trace') and include the
full output.

SECURITY WARNING: Before including your log in the issue tracker, make sure you
have removed sensitive info such as login credentials and urls that you don't want
to share with the world.

We're sorry about this!
`
