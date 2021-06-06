package mockserver

/*
// Library headers
#include <stdlib.h>
#include <stdint.h>
typedef int bool;
#define true 1
#define false 0

typedef struct MessageHandle MessageHandle;
struct MessageHandle {
	uintptr_t pact;
  uintptr_t message;
};

/// Wraps a PactMessage model struct
typedef struct MessagePactHandle MessagePactHandle;
struct MessagePactHandle {
  uintptr_t pact;
};

MessagePactHandle new_message_pact(const char *consumer_name, const char *provider_name);
MessageHandle new_message(MessagePactHandle pact, const char *description);
void message_expects_to_receive(MessageHandle message, const char *description);
void message_given(MessageHandle message, const char *description);
void message_given_with_param(MessageHandle message, const char *description, const char *name, const char *value);
void message_with_contents(MessageHandle message, const char *content_type, const char *body, int size);
void message_with_metadata(MessageHandle message, const char *key, const char *value);
char* message_reify(MessageHandle message);
int write_message_pact_file(MessagePactHandle pact, const char *directory, bool overwrite);
void with_message_pact_metadata(MessagePactHandle pact, const char *namespace, const char *name, const char *value);
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

type MessagePact struct {
	handle C.MessagePactHandle
}

type Message struct {
	handle C.MessageHandle
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

	return &MessageServer{messagePact: &MessagePact{handle: C.new_message_pact(cConsumer, cProvider)}}
}

// Sets the additional metadata on the Pact file. Common uses are to add the client library details such as the name and version
func (m *MessageServer) WithMetadata(namespace, k, v string) *MessageServer {
	cNamespace := C.CString(namespace)
	defer free(cNamespace)
	cName := C.CString(k)
	defer free(cName)
	cValue := C.CString(v)
	defer free(cValue)

	C.with_message_pact_metadata(m.messagePact.handle, cNamespace, cName, cValue)

	return m
}

// NewMessage initialises a new message for the current contract
func (m *MessageServer) NewMessage() *Message {
	cDescription := C.CString("")
	defer free(cDescription)

	i := &Message{
		handle: C.new_message(m.messagePact.handle, cDescription),
	}
	m.messages = append(m.messages, i)

	return i
}

func (i *Message) Given(state string) *Message {
	cState := C.CString(state)
	defer free(cState)

	C.message_given(i.handle, cState)

	return i
}

func (i *Message) GivenWithParameter(state string, params map[string]interface{}) *Message {
	cState := C.CString(state)
	defer free(cState)

	for k, v := range params {
		cKey := C.CString(k)
		defer free(cKey)
		param := stringFromInterface(v)
		cValue := C.CString(param)
		defer free(cValue)

		C.message_given_with_param(i.handle, cState, cKey, cValue)

	}

	return i
}

func (i *Message) ExpectsToReceive(description string) *Message {
	cDescription := C.CString(description)
	defer free(cDescription)

	C.message_expects_to_receive(i.handle, cDescription)

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

		C.message_with_metadata(i.handle, cName, cValue)
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

func (i *Message) WithContents(contentType string, body []byte) *Message {
	cHeader := C.CString(contentType)
	defer free(cHeader)

	cBytes := C.CString(string(body))
	defer free(cBytes)
	C.message_with_contents(i.handle, cHeader, (*C.char)(unsafe.Pointer(&body[0])), C.int(len(body)))

	return i
}

func (i *Message) ReifyMessage() string {
	return C.GoString(C.message_reify(i.handle))
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

	res := int(C.write_message_pact_file(m.messagePact.handle, cDir, C.int(overwritePact)))

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
