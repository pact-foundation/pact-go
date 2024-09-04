package native

/*
#include "pact.h"
*/
import "C"

import (
	"fmt"
	"log"
	"strings"
	"unsafe"
)

type Verifier struct {
	handle *C.VerifierHandle
}

func (v *Verifier) Verify(args []string) error {
	log.Println("[DEBUG] executing verifier FFI with args", args)
	cargs := C.CString(strings.Join(args, "\n"))
	defer free(cargs)
	result := C.pactffi_verify(cargs)

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1 | The verification process failed, see output for errors |
	/// | 2 | A null pointer was received |
	/// | 3 | The method panicked |
	switch int(result) {
	case 0:
		return nil
	case 1:
		return ErrVerifierFailed
	case 2:
		return ErrInvalidVerifierConfig
	case 3:
		return ErrVerifierPanic
	default:
		return fmt.Errorf("an unknown error (%d) ocurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

// Version returns the current semver FFI interface version
func (v *Verifier) Version() string {
	return Version()
}

var (
	// ErrVerifierPanic indicates a panic ocurred when invoking the verifier.
	ErrVerifierPanic = fmt.Errorf("a general panic occured when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier
	ErrInvalidVerifierConfig = fmt.Errorf("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	//ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful)
	ErrVerifierFailed = fmt.Errorf("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
	//ErrVerifierFailedToRun indicates the verification process was unable to run
	ErrVerifierFailedToRun = fmt.Errorf("the verifier failed to execute (this is most likely a defect in the framework)")
)

func NewVerifier(name string, version string) *Verifier {
	cName := C.CString(name)
	cVersion := C.CString(version)
	defer free(cName)
	defer free(cVersion)

	h := C.pactffi_verifier_new_for_application(cName, cVersion)

	return &Verifier{
		handle: h,
	}
}

func (v *Verifier) Shutdown() {
	C.pactffi_verifier_shutdown(v.handle)
}

func (v *Verifier) SetProviderInfo(name string, scheme string, host string, port uint16, path string) {
	cName := C.CString(name)
	defer free(cName)
	cScheme := C.CString(scheme)
	defer free(cScheme)
	cHost := C.CString(host)
	defer free(cHost)
	cPort := C.ushort(port)
	cPath := C.CString(path)
	defer free(cPath)

	C.pactffi_verifier_set_provider_info(v.handle, cName, cScheme, cHost, cPort, cPath)
}

func (v *Verifier) AddTransport(protocol string, port uint16, path string, scheme string) {
	log.Println("[DEBUG] Adding transport with protocol:", protocol, "port:", port, "path:", path, "scheme:", scheme)
	cProtocol := C.CString(protocol)
	defer free(cProtocol)
	cPort := C.ushort(port)
	cPath := C.CString(path)
	defer free(cPath)
	cScheme := C.CString(scheme)
	defer free(cScheme)

	C.pactffi_verifier_add_provider_transport(v.handle, cProtocol, cPort, cPath, cScheme)
}

func (v *Verifier) SetFilterInfo(description string, state string, noState bool) {
	cFilterDescription := C.CString(description)
	defer free(cFilterDescription)
	cFilterState := C.CString(state)
	defer free(cFilterState)

	C.pactffi_verifier_set_filter_info(v.handle, cFilterDescription, cFilterState, boolToCUchar(noState))
}

func (v *Verifier) SetProviderState(url string, teardown bool, body bool) {
	cURL := C.CString(url)
	defer free(cURL)

	C.pactffi_verifier_set_provider_state(v.handle, cURL, boolToCUchar(teardown), boolToCUchar(body))
}

func (v *Verifier) SetVerificationOptions(disableSSLVerification bool, requestTimeout int64) {
	// TODO: this returns an int and therefore can error. We should have all of these functions return values??
	C.pactffi_verifier_set_verification_options(v.handle, boolToCUchar(disableSSLVerification), C.ulong(requestTimeout))
}

func (v *Verifier) SetConsumerFilters(consumers []string) {
	// TODO: check if this actually works!
	C.pactffi_verifier_set_consumer_filters(v.handle, stringArrayToCStringArray(consumers), C.ushort(len(consumers)))
}

func (v *Verifier) AddCustomHeader(name string, value string) {
	cHeaderName := C.CString(name)
	defer free(cHeaderName)
	cHeaderValue := C.CString(value)
	defer free(cHeaderValue)

	C.pactffi_verifier_add_custom_header(v.handle, cHeaderName, cHeaderValue)
}

func (v *Verifier) AddFileSource(file string) {
	cFile := C.CString(file)
	defer free(cFile)

	C.pactffi_verifier_add_file_source(v.handle, cFile)
}

func (v *Verifier) AddDirectorySource(directory string) {
	cDirectory := C.CString(directory)
	defer free(cDirectory)

	C.pactffi_verifier_add_directory_source(v.handle, cDirectory)
}

func (v *Verifier) AddURLSource(url string, username string, password string, token string) {
	cUrl := C.CString(url)
	defer free(cUrl)
	cUsername := C.CString(username)
	defer free(cUsername)
	cPassword := C.CString(password)
	defer free(cPassword)
	cToken := C.CString(token)
	defer free(cToken)

	C.pactffi_verifier_url_source(v.handle, cUrl, cUsername, cPassword, cToken)
}

func (v *Verifier) BrokerSourceWithSelectors(url string, username string, password string, token string, enablePending bool, includeWipPactsSince string, providerTags []string, providerBranch string, selectors []string, consumerVersionTags []string) {
	cUrl := C.CString(url)
	defer free(cUrl)
	cUsername := C.CString(username)
	defer free(cUsername)
	cPassword := C.CString(password)
	defer free(cPassword)
	cToken := C.CString(token)
	defer free(cToken)
	cIncludeWipPactsSince := C.CString(includeWipPactsSince)
	defer free(cIncludeWipPactsSince)
	cProviderBranch := C.CString(providerBranch)
	defer free(cProviderBranch)

	C.pactffi_verifier_broker_source_with_selectors(v.handle, cUrl, cUsername, cPassword, cToken, boolToCUchar(enablePending), cIncludeWipPactsSince, stringArrayToCStringArray(providerTags), C.ushort(len(providerTags)), cProviderBranch, stringArrayToCStringArray(selectors), C.ushort(len(selectors)), stringArrayToCStringArray(consumerVersionTags), C.ushort(len(consumerVersionTags)))
}

func (v *Verifier) SetPublishOptions(providerVersion string, buildUrl string, providerTags []string, providerBranch string) {
	cProviderVersion := C.CString(providerVersion)
	defer free(cProviderVersion)
	cBuildUrl := C.CString(buildUrl)
	defer free(cBuildUrl)
	cProviderBranch := C.CString(providerBranch)
	defer free(cProviderBranch)

	C.pactffi_verifier_set_publish_options(v.handle, cProviderVersion, cBuildUrl, stringArrayToCStringArray(providerTags), C.ushort(len(providerTags)), cProviderBranch)
}

func (v *Verifier) Execute() error {
	// TODO: Validate
	result := C.pactffi_verifier_execute(v.handle)

	/// | Error | Description |
	/// |-------|-------------|
	/// | 1     | The verification process failed, see output for errors |
	switch int(result) {
	case 0:
		return nil
	case 1:
		return ErrVerifierFailed
	case 2:
		return ErrVerifierFailedToRun
	default:
		return fmt.Errorf("an unknown error (%d) ocurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

func (v *Verifier) SetNoPactsIsError(isError bool) {
	C.pactffi_verifier_set_no_pacts_is_error(v.handle, boolToCUchar(isError))
}

func (v *Verifier) SetColoredOutput(isColoredOutput bool) {
	C.pactffi_verifier_set_coloured_output(v.handle, boolToCUchar(isColoredOutput))
}

func stringArrayToCStringArray(inputs []string) **C.char {
	if len(inputs) == 0 {
		return nil
	}

	output := make([]*C.char, len(inputs))

	for i, consumer := range inputs {
		output[i] = C.CString(consumer)
	}

	return (**C.char)(unsafe.Pointer(&output[0]))
}

func boolToCUchar(val bool) C.uchar {
	if val {
		return C.uchar(1)
	}
	return C.uchar(0)
}
