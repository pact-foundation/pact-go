package mockserver

/*
#cgo darwin,amd64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_mock_server_ffi
#cgo windows,amd64 LDFLAGS: -lpact_mock_server_ffi
#cgo linux,amd64 LDFLAGS: -L/tmp -L/opt/pact/lib -L/usr/local/lib -Wl,-rpath -Wl,/opt/pact/lib -Wl,-rpath -Wl,/tmp -Wl,-rpath -Wl,/usr/local/lib -lpact_mock_server_ffi

// Mac OSX (until https://github.com/pact-foundation/pact-reference/pull/93 is done)
//  install_name_tool -id "libpact_mock_server_ffi" libpact_mock_server_ffi.dylib

// NOTE: could also add custome multiple directories here with -L. It shows a warning, but might be supressable

// Library headers
typedef int bool;
#define true 1
#define false 0

void init(char* log);
int create_mock_server(char* pact, char* addr, bool tls);
int mock_server_matched(int port);
char* mock_server_mismatches(int port);
bool cleanup_mock_server(int port);
int write_pact_file(int port, char* dir);
void free_string(char* s);
char* get_tls_ca_certificate();
char* version();
*/
import "C"
