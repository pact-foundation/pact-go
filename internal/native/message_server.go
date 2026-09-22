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

// MessagePact is a Go representation of the message-based PactHandle struct.
type MessagePact struct {
	handle C.PactHandle
}

type messageType int

// Whether a Message was created as an asynchronous or synchronous
// (request/response) message interaction.
const (
	MESSAGE_TYPE_ASYNC messageType = iota
	MESSAGE_TYPE_SYNC
)

// Message is a single message interaction on a MessageServer, either
// asynchronous (one body) or synchronous (a request/response pair).
type Message struct {
	handle      C.InteractionHandle
	messageType messageType
	pact        *MessagePact
	index       int
	server      *MessageServer
}

// MessageServer is the public interface for managing the message based interface.
type MessageServer struct {
	messagePact *MessagePact
	messages    []*Message
}

// NewMessageServer creates a new message-based Pact for a given
// consumer/provider.
func NewMessageServer(consumer string, provider string) *MessageServer {
	cConsumer := C.CString(consumer)
	cProvider := C.CString(provider)
	defer free(cConsumer)
	defer free(cProvider)

	return &MessageServer{messagePact: &MessagePact{handle: C.pactffi_new_message_pact(cConsumer, cProvider)}}
}

// WithMetadata sets additional metadata on the Pact file, grouped under
// namespace. Common uses are to add client library details such as the
// name and version.
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

// NewMessage initialises a new message for the current contract.
//
// Deprecated: use NewAsyncMessageInteraction instead.
func (m *MessageServer) NewMessage() *Message {
	// Alias
	return m.NewAsyncMessageInteraction("")
}

// NewSyncMessageInteraction initialises a new synchronous message interaction for the current contract.
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

// NewAsyncMessageInteraction initialises a new asynchronous message interaction for the current contract.
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

// WithSpecificationVersion sets the Pact specification version this
// message pact is written and verified against.
func (m *MessageServer) WithSpecificationVersion(version specificationVersion) {
	C.pactffi_with_specification(m.messagePact.handle, C.int(version))
}

// Given adds a provider state that must hold for this message.
func (m *Message) Given(state string) *Message {
	interactionGiven(m.handle, state)

	return m
}

// GivenWithParameter adds a provider state, parameterised by params, that
// must hold for this message.
func (m *Message) GivenWithParameter(state string, params map[string]any) *Message {
	if len(params) == 0 {
		interactionGiven(m.handle, state)
	} else {
		interactionGivenWithParams(m.handle, state, params)
	}

	return m
}

// ExpectsToReceive sets the description of this message interaction.
func (m *Message) ExpectsToReceive(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.pactffi_message_expects_to_receive(m.handle, cDescription)

	return m
}

// WithMetadata sets metadata key/value pairs to expect alongside this
// message's contents.
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

// WithRequestMetadata sets metadata key/value pairs to expect on the
// request half of this synchronous message.
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

// WithResponseMetadata sets metadata key/value pairs to expect on the
// response half of this synchronous message.
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

// WithRequestBinaryContents sets the request contents of this message to
// body, with an application/octet-stream content type.
func (m *Message) WithRequestBinaryContents(body []byte) *Message {
	cHeader := C.CString("application/octet-stream")
	defer free(cHeader)

	// TODO: handle response
	res := C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_REQUEST), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", bool(res))

	return m
}

// WithRequestBinaryContentType sets the request contents of this message
// to body, with the given content type.
func (m *Message) WithRequestBinaryContentType(contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	// TODO: handle response
	res := C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_REQUEST), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}

// WithRequestJSONContents JSON-encodes body (which may contain matchers)
// as the request contents of this message.
func (m *Message) WithRequestJSONContents(body any) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_REQUEST, "application/json", []byte(value))
}

// WithResponseBinaryContents sets the response contents of this
// synchronous message to body, with an application/octet-stream content
// type.
func (m *Message) WithResponseBinaryContents(body []byte) *Message {
	cHeader := C.CString("application/octet-stream")
	defer free(cHeader)

	// TODO: handle response
	C.pactffi_with_binary_file(m.handle, C.int(INTERACTION_PART_RESPONSE), cHeader, (*C.uchar)(unsafe.Pointer(&body[0])), CUlong(len(body)))

	return m
}

// WithResponseJSONContents JSON-encodes body (which may contain matchers)
// as the response contents of this synchronous message.
func (m *Message) WithResponseJSONContents(body any) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_RESPONSE, "application/json", []byte(value))
}

// WithContents sets the request or response contents of this message to
// body with the given content type. Note that string values here must be
// NUL terminated.
func (m *Message) WithContents(part interactionPart, contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cBody := C.CString(string(body))
	defer free(cBody)

	res := C.pactffi_with_body(m.handle, C.int(part), cHeader, cBody)
	log.Println("[DEBUG] response from pactffi_interaction_contents", bool(res))

	return m
}

// TODO: migrate plugin code to shared struct/code?

// UsingPlugin loads a pact_ffi plugin by name and version so subsequent
// messages on this pact can use plugin-provided matchers and content
// types (e.g. protobuf, gRPC).
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
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code", res)
		}
	}

	return nil
}

// WithPluginInteractionContents sets the request or response contents of
// this message from a plugin-provided contentType (e.g. a protobuf
// message type), delegating the encoding to the loaded plugin.
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
		return ErrPluginInvalidJSON
	case 6:
		return ErrPluginSpecificError
	default:
		if res != 0 {
			return fmt.Errorf("an unknown error (code: %v) occurred when adding a plugin for the test. Received error code", res)
		}
	}

	return nil
}

// GetMessageRequestContents retrieves the binary contents of the request
// for a given message; any matchers are stripped away if given. If the
// contents are from a plugin, the byte[] representation of the parsed
// plugin data is returned, again with any matchers etc. removed.
func (m *Message) GetMessageRequestContents() ([]byte, error) {
	log.Println("[DEBUG] GetMessageRequestContents")
	if m.messageType == MESSAGE_TYPE_ASYNC {
		return m.getAsyncMessageRequestContents()
	}
	return m.getSyncMessageRequestContents()
}

// GetMessageResponseContents retreives the binary contents of the response for a given message
// any matchers are stripped away if given
// if the contents is from a plugin, the byte[] representation of the parsed
// plugin data is returned, again, with any matchers etc. removed.
func (m *Message) GetMessageResponseContents() ([][]byte, error) {
	responses := make([][]byte, len(m.server.messages))
	if m.messageType == MESSAGE_TYPE_ASYNC {
		return nil, errors.New("invalid request: asynchronous messages do not have response")
	}
	iter := C.pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}

	for i := range len(m.server.messages) {
		message := C.pactffi_pact_sync_message_iter_next(iter)

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		// Get Response body
		contentsLen := C.pactffi_sync_message_get_response_contents_length(message, C.size_t(i))
		if contentsLen != 0 {
			data := C.pactffi_sync_message_get_response_contents_bin(message, C.size_t(i))
			if data == nil {
				return nil, errors.New("retrieved an empty pointer to the message contents")
			}
			ptr := unsafe.Pointer(data)
			bytes := C.GoBytes(ptr, C.int(contentsLen))
			responses[i] = bytes
		}
	}

	return responses, nil
}

// StartTransport starts up a mock server on the given address:port for the given transport
// https://docs.rs/pact_ffi/latest/pact_ffi/mock_server/fn.pactffi_create_mock_server_for_transport.html
func (m *MessageServer) StartTransport(transport string, address string, port int, config map[string][]any) (int, error) {
	if len(m.messages) == 0 {
		return 0, ErrNoInteractions
	}

	log.Println("[DEBUG] mock server starting on address:", address, port)
	cAddress := C.CString(address)
	defer free(cAddress)

	cTransport := C.CString(transport)
	defer free(cTransport)

	configJSON := stringFromInterface(config)
	cConfig := C.CString(configJSON)
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
			waitForTransport(address, msPort)
			return msPort, nil
		}
		return msPort, fmt.Errorf("an unknown error (code: %v) occurred when starting a mock server for the test", msPort)
	}
}

// CleanupPlugins releases the plugins loaded on this pact via UsingPlugin.
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

// MockServerMatched reports whether every interaction registered against
// the mock server on port was matched by an actual request.
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

	overwritePact := overwrite

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
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// WritePactFileForServer writes the Pact for the mock server running on
// port to dir, optionally overwriting an existing file.
func (m *MessageServer) WritePactFileForServer(port int, dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)
	cDir := C.CString(dir)
	defer free(cDir)

	overwritePact := overwrite

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
		return errors.New("an unknown error occurred when writing to pact file")
	}
}

// WithReference records an external reference (e.g. a ticket or pull request)
// against the interaction. References are stored under comments.references[group][name]
// in the Pact file. This is a V4-only feature.
func (m *Message) WithReference(group, name, value string) *Message {
	cGroup := C.CString(group)
	defer free(cGroup)
	cName := C.CString(name)
	defer free(cName)
	cValue := C.CString(value)
	defer free(cValue)

	C.pactffi_add_interaction_reference(m.handle, cGroup, cName, cValue)

	return m
}

// getAsyncMessageRequestContents is the MESSAGE_TYPE_ASYNC branch of
// GetMessageRequestContents, split out to keep both branches readable.
func (m *Message) getAsyncMessageRequestContents() ([]byte, error) {
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

	for i := range len(m.server.messages) {
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - index", i)
		message := C.pactffi_pact_message_iter_next(iter)
		log.Println("[DEBUG] pactffi_pact_message_iter_next - message", message)

		if i != m.index {
			continue
		}
		log.Println("[DEBUG] pactffi_pact_message_iter_next - index match", message)

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		contentsLen := C.pactffi_message_get_contents_length(message)
		log.Println("[DEBUG] pactffi_message_get_contents_length - len", contentsLen)
		if contentsLen == 0 {
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
		bytes := C.GoBytes(ptr, C.int(contentsLen))

		return bytes, nil
	}

	return nil, errors.New("unable to find the message")
}

// getSyncMessageRequestContents is the synchronous-message branch of
// GetMessageRequestContents, split out to keep both branches readable.
func (m *Message) getSyncMessageRequestContents() ([]byte, error) {
	iter := C.pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if iter == nil {
		return nil, errors.New("unable to get a message iterator")
	}

	for i := range len(m.server.messages) {
		message := C.pactffi_pact_sync_message_iter_next(iter)

		if i != m.index {
			continue
		}

		if message == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		contentsLen := C.pactffi_sync_message_get_request_contents_length(message)
		if contentsLen == 0 {
			log.Println("[DEBUG] message body is empty")
			return nil, nil
		}
		data := C.pactffi_sync_message_get_request_contents_bin(message)
		if data == nil {
			log.Println("[DEBUG] message binary contents are empty")
			return nil, nil
		}
		ptr := unsafe.Pointer(data)
		bytes := C.GoBytes(ptr, C.int(contentsLen))

		return bytes, nil
	}

	return nil, errors.New("unable to find the message")
}
