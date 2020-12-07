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
	"log"
	"strings"
	"unsafe"
)

type Verifier struct{}

// Version returns the current semver FFI interface version
func (v *Verifier) Version() string {
	version := C.version()

	return C.GoString(version)
}

func (v *Verifier) Init() {
	log.Println("[DEBUG] initialising rust verifier interface")
	logLevel := C.CString("LOG_LEVEL")
	defer freeString(logLevel)

	C.init(logLevel)
}

func (v *Verifier) Verify(args []string) int {
	log.Println("[DEBUG] executing verifier FFI with args", args)
	cargs := C.CString(strings.Join(args, "\n"))
	defer freeString(cargs)
	result := C.verify(cargs)

	return int(result)
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
