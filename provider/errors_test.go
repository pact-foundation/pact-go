package provider

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/v2/internal/native"
)

// TestErrVerificationFailed_IsNativeSentinel guards the re-export: the
// public provider.ErrVerificationFailed must remain identical to the
// internal/native sentinel so callers can use errors.Is to discriminate
// the error returned by Verifier.VerifyProvider.
func TestErrVerificationFailed_IsNativeSentinel(t *testing.T) {
	if !errors.Is(ErrVerificationFailed, native.ErrVerifierFailed) {
		t.Errorf("ErrVerificationFailed must be the same sentinel as internal/native.ErrVerifierFailed")
	}
	if !errors.Is(ErrVerifierFailedToRun, native.ErrVerifierFailedToRun) {
		t.Errorf("ErrVerifierFailedToRun must be the same sentinel as internal/native.ErrVerifierFailedToRun")
	}
}

// TestErrVerificationFailed_DiscriminatesFromGenericError makes sure the
// sentinels are not confused with arbitrary error values.
func TestErrVerificationFailed_DiscriminatesFromGenericError(t *testing.T) {
	generic := fmt.Errorf("something else broke")
	if errors.Is(generic, ErrVerificationFailed) {
		t.Error("generic error should not match ErrVerificationFailed")
	}
	if errors.Is(generic, ErrVerifierFailedToRun) {
		t.Error("generic error should not match ErrVerifierFailedToRun")
	}
}

// TestErrVerificationFailed_DistinctFromEachOther ensures the two sentinels
// are not interchangeable.
func TestErrVerificationFailed_DistinctFromEachOther(t *testing.T) {
	if errors.Is(ErrVerificationFailed, ErrVerifierFailedToRun) {
		t.Error("ErrVerificationFailed and ErrVerifierFailedToRun must be distinct")
	}
}
