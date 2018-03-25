package client

import (
	"log"
)

// MessageVerificationService is a wrapper for the Pact Provider Verifier Service.
type MessageVerificationService struct {
	ServiceManager
}

// NewService creates a new MessageVerificationService with default settings.
// Named Arguments allowed:
// 		--consumer
// 		--provider
//    --pact-dir
func (v *MessageVerificationService) NewService(args []string) Service {
	v.Args = args
	// Currently has an issue, see https://travis-ci.org/pact-foundation/pact-message-ruby/builds/357675751
	// v.Args = []string{"update", `{ "description": "a test mesage", "content": { "name": "Mary" } }`, "--consumer", "from", "--provider", "golang", "--pact-dir", "/tmp"}

	log.Printf("[DEBUG] starting message service with args: %v\n", v.Args)
	v.Cmd = "pact-message"

	return v
}
