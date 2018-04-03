package client

import (
	"log"
)

// MessageService is a wrapper for the Pact Message service.
type MessageService struct {
	ServiceManager
}

// NewService creates a new MessageService with default settings.
// Named Arguments allowed:
// 		--consumer
// 		--provider
//    --pact-dir
func (v *MessageService) NewService(args []string) Service {
	v.Args = args

	log.Printf("[DEBUG] starting message service with args: %v\n", v.Args)
	v.Cmd = "pact-message"

	return v
}
