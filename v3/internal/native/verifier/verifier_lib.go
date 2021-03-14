package verifier

/*
#cgo darwin,amd64 LDFLAGS: -lpact_verifier_ffi
#cgo windows,amd64 LDFLAGS: -lpact_verifier_ffi
#cgo linux,amd64 LDFLAGS: -L/tmp -L/usr/local/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_verifier_ffi

// Mac OSX (until https://github.com/pact-foundation/pact-reference/pull/93 is done)
// install_name_tool -id "libpact_verifier_ffi.dylib" /usr/local/lib/libpact_verifier_ffi.dylib

// Library headers
typedef int bool;
#define true 1
#define false 0

void init(char* log);
char* version();
void free_string(char* s);
int verify(char* s);
*/
import "C"
