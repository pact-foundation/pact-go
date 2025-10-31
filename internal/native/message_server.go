package native

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"unsafe"
)

type MessagePact struct {
	handle uintptr
}

type messageType int

const (
	MESSAGE_TYPE_ASYNC messageType = iota
	MESSAGE_TYPE_SYNC
)

type Message struct {
	handle      uintptr
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
	// cConsumer := C.CString(consumer)
	// cProvider := C.CString(provider)
	// defer free(cConsumer)
	// defer free(cProvider)

	return &MessageServer{messagePact: &MessagePact{handle: pactffi_new_message_pact(consumer, provider)}}
}

// Sets the additional metadata on the Pact file. Common uses are to add the client library details such as the name and version
func (m *MessageServer) WithMetadata(namespace, k, v string) *MessageServer {
	// cNamespace := C.CString(namespace)
	// defer free(cNamespace)
	// cName := C.CString(k)
	// defer free(cName)
	// cValue := C.CString(v)
	// defer free(cValue)

	pactffi_with_message_pact_metadata(m.messagePact.handle, namespace, k, v)

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

	i := &Message{
		handle:      pactffi_new_sync_message_interaction(m.messagePact.handle, description),
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

	i := &Message{
		handle:      pactffi_new_message_interaction(m.messagePact.handle, description),
		messageType: MESSAGE_TYPE_ASYNC,
		pact:        m.messagePact,
		index:       len(m.messages),
		server:      m,
	}
	m.messages = append(m.messages, i)

	return i
}

func (m *MessageServer) WithSpecificationVersion(version specificationVersion) {
	pactffi_with_specification(m.messagePact.handle, int32(version))
}

func (m *Message) Given(state string) *Message {
	pactffi_given(m.handle, state)

	return m
}

func (m *Message) GivenWithParameter(state string, params map[string]interface{}) *Message {
	if len(params) == 0 {
		pactffi_given(m.handle, state)
	} else {
		for k, v := range params {
			param := stringFromInterface(v)
			pactffi_given_with_param(m.handle, state, k, param)
		}
	}

	return m
}

func (m *Message) ExpectsToReceive(description string) *Message {
	pactffi_message_expects_to_receive(m.handle, description)

	return m
}

func (m *Message) WithMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		// TODO: check if matching rules allowed here
		// value := stringFromInterface(v)
		// fmt.Printf("withheaders, sending: %+v \n\n", value)
		// cValue := C.CString(value)

		pactffi_message_with_metadata(m.handle, k, v)
	}

	return m
}
func (m *Message) WithRequestMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		pactffi_with_metadata(m.handle, k, v, 0)
	}

	return m
}
func (m *Message) WithResponseMetadata(valueOrMatcher map[string]string) *Message {
	for k, v := range valueOrMatcher {
		pactffi_with_metadata(m.handle, k, v, 1)
	}

	return m
}

func (m *Message) WithRequestBinaryContents(body []byte) *Message {

	// TODO: handle response
	res := pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_REQUEST), "application/octet-stream", string(body), size_t(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}
func (m *Message) WithRequestBinaryContentType(contentType string, body []byte) *Message {

	// TODO: handle response
	res := pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_REQUEST), contentType, string(body), size_t(len(body)))

	log.Println("[DEBUG] WithRequestBinaryContents - pactffi_with_binary_file returned", res)

	return m
}

func (m *Message) WithRequestJSONContents(body interface{}) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_REQUEST, "application/json", []byte(value))
}

func (m *Message) WithResponseBinaryContents(body []byte) *Message {

	// TODO: handle response
	pactffi_with_binary_file(m.handle, int32(INTERACTION_PART_RESPONSE), "application/octet-stream", string(body), size_t(len(body)))

	return m
}

func (m *Message) WithResponseJSONContents(body interface{}) *Message {
	value := stringFromInterface(body)

	log.Println("[DEBUG] message WithJSONContents", value)

	return m.WithContents(INTERACTION_PART_RESPONSE, "application/json", []byte(value))
}

// Note that string values here must be NUL terminated.
func (m *Message) WithContents(part interactionPart, contentType string, body []byte) *Message {

	res := pactffi_with_body(m.handle, int32(part), contentType, string(body))
	log.Println("[DEBUG] response from pactffi_interaction_contents", res)

	return m
}

// TODO: migrate plugin code to shared struct/code?

// NewInteraction initialises a new interaction for the current contract
func (m *MessageServer) UsingPlugin(pluginName string, pluginVersion string) error {

	r := pactffi_using_plugin(m.messagePact.handle, pluginName, pluginVersion)
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

	r := pactffi_interaction_contents(m.handle, int32(part), contentType, contents)

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

func GoByteArrayFromC(ptr uintptr, length int) []byte {
	if unsafe.Pointer(ptr) == nil || length == 0 {
		return []byte{}
	}
	// Create a Go byte slice from the C string
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + uintptr(i)))
	}
	return b
}

// GetMessageContents retreives the binary contents of the request for a given message
// any matchers are stripped away if given
// if the contents is from a plugin, the byte[] representation of the parsed
// plugin data is returned, again, with any matchers etc. removed
func (m *Message) GetMessageRequestContents() ([]byte, error) {
	log.Println("[DEBUG] GetMessageRequestContents")
	if m.messageType == MESSAGE_TYPE_ASYNC {
		iter := pactffi_pact_handle_get_message_iter(m.pact.handle)
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter")
		// TODO
		if unsafe.Pointer(iter) == nil {
			return nil, errors.New("unable to get a message iterator")
		}
		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - OK")

		///////
		// TODO: some debugging in here to see what's exploding.......
		///////

		log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - len", len(m.server.messages))

		for i := 0; i < len(m.server.messages); i++ {
			log.Println("[DEBUG] pactffi_pact_handle_get_message_iter - index", i)
			message := pactffi_pact_message_iter_next(iter)
			log.Println("[DEBUG] pactffi_pact_message_iter_next - message", message)

			if i == m.index {
				log.Println("[DEBUG] pactffi_pact_message_iter_next - index match", message)

				if unsafe.Pointer(message) == nil {
					return nil, errors.New("retrieved a null message pointer")
				}
				len := pactffi_message_get_contents_length(message)
				log.Println("[DEBUG] pactffi_message_get_contents_length - len", len)
				if len == 0 {
					// You can have empty bodies
					log.Println("[DEBUG] message body is empty")
					return []byte{}, nil
				}
				data := pactffi_message_get_contents_bin(message)
				log.Println("[DEBUG] pactffi_message_get_contents_bin - data", data)
				if unsafe.Pointer(data) == nil {
					// You can have empty bodies
					log.Println("[DEBUG] message binary contents are empty")
					return nil, nil
				}
				bytes := GoByteArrayFromC(data, int(len))
				return bytes, nil
			}
		}

	} else {
		iter := pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
		if unsafe.Pointer(iter) == nil {
			return nil, errors.New("unable to get a message iterator")
		}

		for i := 0; i < len(m.server.messages); i++ {
			message := pactffi_pact_sync_message_iter_next(iter)

			if i == m.index {
				if unsafe.Pointer(message) == nil {
					return nil, errors.New("retrieved a null message pointer")
				}

				len := pactffi_sync_message_get_request_contents_length(message)
				if len == 0 {
					log.Println("[DEBUG] message body is empty")
					return []byte{}, nil
				}
				data := pactffi_sync_message_get_request_contents_bin(message)
				if unsafe.Pointer(data) == nil {
					log.Println("[DEBUG] message binary contents are empty")
					return nil, nil
				}
				bytes := GoByteArrayFromC(data, int(len))

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
	iter := pactffi_pact_handle_get_sync_message_iter(m.pact.handle)
	if unsafe.Pointer(iter) == nil {
		return nil, errors.New("unable to get a message iterator")
	}

	for i := 0; i < len(m.server.messages); i++ {
		message := pactffi_pact_sync_message_iter_next(iter)

		if unsafe.Pointer(message) == nil {
			return nil, errors.New("retrieved a null message pointer")
		}

		// Get Response body
		len := pactffi_sync_message_get_response_contents_length(message, size_t(i))
		// if len == 0 {
		// 	return nil, errors.New("retrieved an empty message")
		// }
		if len == 0 {
			// You can have empty bodies
			log.Println("[DEBUG] message body is empty")
			responses[i] = []byte{}
			return responses, nil
		}
		data := pactffi_sync_message_get_response_contents_bin(message, size_t(i))
		if unsafe.Pointer(data) == nil {
			return nil, errors.New("retrieved an empty pointer to the message contents")
		}
		bytes := GoByteArrayFromC(data, int(len))
		responses[i] = bytes
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
	configJson := stringFromInterface(config)

	p := pactffi_create_mock_server_for_transport(m.messagePact.handle, address, uint16(port), transport, configJson)

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
	pactffi_cleanup_plugins(m.messagePact.handle)
}

// CleanupMockServer frees the memory from the previous mock server.
func (m *MessageServer) CleanupMockServer(port int) bool {
	if len(m.messages) == 0 {
		return true
	}
	log.Println("[DEBUG] mock server cleaning up port:", port)
	res := pactffi_cleanup_mock_server(int32(port))

	return res
}

// MockServerMismatchedRequests returns a JSON object containing any mismatches from
// the last set of interactions.
func (m *MessageServer) MockServerMismatchedRequests(port int) []MismatchedRequest {
	log.Println("[DEBUG] mock server determining mismatches:", port)
	var res []MismatchedRequest

	mismatches := pactffi_mock_server_mismatches(int32(port))
	// This method can return a nil pointer, in which case, it
	// should be considered a failure (or at least, an issue)
	// converting it to a string might also do nasty things here!
	// TODO change return type to uintptr
	if mismatches == "" {
		log.Println("[WARN] received a null pointer from the native interface, returning empty list of mismatches")
		return []MismatchedRequest{}
	}

	err := json.Unmarshal([]byte(mismatches), &res)
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

	res := pactffi_mock_server_matched(int32(port))

	// TODO: why this number is so big and not a bool? Type def wrong? Port value wrong?
	// log.Println("MATCHED RES?")
	// log.Println(int(res))

	return res
}

// WritePactFile writes the Pact to file.
func (m *MessageServer) WritePactFile(dir string, overwrite bool) error {
	log.Println("[DEBUG] writing pact file for message pact at dir:", dir)

	overwritePact := false
	if overwrite {
		overwritePact = true
	}

	res := pactffi_write_message_pact_file(m.messagePact.handle, dir, overwritePact)

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

	overwritePact := false
	if overwrite {
		overwritePact = true
	}

	res := pactffi_write_pact_file(int32(port), dir, overwritePact)

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
