package log

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

var logFilter *logutils.LevelFilter
var defaultLogLevel = "INFO"

const (
	logLevelTrace logutils.LogLevel = "TRACE"
	logLevelDebug logutils.LogLevel = "DEBUG"
	logLevelInfo  logutils.LogLevel = "INFO"
	logLevelWarn  logutils.LogLevel = "WARN"
	logLevelError logutils.LogLevel = "ERROR"
)

func InitLogging() {
	if logFilter == nil {
		logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError},
			MinLevel: logutils.LogLevel(defaultLogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(logFilter)
		log.Println("[DEBUG] initialised logging")
	} else {
		log.Println("[WARN] log level cannot be set after initialising, changing will have no effect")
	}
}

// TODO: use the unified logging method to the FFI

// SetLogLevel sets the default log level for the Pact framework
func SetLogLevel(level logutils.LogLevel) error {
	InitLogging()

	if logFilter == nil {
		switch level {
		case logLevelTrace, logLevelDebug, logLevelError, logLevelInfo, logLevelWarn:
			logFilter.SetMinLevel(level)
			return nil
		default:
			return fmt.Errorf(`invalid logLevel '%s'. Please specify one of "TRACE", "DEBUG", "INFO", "WARN", "ERROR"`, level)
		}
	}
	return fmt.Errorf("log level ('%s') cannot be set to '%s' after initialisation", LogLevel(), level)
}

// LogLevel gets the current log level for the Pact framework
func LogLevel() logutils.LogLevel {
	if logFilter != nil {
		return logFilter.MinLevel
	}

	return logutils.LogLevel(defaultLogLevel)
}
