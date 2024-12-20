package native

/*
#include "pact.h"
*/
import "C"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"unsafe"
)

type MessagePact struct {
	handle C.PactHandle
}

type messageType int

const (
	MESSAGE_TYPE_ASYNC messageType = iota
	MESSAGE_TYPE_SYNC
)

type Message struct {
	handle      C.InteractionHandle
	messageType messageType
	pact        *MessagePact
	index       int
	server      *MessageServer
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
		handle:      C.pactffi_new_sync_message_interaction(m.messagePact.handle, cDescription),
		messageType: MESSAGE_TYPE_SYNC,
		pact:        m.messagePact,
		index:       len(m.messages),
		server:      m,
	}
	m.messages = append(m.messages, i)

	return i
}

// NewAsyncMessageInteraction initialises a new asynchronous message interaction for the current contract
func (m *MessageServer) NewAsyncMessageInteraction(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	i := &Message{
		handle:      C.pactffi_new_message_interaction(m.messagePact.handle, cDescription),
		messageType: MESSAGE_TYPE_ASYNC,
		pact:        m.messagePact,
		index:       len(m.messages),
		server:      m,
	}
	m.messages = append(m.messages, i)

	return i
}

func (m *MessageServer) WithSpecificationVersion(version specificationVersion) {
	C.pactffi_with_specification(m.messagePact.handle, C.int(version))
}

func (m *Message) Given(state string) *Message {
	cState := C.CString(state)
	defer free(cState)

	C.pactffi_given(m.handle, cState)

	return m
}

func (m *Message) GivenWithParameter(state string, params map[string]interface{}) *Message {
	cState := C.CString(state)
	defer free(cState)

	if len(params) == 0 {
		C.pactffi_given(m.handle, cState)
	} else {
		for k, v := range params {
			cKey := C.CString(k)

			param := stringFromInterface(v)
			cValue := C.CString(param)

			C.pactffi_given_with_param(m.handle, cState, cKey, cValue)

			free(cValue)
			free(cKey)
		}
	}

	return m
}

func (m *Message) ExpectsToReceive(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.pactffi_message_expects_to_receive(m.handle, cDescription)

	return m
}

func (m *Message) WithMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		cName := C.CString(k)

		// TODO: check if matching rules allowed here
		// value := stringFromInterface(v)
		// fmt.Printf("withheaders, sending: %+v \n\n", value)
		// cValue := C.CString(value)
		cValue := C.CString(v)

		C.pactffi_message_with_metadata(m.handle, cName, cValue)

		free(cValue)
		free(cName)
	}

	return m
}
func (m *Message) WithRequestMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		cName := C.CString(k)
		cValue := C.CString(v)

		C.pactffi_with_metadata(m.handle, cName, cValue, 0)

		free(cValue)
		free(cName)
	}

	return m
}
func (m *Message) WithResponseMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		cName := C.CString(k)
		cValue := C.CString(v)

		C.pactffi_with_metadata(m.handle, cName, cValue, 1)

		free(cValue)
		free(cName)
	}

	return m
}

func (m *Message) WithRequestBinaryContents(body []byte) *Message {
	cHeader := C.CString("application/octet-stream")
	defer free(cHeader)

	// TODO: handle response
	res := C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_REQUEST), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", bool(res))

	return m
}
func (m *Message) WithRequestBinaryContentType(contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	// TODO: handle response
	res := C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_REQUEST), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}

func (m *Message) WithRequestJSONContents(body interface{}) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_REQUEST, "application/json", []byte(value))
}

func (m *Message) WithResponseBinaryContents(body []byte) *Message {
	cHeader := C.CString("application/octet-stream")
	defer free(cHeader)

	// TODO: handle response
	C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_RESPONSE), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	return m
}

func (m *Message) WithResponseJSONContents(body interface{}) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_RESPONSE, "application/json", []byte(value))
}

// Note that string values here must be NUL terminated.
func (m *Message) WithContents(part interactionPart, contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	res := C.pactffi_with_body(m.handle, C.int(part), cHeader, (*C.char)(unsafe.Pointer(&body[0])))
	log.Println("[DEBUG] response from pactffi_interaction_contents", (bool(res)))

	return m
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
func (m *Message) WithPluginInteractionContents(part interactionPart, contentType string, contents string) error {
	cContentType := C.CString(contentType)
	defer free(cContentType)
	cContents := C.CString(contents)
	defer free(cContents)

	r := C.pactffi_interaction_contents(m.handle, C.int(part), cContentType, cContents)

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

// GetMessageContents retreives the binary contents of the request for a given message
// any matchers are stripped away if given
// if the contents is from a plugin, the byte[] representation of the parsed
// plugin data is returned, again, with any matchers etc. removed
func (m *Message) GetMessageRequestContents() ([]byte, error) {
	log.Println("[DEBUG] GetMessageRequestContents")
	if m.messageType == MESSAGE_TYPE_ASYNC {
		iter := C.pactffi_pact_handle_get_message_iter(m.pact.handle)
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter")
		if iter == nil {
			return nil, errors.New("unable to get a message iterator")
		}
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - OK")

		///////
		// TODO: some debugging in here to see what's exploding.......
		///////

		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - len", len(m.server.messages))

		for i := 0; i < len(m.server.messages); i++ {
			log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - index", i)
			message := C.pactffi_pact_message_iter_next(iter)
			log.Println("[DEBUG] pactffi_pact_message_iter_next - message", message)

			if i == m.index {
				log.Println("[DEBUG] pactffi_pact_message_iter_next - index match", message)

				if message == nil {
					return nil, errors.New("retrieved a null message pointer")
				}

				len := C.pactffi_message_get_contents_length(message)
				log.Println("[DEBUG] pactffi_message_get_contents_length - len", len)
				if len == 0 {
					// You can have empty bodies
					log.Println("[DEBUG] message body is empty")
					return nil, nil
				}
				data := C.pactffi_message_get_contents_bin(message)
				log.Println("[DEBUG] pactffi_message_get_contents_bin - data", data)
				if data == nil {
					// You can have empty bodies
					log.Println("[DEBUG] message binary contents are empty")
					return nil, nil
				}
				ptr := unsafe.Pointer(data)
				bytes := C.GoBytes(ptr, C.int(len))

				return bytes, nil
			}
		}

	} else {
		iter := C.pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
		if iter == nil {
			return nil, errors.New("unable to get a message iterator")
		}

		for i := 0; i < len(m.server.messages); i++ {
			message := C.pactffi_pact_sync_message_iter_next(iter)

			if i == m.index {
				if message == nil {
					return nil, errors.New("retrieved a null message pointer")
				}

				len := C.pactffi_sync_message_get_request_contents_length(message)
				if len == 0 {
					log.Println("[DEBUG] message body is empty")
					return nil, nil
				}
				data := C.pactffi_sync_message_get_request_contents_bin(message)
				if data == nil {
					log.Println("[DEBUG] message binary contents are empty")
					return nil, nil
				}
				ptr := unsafe.Pointer(data)
				bytes := C.GoBytes(ptr, C.int(len))

				return bytes, nil
			}
		}
	}

	return nil, errors.New("unable to find the message")
}

// GetMessageResponseContents retreives the binary contents of the response for a given message
// any matchers are stripped away if given
// if the contents is from a plugin, the byte[] representation of the parsed
// plugin data is returned, again, with any matchers etc. removed
func (m *Message) GetMessageResponseContents() ([][]byte, error) {

	responses := make([][]byte, len(m.server.messages))
	if m.messageType == MESSAGE_TYPE_ASYNC {
		return nil, errors.New("invalid request: asynchronous messages do not have response")
	}
	iter := C.pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}

	for i := 0; i < len(m.server.messages); i++ {
		message := C.pactffi_pact_sync_message_iter_next(iter)

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		// Get Response body
		len := C.pactffi_sync_message_get_response_contents_length(message, C.size_t(i))
		if len != 0 {
			data := C.pactffi_sync_message_get_response_contents_bin(message, C.size_t(i))
			if data == nil {
				return nil, errors.New("retrieved an empty pointer to the message contents")
			}
			ptr := unsafe.Pointer(data)
			bytes := C.GoBytes(ptr, C.int(len))
			responses[i] = bytes
		}
	}

	return responses, nil
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

	p := C.pactffi_create_mock_server_for_transport(m.messagePact.handle, cAddress, C.ushort(port), cTransport, cConfig)

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
func (m *MessageServer) CleanupPlugins() {
	C.pactffi_cleanup_plugins(m.messagePact.handle)
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MessageServer) CleanupMockServer(port int) bool {
	if len(m.messages) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := C.pactffi_cleanup_mock_server(C.int(port))

	return bool(res)
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

	err := json.Unmarshal([]byte(C.GoString(mismatches)), &res)
	if err != nil {
		log.Println("[ERROR] failed to unmarshal mismatches response, returning empty list of mismatches")
		return []MismatchedRequest{}
	}

	return res
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MessageServer) MockServerMatched(port int) bool {
	log.Println("[DEBUG] mock server determining mismatches:", port)

	res := C.pactffi_mock_server_matched(C.int(port))

	// TODO: why this number is so big and not a bool? Type def wrong? Port value wrong?
	// log.Println("MATCHED RES?")
	// log.Println(int(res))

	return bool(res)
}

// WritePactFile writes the Pact to file.
func (m *MessageServer) WritePactFile(dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	overwritePact := false
	if overwrite {
		overwritePact = true
	}

	res := int(C.pactffi_write_message_pact_file(m.messagePact.handle, cDir, C.bool(overwritePact)))

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

	overwritePact := false
	if overwrite {
		overwritePact = true
	}

	res := int(C.pactffi_write_pact_file(C.int(port), cDir, C.bool(overwritePact)))

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
