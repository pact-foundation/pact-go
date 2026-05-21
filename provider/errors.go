package provider

import "github.com/pact-foundation/pact-go/v2/internal/native"

// Sentinel errors returned by Verifier.VerifyProvider so callers can
// discriminate between a verification mismatch and an infrastructure failure
// of the verifier itself. Pact-go publishes verification results to the broker
// before returning ErrVerificationFailed (when PublishVerificationResults is
// enabled), so callers that gate via can-i-merge / can-i-deploy can choose to
// treat verification mismatches as non-fatal at the test step.
var (
	// ErrVerificationFailed is returned when one or more consumer pact
	// verifications did not match the provider's responses. Verification
	// results are published to the broker before this error is returned.
	ErrVerificationFailed = native.ErrVerifierFailed

	// ErrVerifierFailedToRun is returned when the verifier could not execute
	// at all — typically an infrastructure, configuration, or framework
	// problem. No verification results were published.
	ErrVerifierFailedToRun = native.ErrVerifierFailedToRun
)
