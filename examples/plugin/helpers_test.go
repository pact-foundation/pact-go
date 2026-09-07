package plugin

import (
	"fmt"
	"os"
	"strings"
)

var dir, _ = os.Getwd()

func generateMattMessage(message string) string {
	return fmt.Sprintf("MATT%sMATT\n", message)
}

func parseMattMessage(message string) string {
	return strings.TrimSpace(strings.ReplaceAll(message, "MATT", ""))
}
