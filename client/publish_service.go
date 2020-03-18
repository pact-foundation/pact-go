package client

import (
	"log"
)

// PublishService is a wrapper for the Pact Provider Verifier Service.
type PublishService struct {
	ServiceManager
}

// NewService creates a new PublishService with default settings.
// Arguments allowed:
//
// 		--provider-base-url
// 		--pact-urls
// 		--provider-states-url
// 		--provider-states-setup-url
// 		--broker-username
// 		--broker-password
//    --publish-verification-results
//    --provider-app-version
//    --custom-provider-headers
func (v *PublishService) NewService(args []string) Service {
	log.Printf("[DEBUG] starting publish service with args: %v\n", args)

	v.Args = []string{
		"publish",
	}

	v.Args = append(v.Args, args...)
	v.Cmd = getPublisherCommandPath()

	return v
}

func getPublisherCommandPath() string {
	return "pact-broker"
}
