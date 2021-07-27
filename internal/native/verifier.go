package native

/*
// Library headers
#include <stdlib.h>
typedef int bool;
#define true 1
#define false 0

char* pactffi_version();
void pactffi_free_string(char* s);
int pactffi_verify(char* s);
*/
import "C"

import (
	"fmt"
	"log"
	"strings"
)

type Verifier struct{}

func (v *Verifier) Verify(args []string) error {
	log.Println("[DEBUG] executing verifier FFI with args", args)
	cargs := C.CString(strings.Join(args, "\n"))
	defer free(cargs)
	result := C.pactffi_verify(cargs)

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The verification process failed, see output for errors |
	/// | 2 | A null pointer was received |
	/// | 3 | The method panicked |
	switch int(result) {
	case 0:
		return nil
	case 1:
		return ErrVerifierFailed
	case 2:
		return ErrInvalidVerifierConfig
	case 3:
		return ErrVerifierPanic
	default:
		return fmt.Errorf("an unknown error ocurred when verifying the provider (this indicates a defect in the framework")
	}
}

// Version returns the current semver FFI interface version
func (v *Verifier) Version() string {
	return Version()
}

var (
	// ErrVerifierPanic indicates a panic ocurred when invoking the verifier.
	ErrVerifierPanic = fmt.Errorf("a general panic occured when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier
	ErrInvalidVerifierConfig = fmt.Errorf("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	//ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful)
	ErrVerifierFailed = fmt.Errorf("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
)
