package native

/*
#include "pact.h"
*/
import "C"

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"unsafe"
)

type interactionPart int

const (
	INTERACTION_PART_REQUEST interactionPart = iota
	INTERACTION_PART_RESPONSE
)

const (
	RESULT_OK interactionPart = iota
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

var logLevelStringToInt = map[string]logLevel{
	"OFF":   LOG_LEVEL_OFF,
	"ERROR": LOG_LEVEL_ERROR,
	"WARN":  LOG_LEVEL_WARN,
	"INFO":  LOG_LEVEL_INFO,
	"DEBUG": LOG_LEVEL_DEBUG,
	"TRACE": LOG_LEVEL_TRACE,
}

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
	v := C.pactffi_version()

	return C.GoString(v)
}

var loggingInitialised string

// Init initialises the library
func Init(logLevel string) {
	log.Println("[DEBUG] initialising native interface")
	logLevel = strings.ToUpper(logLevel)

	if loggingInitialised != "" {
		log.Printf("log level ('%s') cannot be set to '%s' after initialisation\n", loggingInitialised, logLevel)
	} else {
		l, ok := logLevelStringToInt[logLevel]
		if !ok {
			l = LOG_LEVEL_INFO
		}
		log.Printf("[DEBUG] initialised native log level to %s (%d)", logLevel, l)

		if os.Getenv("PACT_LOG_PATH") != "" {
			log.Println("[DEBUG] initialised native log to log to file:", os.Getenv("PACT_LOG_PATH"))
			err := logToFile(os.Getenv("PACT_LOG_PATH"), l)
			if err != nil {
				log.Println("[ERROR] failed to log to file:", err)
			}
		} else {
			log.Println("[DEBUG] initialised native log to log to stdout")
			err := logToStdout(l)
			if err != nil {
				log.Println("[ERROR] failed to log to stdout:", err)
			}
		}
	}
}

// MockServer is the public interface for managing the HTTP mock server
type MockServer struct {
	pact         *Pact
	messagePact  *MessagePact
	interactions []*Interaction
}

// NewHTTPPact creates a new HTTP mock server for a given consumer/provider
func NewHTTPPact(consumer string, provider string) *MockServer {
	cConsumer := C.CString(consumer)
	cProvider := C.CString(provider)
	defer free(cConsumer)
	defer free(cProvider)

	return &MockServer{pact: &Pact{handle: C.pactffi_new_pact(cConsumer, cProvider)}}
}

// Version returns the current semver FFI interface version
func (m *MockServer) Version() string {
	return Version()
}

func (m *MockServer) WithSpecificationVersion(version specificationVersion) {
	C.pactffi_with_specification(m.pact.handle, C.int(version))
}

// CreateMockServer creates a new Mock Server from a given Pact file.
// Returns the port number it started on or an error if failed
func (m *MockServer) CreateMockServer(pact string, address string, tls bool) (int, error) {
	log.Println("[DEBUG] mock server starting on address:", address)
	cPact := C.CString(pact)
	cAddress := C.CString(address)
	defer free(cPact)
	defer free(cAddress)
	tlsEnabled := false
	if tls {
		tlsEnabled = true
	}

	p := C.pactffi_create_mock_server(cPact, cAddress, C.bool(tlsEnabled))

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

	mismatches := C.pactffi_mock_server_mismatches(C.int(port))
	// This method can return a nil pointer, in which case, it
	// should be considered a failure (or at least, an issue)
	// converting it to a string might also do nasty things here!
	if mismatches == nil {
		log.Println("[WARN] received a null pointer from the native interface, returning empty list of mismatches")
		return []MismatchedRequest{}
	}

	err := json.Unmarshal([]byte(C.GoString(mismatches)), &res)
	if err != nil {
		log.Println("[ERROR] failed to unmarshal mismatches response, returning empty list of mismatches")
		return []MismatchedRequest{}
	}
	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MockServer) CleanupMockServer(port int) bool {
	if len(m.interactions) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := C.pactffi_cleanup_mock_server(C.int(port))

	return bool(res)
}

// WritePactFile writes the Pact to file.
// TODO: expose overwrite
func (m *MockServer) WritePactFile(port int, dir string) error {
	log.Println("[DEBUG] writing pact file for mock server on port:", port, ", dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	// overwritePact := 0
	// if overwrite {
	// 	overwritePact = 1
	// }

	// res := int(C.pactffi_write_pact_file(C.int(port), cDir, C.int(overwritePact)))
	res := int(C.pactffi_write_pact_file(C.int(port), cDir, C.bool(false)))

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
	cert := C.pactffi_get_tls_ca_certificate()
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
	C.pactffi_free_string(str)
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
	tlsEnabled := false
	if tls {
		tlsEnabled = true
	}

	p := C.pactffi_create_mock_server_for_pact(m.pact.handle, cAddress, C.bool(tlsEnabled))

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

// StartTransport starts up a mock server on the given address:port for the given transport
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MockServer) StartTransport(transport string, address string, port int, config map[string][]interface{}) (int, error) {
	if len(m.interactions) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address, port)
	cAddress := C.CString(address)
	defer free(cAddress)

	cTransport := C.CString(transport)
	defer free(cTransport)

	configJson := stringFromInterface(config)
	cConfig := C.CString(configJson)
	defer free(cConfig)

	p := C.pactffi_create_mock_server_for_transport(m.pact.handle, cAddress, C.ushort(port), cTransport, cConfig)

	// | Error | Description
	// |-------|-------------
	// | -1	   | An invalid handle was received. Handles should be created with pactffi_new_pact
	// | -2	   | transport_config is not valid JSON
	// | -3	   | The mock server could not be started
	// | -4	   | The method panicked
	// | -5	   | The address is not valid
	msPort := int(p)
	switch msPort {
	case -1:
		return 0, ErrInvalidMockServerConfig
	case -2:
		return 0, ErrInvalidMockServerConfig
	case -3:
		return 0, ErrMockServerUnableToStart
	case -4:
		return 0, ErrMockServerPanic
	case -5:
		return 0, ErrInvalidAddress
	default:
		if msPort > 0 {
			log.Println("[DEBUG] mock server running on port:", msPort)
			return msPort, nil
		}
		return msPort, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", msPort)
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

	C.pactffi_with_pact_metadata(m.pact.handle, cNamespace, cName, cValue)

	return m
}

// NewInteraction initialises a new interaction for the current contract
func (m *MockServer) UsingPlugin(pluginName string, pluginVersion string) error {
	cPluginName := C.CString(pluginName)
	defer free(cPluginName)
	cPluginVersion := C.CString(pluginVersion)
	defer free(cPluginVersion)

	r := C.pactffi_using_plugin(m.pact.handle, cPluginName, cPluginVersion)

	// 1 - A general panic was caught.
	// 2 - Failed to load the plugin.
	// 3 - Pact Handle is not valid.
	res := int(r)
	switch res {
	case 1:
		return ErrPluginGenericPanic
	case 2:
		return ErrPluginFailed
	case 3:
		return ErrHandleNotFound
	default:
		if res != 0 {
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code:", res)
		}
	}

	return nil
}

// NewInteraction initialises a new interaction for the current contract
func (m *MockServer) CleanupPlugins() {
	C.pactffi_cleanup_plugins(m.pact.handle)
}

// NewInteraction initialises a new interaction for the current contract
func (m *MockServer) NewInteraction(description string) *Interaction {
	cDescription := C.CString(description)
	defer free(cDescription)

	i := &Interaction{
		handle: C.pactffi_new_interaction(m.pact.handle, cDescription),
	}
	m.interactions = append(m.interactions, i)

	return i
}

// NewInteraction initialises a new interaction for the current contract
func (i *Interaction) WithPluginInteractionContents(part interactionPart, contentType string, contents string) error {
	cContentType := C.CString(contentType)
	defer free(cContentType)
	cContents := C.CString(contents)
	defer free(cContents)

	r := C.pactffi_interaction_contents(i.handle, C.int(part), cContentType, cContents)

	// 1 - A general panic was caught.
	// 2 - The mock server has already been started.
	// 3 - The interaction handle is invalid.
	// 4 - The content type is not valid.
	// 5 - The contents JSON is not valid JSON.
	// 6 - The plugin returned an error.
	res := int(r)
	switch res {
	case 1:
		return ErrPluginGenericPanic
	case 2:
		return ErrPluginMockServerStarted
	case 3:
		return ErrPluginInteractionHandleInvalid
	case 4:
		return ErrPluginInvalidContentType
	case 5:
		return ErrPluginInvalidJson
	case 6:
		return ErrPluginSpecificError
	default:
		if res != 0 {
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code:", res)
		}
	}

	return nil
}

func (i *Interaction) UponReceiving(description string) *Interaction {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.pactffi_upon_receiving(i.handle, cDescription)

	return i
}

func (i *Interaction) Given(state string) *Interaction {
	cState := C.CString(state)
	defer free(cState)

	C.pactffi_given(i.handle, cState)

	return i
}

func (i *Interaction) GivenWithParameter(state string, params map[string]interface{}) *Interaction {
	cState := C.CString(state)
	defer free(cState)

	for k, v := range params {
		cKey := C.CString(k)
		param := stringFromInterface(v)
		cValue := C.CString(param)

		C.pactffi_given_with_param(i.handle, cState, cKey, cValue)

		free(cValue)
		free(cKey)
	}

	return i
}

func (i *Interaction) WithRequest(method string, pathOrMatcher interface{}) *Interaction {
	cMethod := C.CString(method)
	defer free(cMethod)

	path := stringFromInterface(pathOrMatcher)
	cPath := C.CString(path)
	defer free(cPath)

	C.pactffi_with_request(i.handle, cMethod, cPath)

	return i
}

func (i *Interaction) WithRequestHeaders(valueOrMatcher map[string][]interface{}) *Interaction {
	return i.withHeaders(INTERACTION_PART_REQUEST, valueOrMatcher)
}

func (i *Interaction) WithResponseHeaders(valueOrMatcher map[string][]interface{}) *Interaction {
	return i.withHeaders(INTERACTION_PART_RESPONSE, valueOrMatcher)
}

func (i *Interaction) withHeaders(part interactionPart, valueOrMatcher map[string][]interface{}) *Interaction {
	for k, v := range valueOrMatcher {
		cName := C.CString(k)

		for _, header := range v {
			value := stringFromInterface(header)
			cValue := C.CString(value)

			C.pactffi_with_header_v2(i.handle, C.int(part), cName, CUlong(0), cValue)

			free(cValue)
		}

		free(cName)
	}

	return i
}

func (i *Interaction) WithQuery(valueOrMatcher map[string][]interface{}) *Interaction {
	for k, values := range valueOrMatcher {
		cName := C.CString(k)

		for idx, v := range values {
			value := stringFromInterface(v)
			cValue := C.CString(value)

			C.pactffi_with_query_parameter_v2(i.handle, cName, CUlong(idx), cValue)

			free(cValue)
		}

		free(cName)
	}

	return i
}

func (i *Interaction) WithJSONRequestBody(body interface{}) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_REQUEST)
}

func (i *Interaction) WithJSONResponseBody(body interface{}) *Interaction {
	return i.withJSONBody(body, INTERACTION_PART_RESPONSE)
}

func (i *Interaction) withJSONBody(body interface{}, part interactionPart) *Interaction {
	cHeader := C.CString("application/json")
	defer free(cHeader)

	jsonBody := stringFromInterface(body)
	cBody := C.CString(jsonBody)
	defer free(cBody)

	C.pactffi_with_body(i.handle, C.int(part), cHeader, cBody)

	return i
}

func (i *Interaction) WithRequestBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 0)
}

func (i *Interaction) WithResponseBody(contentType string, body []byte) *Interaction {
	return i.withBody(contentType, body, 1)
}

func (i *Interaction) withBody(contentType string, body []byte, part interactionPart) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cBody := C.CString(string(body))
	defer free(cBody)

	C.pactffi_with_body(i.handle, C.int(part), cHeader, cBody)

	return i
}

func (i *Interaction) withBinaryBody(contentType string, body []byte, part interactionPart) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	C.pactffi_with_binary_file(i.handle, C.int(part), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

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

func (i *Interaction) withMultipartFile(contentType string, filename string, mimePartName string, part interactionPart) *Interaction {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cPartName := C.CString(mimePartName)
	defer free(cPartName)

	cFilename := C.CString(filename)
	defer free(cFilename)

	C.pactffi_with_multipart_file(i.handle, C.int(part), cHeader, cFilename, cPartName)

	return i
}

// Set the expected HTTTP response status
func (i *Interaction) WithStatus(status int) *Interaction {
	C.pactffi_response_status(i.handle, C.ushort(status))

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
		return quotedString(string(bytes))
	}
}

// This fixes a quirk where certain "matchers" (e.g. matchers.S/String) are
// really just strings. However, whene we JSON encode them they get wrapped in quotes
// and the rust core sees them as plain strings, requiring then the quotes to be matched
func quotedString(s string) string {
	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// Experimental logging options
// func LogMessage(pkg, level, message string) {
// 	cPkg := C.CString(pkg)
// 	defer free(cPkg)

// 	cLevel := C.CString(level)
// 	defer free(cLevel)

// 	cMessage := C.CString(message)
// 	defer free(cMessage)

// 	res := C.pactffi_log_message(cPkg, cLevel, cMessage)
// 	log.Println("[DEBUG] log_to_buffer res", res)
// }

// func logToBuffer(level logLevel) error {
// 	res := C.pactffi_log_to_buffer(C.int(level))
// 	log.Println("[DEBUG] log_to_buffer res", res)

// 	return logResultToError(int(res))
// }

func logToStdout(level logLevel) error {
	res := C.pactffi_log_to_stdout(C.int(level))
	log.Println("[DEBUG] log_to_stdout res", res)

	return logResultToError(int(res))
}

func logToFile(file string, level logLevel) error {
	cFile := C.CString(file)
	defer free(cFile)

	res := C.pactffi_log_to_file(cFile, C.int(level))
	log.Println("[DEBUG] log_to_file res", res)

	return logResultToError(int(res))
}

// func getLogBuffer() string {
// 	buf := C.pactffi_fetch_log_buffer()
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

	// ErrPluginFailed indicates the plugin could not be started
	ErrPluginFailed = fmt.Errorf("the plugin could not be started")
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
