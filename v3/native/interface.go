// Package native contains the c bindings into the Pact Reference types.
// TODO: move this inter an internal only package
package native

/*
// Library headers
#include <stdlib.h>
typedef int bool;
#define true 1
#define false 0

void init(char* log);
int create_mock_server(char* pact, char* addr, bool tls);
int mock_server_matched(int port);
char* mock_server_mismatches(int port);
bool cleanup_mock_server(int port);
int write_pact_file(int port, char* dir);
char* get_tls_ca_certificate();
char* get_tls_key();
void free_string(char* s);

*/
import "C"

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"unsafe"
)

// Request is the sub-struct of Mismatch
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// MismatchRequest is a type returned from the validation process
// [
//   {
//     "method": "GET",
//     "mismatches": [
//       {
//         "actual": "",
//         "expected": "\"Bearer 1234\"",
//         "key": "Authorization",
//         "mismatch": "Expected header 'Authorization' but was missing",
//         "type": "HeaderMismatch"
//       }
//     ],
//     "path": "/foobar",
//     "type": "request-mismatch"
//   }
// ]
type MismatchDetail struct {
	Actual   string
	Expected string
	Key      string
	Mismatch string
	Type     string
}
type MismatchedRequest struct {
	Request
	Mismatches []MismatchDetail
	Type       string
}

// Init initialises the library
func Init() {
	log.Println("[DEBUG] initialising rust mock server interface")
	logLevel := C.CString("LOG_LEVEL")
	defer free(logLevel)

	C.init(logLevel)
}

// CreateMockServer creates a new Mock Server from a given Pact file.
// Returns the port number it started on or an error if failed
func CreateMockServer(pact string, address string, tls bool) (int, error) {
	log.Println("[DEBUG] mock server starting on address:", address)
	cPact := C.CString(pact)
	cAddress := C.CString(address)
	defer free(cPact)
	defer free(cAddress)
	tlsEnabled := 0
	if tls {
		tlsEnabled = 1
	}

	port := C.create_mock_server(cPact, cAddress, C.int(tlsEnabled))

	// | Error | Description |
	// |-------|-------------|
	// | -1 | A null pointer was received |
	// | -2 | The pact JSON could not be parsed |
	// | -3 | The mock server could not be started |
	// | -4 | The method panicked |
	// | -5 | The address is not valid |
	// | -6 | Could not create the TLS configuration with the self-signed certificate |
	switch int(port) {
	case -1:
		return 0, ErrInvalidMockServerConfig
	case -2:
		return 0, ErrInvalidPact
	case -3:
		return 0, ErrMockServerUnableToStart
	case -4:
		return 0, ErrMockServerPanic
	case -5:
		return 0, ErrInvalidAddress
	case -6:
		return 0, ErrMockServerTLSConfiguration
	default:
		log.Println("[DEBUG] mock server running on port:", port)
		return int(port), nil
	}
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es. Does not write the pact or cleanup server.
func Verify(port int, dir string) (bool, []MismatchedRequest) {
	res := C.mock_server_matched(C.int(port))

	mismatches := MockServerMismatchedRequests(port)
	log.Println("[DEBUG] mock server mismatches:", len(mismatches))

	return int(res) == 1, mismatches
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func MockServerMismatchedRequests(port int) []MismatchedRequest {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []MismatchedRequest

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
	ErrMockServerPanic = fmt.Errorf("a general panic occured when starting/invoking mock service")

	// ErrUnableToWritePactFile indicates an error when writing the pact file to disk.
	ErrUnableToWritePactFile = fmt.Errorf("unable to write to file")

	// ErrMockServerNotfound indicates the Mock Server could not be found.
	ErrMockServerNotfound = fmt.Errorf("unable to find mock server with the given port")

	ErrInvalidMockServerConfig = fmt.Errorf("ErrInvalidMockServerConfig")

	ErrInvalidPact = fmt.Errorf("pact given to mock server is invalid")

	ErrMockServerUnableToStart = fmt.Errorf("unable to start the mock server")

	ErrInvalidAddress = fmt.Errorf("invalid address provided to the mock server")

	ErrMockServerTLSConfiguration = fmt.Errorf("a tls mock server could not be started")
)

// WritePactFile writes the Pact to file.
func WritePactFile(port int, dir string) error {
	log.Println("[DEBUG] writing pact file for mock server on port:", port, ", dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	res := int(C.write_pact_file(C.int(port), cDir))

	// | Error | Description |
	// |-------|-------------|
	// | 1 | A general panic was caught |
	// | 2 | The pact file was not able to be written |
	// | 3 | A mock server with the provided port was not found |
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

func GetTLSConfig() *tls.Config {
	cert := C.get_tls_ca_certificate()
	defer libRustFree(cert)

	goCert := C.GoString(cert)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(goCert))

	return &tls.Config{
		RootCAs: certPool,
		// InsecureSkipVerify: true, // Disable SSL verification altogether?
	}
}

func free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func libRustFree(str *C.char) {
	C.free_string(str)
}
