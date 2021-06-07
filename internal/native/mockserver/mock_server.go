package mockserver

/*
// Library headers
#include <stdlib.h>
#include <stdint.h>
typedef int bool;
#define true 1
#define false 0

void init(char* log);
char* version();

/// Wraps a Pact model struct
typedef struct InteractionHandle InteractionHandle;

struct InteractionHandle {
	uintptr_t pact;
  uintptr_t interaction;
};

typedef struct MessageHandle MessageHandle;
struct MessageHandle {
	uintptr_t pact;
  uintptr_t message;
};

/// Wraps a Pact model struct
typedef struct PactHandle PactHandle;
struct PactHandle {
	uintptr_t pact;
	};

/// Wraps a PactMessage model struct
typedef struct MessagePactHandle MessagePactHandle;
struct MessagePactHandle {
  uintptr_t pact;
};

/// External interface to cleanup a mock server. This function will try terminate the mock server
/// with the given port number and cleanup any memory allocated for it. Returns true, unless a
/// mock server with the given port number does not exist, or the function panics.
///
/// **NOTE:** Although `close()` on the listener for the mock server is called, this does not
/// currently work and the listener will continue handling requests. In this
/// case, it will always return a 404 once the mock server has been cleaned up.
bool cleanup_mock_server(int mock_server_port);

/// External interface to create a mock server. A pointer to the pact JSON as a C string is passed in,
/// as well as the port for the mock server to run on. A value of 0 for the port will result in a
/// port being allocated by the operating system. The port of the mock server is returned.
///
/// # Errors
///
/// Errors are returned as negative values.
///
/// | Error | Description |
/// |-------|-------------|
/// | -1 | A null pointer was received |
/// | -2 | The pact JSON could not be parsed |
/// | -3 | The mock server could not be started |
/// | -4 | The method panicked |
/// | -5 | The address is not valid |
///
int create_mock_server(const char *pact_str, const char *addr_str, bool tls);

/// As above, but creates it for a PactHandle
int create_mock_server_for_pact(PactHandle pact, const char *addr_str, bool tls);

void with_specification(PactHandle pact, int specification_version);

/// Adds a provider state to the Interaction
void given(InteractionHandle interaction, const char *description);

/// Adds a provider state with params to the Interaction
void given_with_param(InteractionHandle interaction, const char *description, const char *name, const char *value);

/// Get self signed certificate for TLS mode
char* get_tls_ca_certificate();

/// Free a string allocated on the Rust heap
void free_string(const char *s);

/// External interface to check if a mock server has matched all its requests. The port number is
/// passed in, and if all requests have been matched, true is returned. False is returned if there
/// is no mock server on the given port, or if any request has not been successfully matched, or
/// the method panics.
bool mock_server_matched(int mock_server_port);

/// External interface to get all the mismatches from a mock server. The port number of the mock
/// server is passed in, and a pointer to a C string with the mismatches in JSON format is
/// returned.
///
/// **NOTE:** The JSON string for the result is allocated on the heap, and will have to be freed
/// once the code using the mock server is complete. The [`cleanup_mock_server`](fn.cleanup_mock_server.html) function is
/// provided for this purpose.
///
/// # Errors
///
/// If there is no mock server with the provided port number, or the function panics, a NULL
/// pointer will be returned. Don't try to dereference it, it will not end well for you.
///
char* mock_server_mismatches(int mock_server_port);

/// Creates a new Interaction and returns a handle to it
InteractionHandle new_interaction(PactHandle pact, const char *description);

/// Creates a new Pact model and returns a handle to it
PactHandle new_pact(const char *consumer_name, const char *provider_name);

/// Sets the description for the Interaction
void upon_receiving(InteractionHandle interaction, const char *description);

/// Sets the description for the Interaction
void with_request(InteractionHandle interaction, const char *method, const char *path);

/// Sets header expectations
/// https://docs.rs/pact_mock_server_ffi/0.0.7/pact_mock_server_ffi/fn.with_header.html
void with_header(InteractionHandle interaction, int interaction_part, const char *name, int index, const char *value);

/// Sets query string expectation
/// https://docs.rs/pact_mock_server_ffi/0.0.7/pact_mock_server_ffi/fn.with_query_parameter.html
void with_query_parameter(InteractionHandle interaction, const char *name, int index, const char *value);

/// Sets the description for the Interaction
// https://docs.rs/pact_mock_server_ffi/0.0.7/pact_mock_server_ffi/fn.with_body.html
void with_body(InteractionHandle interaction, int interaction_part, const char *content_type, const char *body);

void with_binary_file(InteractionHandle interaction, int interaction_part, const char *content_type, const char *body, int size);

int with_multipart_file(InteractionHandle interaction, int interaction_part, const char *content_type, const char *body, const char *part_name);

// https://docs.rs/pact_mock_server_ffi/0.0.7/pact_mock_server_ffi/fn.response_status.html
void response_status(InteractionHandle interaction, int status);

/// External interface to trigger a mock server to write out its pact file. This function should
/// be called if all the consumer tests have passed. The directory to write the file to is passed
/// as the second parameter. If a NULL pointer is passed, the current working directory is used.
///
/// Returns 0 if the pact file was successfully written. Returns a positive code if the file can
/// not be written, or there is no mock server running on that port or the function panics.
///
/// # Errors
///
/// Errors are returned as positive values.
///
/// | Error | Description |
/// |-------|-------------|
/// | 1 | A general panic was caught |
/// | 2 | The pact file was not able to be written |
/// | 3 | A mock server with the provided port was not found |
int write_pact_file(int mock_server_port, const char *directory);

void with_pact_metadata(PactHandle pact, const char *namespace, const char *name, const char *value);

// Additional global logging functions
// int log_to_buffer(int level);
// int log_to_stdout(int level);
// int log_to_file(const char *file_name, int level_filter);
// char* fetch_memory_buffer();

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

type interactionType int

const (
	INTERACTION_PART_REQUEST interactionType = iota
	INTERACTION_PART_RESPONSE
)

const (
	RESULT_OK interactionType = iota
	RESULT_FAILED
)

type specificationVersion int

const (
	SPECIFICATION_VERSION_UNKNOWN specificationVersion = iota
	SPECIFICATION_VERSION_V1
	SPECIFICATION_VERSION_V1_1
	SPECIFICATION_VERSION_V2
	SPECIFICATION_VERSION_V3
	SPECIFICATION_VERSION_V4
)

type logLevel int

const (
	LOG_LEVEL_OFF logLevel = iota
	LOG_LEVEL_ERROR
	LOG_LEVEL_WARN
	LOG_LEVEL_INFO
	LOG_LEVEL_DEBUG
	LOG_LEVEL_TRACE
)

// Pact is a Go representation of the PactHandle struct
type Pact struct {
	handle C.PactHandle
}

// Interaction is a Go representation of the InteractionHandle struct
type Interaction struct {
	handle C.InteractionHandle
}

// Version returns the current semver FFI interface version
func Version() string {
	v := C.version()

	return C.GoString(v)
}

// Init initialises the library
func Init() {
	log.Println("[DEBUG] initialising rust mock server interface")
	l := C.CString("LOG_LEVEL")
	defer free(l)

	C.init(l)

	// Alternative log destinations
	// NOTE: only one can be applied
	// logToBuffer(LOG_LEVEL_INFO)
	// logToFile("/tmp/pact.log", LOG_LEVEL_TRACE)
}

// MockServer is the public interface for managing the HTTP mock server
type MockServer struct {
	pact         *Pact
	messagePact  *MessagePact
	interactions []*Interaction
}

// NewMockServer creates a new mock server for a given consumer/provider
func NewHTTPMockServer(consumer string, provider string) *MockServer {
	cConsumer := C.CString(consumer)
	cProvider := C.CString(provider)
	defer free(cConsumer)
	defer free(cProvider)

	return &MockServer{pact: &Pact{handle: C.new_pact(cConsumer, cProvider)}}
}

// Version returns the current semver FFI interface version
func (m *MockServer) Version() string {
	return Version()
}

func (m *MockServer) WithSpecificationVersion(version specificationVersion) {
	C.with_specification(m.pact.handle, C.int(version))
}

// CreateMockServer creates a new Mock Server from a given Pact file.
// Returns the port number it started on or an error if failed
func (m *MockServer) CreateMockServer(pact string, address string, tls bool) (int, error) {
	log.Println("[DEBUG] mock server starting on address:", address)
	cPact := C.CString(pact)
	cAddress := C.CString(address)
	defer free(cPact)
	defer free(cAddress)
	tlsEnabled := 0
	if tls {
		tlsEnabled = 1
	}

	p := C.create_mock_server(cPact, cAddress, C.int(tlsEnabled))

	// | Error | Description |
	// |-------|-------------|
	// | -1 | A null pointer was received |
	// | -2 | The pact JSON could not be parsed |
	// | -3 | The mock server could not be started |
	// | -4 | The method panicked |
	// | -5 | The address is not valid |
	// | -6 | Could not create the TLS configuration with the self-signed certificate |
	port := int(p)
	switch port {
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
		if port > 0 {
			log.Println("[DEBUG] mock server running on port:", port)
			return port, nil
		}
		return port, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", port)
	}
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es. Does not write the pact or cleanup server.
func (m *MockServer) Verify(port int, dir string) (bool, []MismatchedRequest) {
	mismatches := m.MockServerMismatchedRequests(port)
	log.Println("[DEBUG] mock server mismatches:", len(mismatches))

	return len(mismatches) == 0, mismatches
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MockServer) MockServerMismatchedRequests(port int) []MismatchedRequest {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []MismatchedRequest

	mismatches := C.mock_server_mismatches(C.int(port))
	// This method can return a nil pointer, in which case, it
	// should be considered a failure (or at least, an issue)
	// converting it to a string might also do nasty things here!
	if mismatches == nil {
		log.Println("[WARN] received a null pointer from the native interface, returning empty list of mismatches")
		return []MismatchedRequest{}
	}

	json.Unmarshal([]byte(C.GoString(mismatches)), &res)

	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MockServer) CleanupMockServer(port int) bool {
	if len(m.interactions) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := C.cleanup_mock_server(C.int(port))

	return int(res) == 1
}

// WritePactFile writes the Pact to file.
func (m *MockServer) WritePactFile(port int, dir string) error {
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

// GetTLSConfig returns a tls.Config compatible with the TLS
// mock server
func GetTLSConfig() *tls.Config {
	cert := C.get_tls_ca_certificate()
	defer libRustFree(cert)

	goCert := C.GoString(cert)
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM([]byte(goCert))

	return &tls.Config{
		RootCAs: certPool,
	}
}

func free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func libRustFree(str *C.char) {
	C.free_string(str)
}

// Start starts up the mock HTTP server on the given address:port and TLS config
// https://docs.rs/pact_mock_server_ffi/0.0.7/pact_mock_server_ffi/fn.create_mock_server_for_pact.html
func (m *MockServer) Start(address string, tls bool) (int, error) {
	if len(m.interactions) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address)
	cAddress := C.CString(address)
	defer free(cAddress)
	tlsEnabled := 0
	if tls {
		tlsEnabled = 1
	}

	p := C.create_mock_server_for_pact(m.pact.handle, cAddress, C.int(tlsEnabled))

	// | Error | Description |
	// |-------|-------------|
	// | -1 | A null pointer was received |
	// | -2 | The pact JSON could not be parsed |
	// | -3 | The mock server could not be started |
	// | -4 | The method panicked |
	// | -5 | The address is not valid |
	// | -6 | Could not create the TLS configuration with the self-signed certificate |
	port := int(p)
	switch port {
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
		if port > 0 {
			log.Println("[DEBUG] mock server running on port:", port)
			return port, nil
		}
		return port, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", port)
	}
}

// Sets the additional metadata on the Pact file. Common uses are to add the client library details such as the name and version
func (m *MockServer) WithMetadata(namespace, k, v string) *MockServer {
	cNamespace := C.CString(namespace)
	defer free(cNamespace)
	cName := C.CString(k)
	defer free(cName)
	cValue := C.CString(v)
	defer free(cValue)

	C.with_pact_metadata(m.pact.handle, cNamespace, cName, cValue)

	return m
}

// NewInteraction initialises a new interaction for the current contract
func (m *MockServer) NewInteraction(description string) *Interaction {
	cDescription := C.CString(description)
	defer free(cDescription)

	i := &Interaction{
		handle: C.new_interaction(m.pact.handle, cDescription),
	}
	m.interactions = append(m.interactions, i)

	return i
}

func (i *Interaction) UponReceiving(description string) *Interaction {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.upon_receiving(i.handle, cDescription)

	return i
}

func (i *Interaction) Given(state string) *Interaction {
	cState := C.CString(state)
	defer free(cState)

	C.given(i.handle, cState)

	return i
}

func (i *Interaction) GivenWithParameter(state string, params map[string]interface{}) *Interaction {
	cState := C.CString(state)
	defer free(cState)

	for k, v := range params {
		cKey := C.CString(k)
		defer free(cKey)
		param := stringFromInterface(v)
		cValue := C.CString(param)
		defer free(cValue)

		C.given_with_param(i.handle, cState, cKey, cValue)

	}

	return i
}

func (i *Interaction) WithRequest(method string, pathOrMatcher interface{}) *Interaction {
	cMethod := C.CString(method)
	defer free(cMethod)

	path := stringFromInterface(pathOrMatcher)
	cPath := C.CString(path)
	defer free(cPath)

	C.with_request(i.handle, cMethod, cPath)

	return i
}

func (i *Interaction) WithRequestHeaders(valueOrMatcher map[string][]interface{}) *Interaction {
	return i.withHeaders(INTERACTION_PART_REQUEST, valueOrMatcher)
}

func (i *Interaction) WithResponseHeaders(valueOrMatcher map[string][]interface{}) *Interaction {
	return i.withHeaders(INTERACTION_PART_RESPONSE, valueOrMatcher)
}

func (i *Interaction) withHeaders(part interactionType, valueOrMatcher map[string][]interface{}) *Interaction {
	for k, v := range valueOrMatcher {

		cName := C.CString(k)
		defer free(cName)

		for _, header := range v {
			value := stringFromInterface(header)
			cValue := C.CString(value)
			defer free(cValue)

			C.with_header(i.handle, C.int(part), cName, C.int(0), cValue)
		}

	}

	return i
}

func (i *Interaction) WithQuery(valueOrMatcher map[string][]interface{}) *Interaction {
	for k, values := range valueOrMatcher {

		cName := C.CString(k)
		defer free(cName)

		for idx, v := range values {
			value := stringFromInterface(v)
			cValue := C.CString(value)
			defer free(cValue)

			C.with_query_parameter(i.handle, cName, C.int(idx), cValue)
		}
	}

	return i
}

func (i *Interaction) WithJSONRequestBody(body interface{}) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_REQUEST)
}

func (i *Interaction) WithJSONResponseBody(body interface{}) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_RESPONSE)
}

func (i *Interaction) withJSONBody(body interface{}, part interactionType) *Interaction {
	cHeader := C.CString("application/json")
	defer free(cHeader)

	jsonBody := stringFromInterface(body)
	cBody := C.CString(jsonBody)
	defer free(cBody)

	C.with_body(i.handle, C.int(part), cHeader, cBody)

	return i
}

func (i *Interaction) WithRequestBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 0)
}

func (i *Interaction) WithResponseBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 1)
}

func (i *Interaction) withBody(contentType string, body []byte, part interactionType) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cBody := C.CString(string(body))
	defer free(cBody)

	C.with_body(i.handle, C.int(part), cHeader, cBody)

	return i
}

func (i *Interaction) withBinaryBody(contentType string, body []byte, part interactionType) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	C.with_binary_file(i.handle, C.int(part), cHeader, (*C.char)(unsafe.Pointer(&body[0])), C.int(len(body)))

	return i
}

func (i *Interaction) WithBinaryRequestBody(body []byte) *Interaction {
	return i.withBinaryBody("application/octet-stream", body, INTERACTION_PART_REQUEST)
}

func (i *Interaction) WithBinaryResponseBody(body []byte) *Interaction {
	return i.withBinaryBody("application/octet-stream", body, INTERACTION_PART_RESPONSE)
}

func (i *Interaction) WithRequestMultipartFile(contentType string, filename string, mimePartName string) *Interaction {
	return i.withMultipartFile(contentType, filename, mimePartName, INTERACTION_PART_REQUEST)
}

func (i *Interaction) WithResponseMultipartFile(contentType string, filename string, mimePartName string) *Interaction {
	return i.withMultipartFile(contentType, filename, mimePartName, INTERACTION_PART_RESPONSE)
}

func (i *Interaction) withMultipartFile(contentType string, filename string, mimePartName string, part interactionType) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cPartName := C.CString(mimePartName)
	defer free(cPartName)

	cFilename := C.CString(filename)
	defer free(cFilename)

	C.with_multipart_file(i.handle, C.int(part), cHeader, cFilename, cPartName)

	return i
}

// Set the expected HTTTP response status
func (i *Interaction) WithStatus(status int) *Interaction {
	C.response_status(i.handle, C.int(status))

	return i
}

type stringLike interface {
	String() string
}

func stringFromInterface(obj interface{}) string {
	switch t := obj.(type) {
	case string:
		return t
	default:
		bytes, err := json.Marshal(obj)
		if err != nil {
			panic(fmt.Sprintln("unable to marshal body to JSON:", err))
		}
		return string(bytes)
	}
}

// Experimental logging options
// func logToBuffer(level logLevel) error {
// 	res := C.log_to_buffer(C.int(level))
// 	log.Println("[DEBUG] log_to_buffer res", res)

// 	return logResultToError(int(res))
// }

// func logToStdout(level logLevel) error {
// 	res := C.log_to_stdout(C.int(level))
// 	log.Println("[DEBUG] log_to_stdout res", res)

// 	return logResultToError(int(res))
// }

// func logToFile(file string, level logLevel) error {
// 	cFile := C.CString(file)
// 	defer free(cFile)

// 	res := C.log_to_file(cFile, C.int(level))
// 	log.Println("[DEBUG] log_to_file res", res)

// 	return logResultToError(int(res))
// }

// func getLogBuffer() string {
// 	buf := C.fetch_memory_buffer()
// 	defer free(buf)

// 	return C.GoString(buf)
// }

func logResultToError(res int) error {
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCantSetLogger
	case -2:
		return ErrNoLogger
	case -3:
		return ErrSpecifierNotUtf8
	case -4:
		return ErrUnknownSinkType
	case -5:
		return ErrMissingFilePath
	case -6:
		return ErrCantOpenSinkToFile
	case -7:
		return ErrCantConstructSink
	default:
		return fmt.Errorf("an unknown error ocurred when writing to pact file")
	}
}

// Errors
var (
	// ErrHandleNotFound indicates the underlying handle was not found, and a logic error in the framework
	ErrHandleNotFound = fmt.Errorf("unable to find the native interface handle (this indicates a defect in the framework)")

	// ErrMockServerPanic indicates a panic ocurred when invoking the remote Mock Server.
	ErrMockServerPanic = fmt.Errorf("a general panic occured when starting/invoking mock service (this indicates a defect in the framework)")

	// ErrUnableToWritePactFile indicates an error when writing the pact file to disk.
	ErrUnableToWritePactFile = fmt.Errorf("unable to write to file")

	// ErrMockServerNotfound indicates the Mock Server could not be found.
	ErrMockServerNotfound = fmt.Errorf("unable to find mock server with the given port")

	// ErrInvalidMockServerConfig indicates an issue configuring the mock server
	ErrInvalidMockServerConfig = fmt.Errorf("configuration for the mock server was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	// ErrInvalidPact indicates the pact file provided to the mock server was not a valid pact file
	ErrInvalidPact = fmt.Errorf("pact given to mock server is invalid")

	// ErrMockServerUnableToStart means the mock server could not be started in the rust library
	ErrMockServerUnableToStart = fmt.Errorf("unable to start the mock server")

	// ErrInvalidAddress means the address provided to the mock server was invalid and could not be understood
	ErrInvalidAddress = fmt.Errorf("invalid address provided to the mock server")

	// ErrMockServerTLSConfiguration indicates a TLS mock server could not be started
	// and is likely a framework level problem
	ErrMockServerTLSConfiguration = fmt.Errorf("a tls mock server could not be started (this is likely a defect in the framework)")

	// ErrNoInteractions indicates no Interactions have been registered to a mock server, and cannot be started/stopped until at least one is added
	ErrNoInteractions = fmt.Errorf("no interactions have been registered for the mock server")
)

// Log Errors
var (
	ErrCantSetLogger      = fmt.Errorf("can't set logger (applying the logger failed, perhaps because one is applied already).")
	ErrNoLogger           = fmt.Errorf("no logger has been initialized (call `logger_init` before any other log function).")
	ErrSpecifierNotUtf8   = fmt.Errorf("The sink specifier was not UTF-8 encoded.")
	ErrUnknownSinkType    = fmt.Errorf(`the sink type specified is not a known type (known types: "buffer", "stdout", "stderr", or "file /some/path").`)
	ErrMissingFilePath    = fmt.Errorf("no file path was specified in a file-type sink specification.")
	ErrCantOpenSinkToFile = fmt.Errorf("opening a sink to the specified file path failed (check permissions).")
	ErrCantConstructSink  = fmt.Errorf("can't construct the log sink")
)
