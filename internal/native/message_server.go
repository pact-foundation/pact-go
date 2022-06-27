package native

/*
// Library headers
#include <stdlib.h>
#include <stdint.h>
typedef int bool;
#define true 1
#define false 0

/// Wraps a Pact model struct
typedef struct InteractionHandle InteractionHandle;

struct InteractionHandle {
	unsigned int interaction_ref;
};

/// Wraps a Pact model struct
typedef struct PactHandle PactHandle;
struct PactHandle {
	unsigned int pact_ref;
};

PactHandle pactffi_new_message_pact(const char *consumer_name, const char *provider_name);
InteractionHandle pactffi_new_message(PactHandle pact, const char *description);
// Creates a new synchronous message interaction (request/response) and return a handle to it
InteractionHandle pactffi_new_sync_message_interaction(PactHandle pact, const char *description);
// Creates a new asynchronous message interaction (request/response) and return a handle to it
InteractionHandle pactffi_new_message_interaction(PactHandle pact, const char *description);
void pactffi_message_expects_to_receive(InteractionHandle message, const char *description);
void pactffi_message_given(InteractionHandle message, const char *description);
void pactffi_message_given_with_param(InteractionHandle message, const char *description, const char *name, const char *value);
void pactffi_message_with_contents(InteractionHandle message, const char *content_type, const char *body, int size);
void pactffi_message_with_metadata(InteractionHandle message, const char *key, const char *value);
char* pactffi_message_reify(InteractionHandle message);
int pactffi_write_message_pact_file(PactHandle pact, const char *directory, bool overwrite);
void pactffi_with_message_pact_metadata(PactHandle pact, const char *namespace, const char *name, const char *value);
int pactffi_write_pact_file(int mock_server_port, const char *directory, bool overwrite);

int pactffi_using_plugin(PactHandle pact, const char *plugin_name, const char *plugin_version);
void pactffi_cleanup_plugins(PactHandle pact);
int pactffi_interaction_contents(InteractionHandle interaction, int interaction_part, const char *content_type, const char *contents);

// Create a mock server for the provided Pact handle and transport.
int pactffi_create_mock_server_for_transport(PactHandle pact, const char *addr, int port, const char *transport, const char *transport_config);
bool pactffi_cleanup_mock_server(int mock_server_port);
char* pactffi_mock_server_mismatches(int mock_server_port);
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"unsafe"
)

type MessagePact struct {
	handle C.PactHandle
}

type Message struct {
	handle C.InteractionHandle
}

// MessageServer is the public interface for managing the message based interface
type MessageServer struct {
	messagePact *MessagePact
	messages    []*Message
}

// NewMessage initialises a new message for the current contract
func NewMessageServer(consumer string, provider string) *MessageServer {
	cConsumer := C.CString(consumer)
	cProvider := C.CString(provider)
	defer free(cConsumer)
	defer free(cProvider)

	return &MessageServer{messagePact: &MessagePact{handle: C.pactffi_new_message_pact(cConsumer, cProvider)}}
}

// Sets the additional metadata on the Pact file. Common uses are to add the client library details such as the name and version
func (m *MessageServer) WithMetadata(namespace, k, v string) *MessageServer {
	cNamespace := C.CString(namespace)
	defer free(cNamespace)
	cName := C.CString(k)
	defer free(cName)
	cValue := C.CString(v)
	defer free(cValue)

	C.pactffi_with_message_pact_metadata(m.messagePact.handle, cNamespace, cName, cValue)

	return m
}

// NewMessage initialises a new message for the current contract
// Deprecated: use NewAsyncMessageInteraction instead
func (m *MessageServer) NewMessage() *Message {
	// Alias
	return m.NewAsyncMessageInteraction("")
}

// NewSyncMessageInteraction initialises a new synchronous message interaction for the current contract
func (m *MessageServer) NewSyncMessageInteraction(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	i := &Message{
		handle: C.pactffi_new_sync_message_interaction(m.messagePact.handle, cDescription),
	}
	m.messages = append(m.messages, i)

	return i
}

// NewAsyncMessageInteraction initialises a new asynchronous message interaction for the current contract
func (m *MessageServer) NewAsyncMessageInteraction(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	i := &Message{
		handle: C.pactffi_new_message_interaction(m.messagePact.handle, cDescription),
	}
	m.messages = append(m.messages, i)

	return i
}

func (i *Message) Given(state string) *Message {
	cState := C.CString(state)
	defer free(cState)

	C.pactffi_message_given(i.handle, cState)

	return i
}

func (i *Message) GivenWithParameter(state string, params map[string]interface{}) *Message {
	cState := C.CString(state)
	defer free(cState)

	if len(params) == 0 {
		cState := C.CString(state)
		defer free(cState)

		C.pactffi_message_given(i.handle, cState)
	} else {
		for k, v := range params {
			cKey := C.CString(k)
			defer free(cKey)
			param := stringFromInterface(v)
			cValue := C.CString(param)
			defer free(cValue)

			C.pactffi_message_given_with_param(i.handle, cState, cKey, cValue)

		}
	}

	return i
}

func (i *Message) ExpectsToReceive(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.pactffi_message_expects_to_receive(i.handle, cDescription)

	return i
}

func (i *Message) WithMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {

		cName := C.CString(k)
		defer free(cName)

		// TODO: check if matching rules allowed here
		// value := stringFromInterface(v)
		// fmt.Printf("withheaders, sending: %+v \n\n", value)
		// cValue := C.CString(value)
		cValue := C.CString(v)
		defer free(cValue)

		C.pactffi_message_with_metadata(i.handle, cName, cValue)
	}

	return i
}

func (i *Message) WithBinaryContents(body []byte) *Message {
	return i.WithContents("application/octet-stream", body)
}

func (i *Message) WithJSONContents(body interface{}) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return i.WithContents("application/json", []byte(value))
}

// TODO: note that string values here must be NUL terminated.
func (i *Message) WithContents(contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	C.pactffi_message_with_contents(i.handle, cHeader, (*C.char)(unsafe.Pointer(&body[0])), C.int(len(body)))

	return i
}

// TODO: migrate plugin code to shared struct/code?

// NewInteraction initialises a new interaction for the current contract
func (m *MessageServer) UsingPlugin(pluginName string, pluginVersion string) error {
	cPluginName := C.CString(pluginName)
	defer free(cPluginName)
	cPluginVersion := C.CString(pluginVersion)
	defer free(cPluginVersion)

	r := C.pactffi_using_plugin(m.messagePact.handle, cPluginName, cPluginVersion)

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
func (i *Message) WithPluginInteractionContents(interactionPart interactionType, contentType string, contents string) error {
	cContentType := C.CString(contentType)
	defer free(cContentType)
	cContents := C.CString(contents)
	defer free(cContents)

	r := C.pactffi_interaction_contents(i.handle, C.int(interactionPart), cContentType, cContents)

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

// StartTransport starts up a mock server on the given address:port for the given transport
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MessageServer) StartTransport(transport string, address string, port int, config map[string][]interface{}) (int, error) {
	if len(m.messages) == 0 {
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

	p := C.pactffi_create_mock_server_for_transport(m.messagePact.handle, cAddress, C.int(port), cTransport, cConfig)

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

// NewInteraction initialises a new interaction for the current contract
func (m *MessageServer) CleanupPlugins(pluginName string, pluginVersion string) {
	C.pactffi_cleanup_plugins(m.messagePact.handle)
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MessageServer) CleanupMockServer(port int) bool {
	if len(m.messages) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := C.pactffi_cleanup_mock_server(C.int(port))

	return int(res) == 1
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MessageServer) MockServerMismatchedRequests(port int) []MismatchedRequest {
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

	json.Unmarshal([]byte(C.GoString(mismatches)), &res)

	return res
}

func (i *Message) ReifyMessage() string {
	return C.GoString(C.pactffi_message_reify(i.handle))
}

// WritePactFile writes the Pact to file.
func (m *MessageServer) WritePactFile(dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	overwritePact := 0
	if overwrite {
		overwritePact = 1
	}

	res := int(C.pactffi_write_message_pact_file(m.messagePact.handle, cDir, C.int(overwritePact)))

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The pact file was not able to be written |
	/// | 2 | The message pact for the given handle was not found |
	switch res {
	case 0:
		return nil
	case 1:
		return ErrUnableToWritePactFile
	case 2:
		return ErrHandleNotFound
	default:
		return fmt.Errorf("an unknown error ocurred when writing to pact file")
	}
}

// WritePactFile writes the Pact to file.
func (m *MessageServer) WritePactFileForServer(port int, dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	overwritePact := 0
	if overwrite {
		overwritePact = 1
	}

	res := int(C.pactffi_write_pact_file(C.int(port), cDir, C.int(overwritePact)))

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The pact file was not able to be written |
	/// | 2 | The message pact for the given handle was not found |
	switch res {
	case 0:
		return nil
	case 1:
		return ErrMockServerPanic
	case 2:
		return ErrUnableToWritePactFile
	case 3:
		return ErrHandleNotFound
	default:
		return fmt.Errorf("an unknown error ocurred when writing to pact file")
	}
}
