package v3

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

// Used to detect if logging has been configured.
var logFilter *logutils.LevelFilter
var defaultLogLevel = "INFO"

const (
	logLevelTrace logutils.LogLevel = "TRACE"
	logLevelDebug                   = "DEBUG"
	logLevelInfo                    = "INFO"
	logLevelWarn                    = "WARN"
	logLevelError                   = "ERROR"
)

func init() {
	if logFilter == nil {
		logFilter = &logutils.LevelFilter{
			Levels:   []logutils.LogLevel{logLevelTrace, logLevelDebug, logLevelInfo, logLevelWarn, logLevelError},
			MinLevel: logutils.LogLevel(defaultLogLevel),
			Writer:   os.Stderr,
		}
		log.SetOutput(logFilter)
	}
	log.Println("[DEBUG] initialised logging")
}

func setLogLevel(level logutils.LogLevel) {
	if level != "" {
		logFilter.SetMinLevel(level)
	}
}
