package native

/*
#cgo LDFLAGS: ${SRCDIR}/../../libs/libpact_mock_server_ffi.dylib

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

*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log"
)

// Request is the sub-struct of Mismatch
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// Mismatch is a type returned from the validation process
//
// [
//   {
//     "method": "GET",
//     "path": "/",
//     "request": {
//       "body": {
//         "pass": 1234,
//         "user": {
//           "address": "some address",
//           "name": "someusername",
//           "phone": 12345678,
//           "plaintext": "plaintext"
//         }
//       },
//       "method": "GET",
//       "path": "/"
//     },
//     "type": "missing-request"
//   }
// ]
type Mismatch struct {
	Request Request
	Type    string
}

// Init initialises the library
func Init() {
	log.Println("[DEBUG] initialising framework")
	C.init(C.CString(""))
}

// CreateMockServer creates a new Mock Server from a given Pact file.
func CreateMockServer(pact string, address string, tls bool) int {
	log.Println("[DEBUG] mock server starting")
	res := C.create_mock_server(C.CString(pact), C.CString(address), 0)
	log.Println("[DEBUG] mock server running on port:", res)
	return int(res)
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es. Does not write the pact or cleanup server.
func Verify(port int, dir string) (bool, []Mismatch) {
	res := C.mock_server_matched(C.int(port))

	mismatches := MockServerMismatches(port)
	log.Println("[DEBUG] mock server mismatches:", len(mismatches))

	return int(res) == 1, mismatches
}

// MockServerMismatches returns a JSON object containing any mismatches from
// the last set of interactions.
func MockServerMismatches(port int) []Mismatch {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []Mismatch

	mismatches := C.mock_server_mismatches(C.int(port))
	json.Unmarshal([]byte(C.GoString(mismatches)), &res)

	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func CleanupMockServer(port int) bool {
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := C.cleanup_mock_server(C.int(port))

	return int(res) == 1
}

var (
	// ErrMockServerPanic indicates a panic ocurred when invoking the remote Mock Server.
	ErrMockServerPanic = fmt.Errorf("a general panic occured when invoking mock service")

	// ErrUnableToWritePactFile indicates an error when writing the pact file to disk.
	ErrUnableToWritePactFile = fmt.Errorf("unable to write to file")

	// ErrMockServerNotfound indicates the Mock Server could not be found.
	ErrMockServerNotfound = fmt.Errorf("unable to find mock server with the given port")
)

// WritePactFile writes the Pact to file.
func WritePactFile(port int, dir string) error {
	log.Println("[DEBUG] pact verify on port:", port, ", dir:", dir)
	res := int(C.write_pact_file(C.int(port), C.CString(dir)))

	switch res {
	case 0:
		return nil
	case 1:
		return ErrMockServerPanic
	case 2:
		return ErrUnableToWritePactFile
	case 3:
		return ErrMockServerNotfound
	default:
		return fmt.Errorf("an unknown error ocurred when writing to pact file")
	}
}
