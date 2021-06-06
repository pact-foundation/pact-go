package verifier

/*
// Library headers
#include <stdlib.h>
typedef int bool;
#define true 1
#define false 0

void init(char* log);
char* version();
void free_string(char* s);
int verify(char* s);
*/
import "C"

import (
	"fmt"
	"log"
	"strings"
	"unsafe"
)

type Verifier struct{}

// Version returns the current semver FFI interface version
func Version() string {
	version := C.version()

	return C.GoString(version)
}

func Init() {
	log.Println("[DEBUG] initialising rust verifier interface")
	logLevel := C.CString("LOG_LEVEL")
	defer freeString(logLevel)

	C.init(logLevel)
}

// Version returns the current semver FFI interface version
func (v *Verifier) Version() string {
	return Version()
}

func (v *Verifier) Verify(args []string) error {
	log.Println("[DEBUG] executing verifier FFI with args", args)
	cargs := C.CString(strings.Join(args, "\n"))
	defer freeString(cargs)
	result := C.verify(cargs)

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

func freeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func free(p unsafe.Pointer) {
	C.free(p)
}

func libRustFree(str *C.char) {
	C.free_string(str)
}

var (
	// ErrVerifierPanic indicates a panic ocurred when invoking the verifier.
	ErrVerifierPanic = fmt.Errorf("a general panic occured when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier
	ErrInvalidVerifierConfig = fmt.Errorf("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	//ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful)
	ErrVerifierFailed = fmt.Errorf("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
)
