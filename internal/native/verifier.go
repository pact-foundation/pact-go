package native

/*
#include "pact.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

// Verifier is a Go representation of the pact_ffi VerifierHandle, used to
// verify a provider against one or more consumer pacts.
type Verifier struct {
	handle *C.VerifierHandle
}

// NewVerifier creates a new provider verifier, identifying the calling
// application as name and version (this module calls it with "pact-go"
// and this module's own version). The caller is responsible for calling
// Shutdown when done with it, to free the underlying resources.
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

// Version returns the current semver FFI interface version.
func (v *Verifier) Version() string {
	return Version()
}

var (
	// ErrVerifierPanic indicates a panic occurred when invoking the verifier.
	ErrVerifierPanic = errors.New("a general panic occurred when starting/invoking verifier (this indicates a defect in the framework)")

	// ErrInvalidVerifierConfig indicates an issue configuring the verifier.
	ErrInvalidVerifierConfig = errors.New("configuration for the verifier was invalid and an unknown error occurred (this is most likely a defect in the framework)")

	// ErrVerifierFailed and ErrVerifierFailedToRun are mutually exclusive: a
	// single Verifier call returns one or the other, never both.
	//
	// ErrVerifierFailed is the standard error if a verification failed (e.g. beacause the pact verification was not successful).
	ErrVerifierFailed = errors.New("the verifier failed to successfully verify the pacts, this indicates an issue with the provider API")
	// ErrVerifierFailedToRun indicates the verification process was unable to run.
	ErrVerifierFailedToRun = errors.New("the verifier failed to execute (this is most likely a defect in the framework)")
)

// Shutdown releases the resources held by this verifier.
func (v *Verifier) Shutdown() {
	C.pactffi_verifier_shutdown(v.handle)
}

// SetProviderInfo sets the name and network location of the provider
// under verification.
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

// AddTransport registers an additional transport (e.g. a message queue
// alongside HTTP) that the provider exposes for verification.
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

// SetFilterInfo restricts verification to interactions matching
// description and/or state; noState additionally restricts to
// interactions with no provider state.
func (v *Verifier) SetFilterInfo(description string, state string, noState bool) {
	cFilterDescription := C.CString(description)
	defer free(cFilterDescription)
	cFilterState := C.CString(state)
	defer free(cFilterState)

	C.pactffi_verifier_set_filter_info(v.handle, cFilterDescription, cFilterState, boolToCUchar(noState))
}

// SetProviderState sets the URL the verifier calls to set up (and, if
// teardown is set, tear down) provider states before each interaction.
// If body is set, state parameters are sent as a JSON request body rather
// than query parameters.
func (v *Verifier) SetProviderState(url string, teardown bool, body bool) {
	cURL := C.CString(url)
	defer free(cURL)

	C.pactffi_verifier_set_provider_state(v.handle, cURL, boolToCUchar(teardown), boolToCUchar(body))
}

// SetVerificationOptions sets the options used by the verifier when
// calling the provider.
func (v *Verifier) SetVerificationOptions(disableSSLVerification bool, requestTimeout int64) {
	// TODO: this returns an int and therefore can error. We should have all of these functions return values??
	C.pactffi_verifier_set_verification_options(v.handle, boolToCUchar(disableSSLVerification), C.ulong(requestTimeout))
}

// SetConsumerFilters is intended to restrict verification to pacts from
// the given consumers.
func (v *Verifier) SetConsumerFilters(consumers []string) {
	// TODO: check if this actually works! It is untested and its effect
	// on the underlying pact_ffi call has not been confirmed.
	C.pactffi_verifier_set_consumer_filters(v.handle, stringArrayToCStringArray(consumers), C.ushort(len(consumers)))
}

// AddCustomHeader adds a header to be sent with every request made to the
// provider during verification.
func (v *Verifier) AddCustomHeader(name string, value string) {
	cHeaderName := C.CString(name)
	defer free(cHeaderName)
	cHeaderValue := C.CString(value)
	defer free(cHeaderValue)

	C.pactffi_verifier_add_custom_header(v.handle, cHeaderName, cHeaderValue)
}

// AddFileSource adds a single Pact file as a source to verify.
func (v *Verifier) AddFileSource(file string) {
	cFile := C.CString(file)
	defer free(cFile)

	C.pactffi_verifier_add_file_source(v.handle, cFile)
}

// AddDirectorySource adds a directory as a source to verify: every pact
// file in it that matches the provider name is verified.
func (v *Verifier) AddDirectorySource(directory string) {
	cDirectory := C.CString(directory)
	defer free(cDirectory)

	C.pactffi_verifier_add_directory_source(v.handle, cDirectory)
}

// AddURLSource adds a URL as a source to verify: the pact file is fetched
// from url, using basic auth if username and password are set, or bearer
// token auth if token is set.
func (v *Verifier) AddURLSource(url string, username string, password string, token string) {
	cURL := C.CString(url)
	defer free(cURL)
	cUsername := C.CString(username)
	defer free(cUsername)
	cPassword := C.CString(password)
	defer free(cPassword)
	cToken := C.CString(token)
	defer free(cToken)

	C.pactffi_verifier_url_source(v.handle, cURL, cUsername, cPassword, cToken)
}

// BrokerSourceWithSelectors adds a Pact Broker as a source to verify,
// fetching every pact matching the provider name and the given consumer
// version selectors (see
// https://docs.pact.io/pact_broker/advanced_topics/consumer_version_selectors/).
// Authentication follows the same rules as AddURLSource.
func (v *Verifier) BrokerSourceWithSelectors(url string, username string, password string, token string, enablePending bool, includeWipPactsSince string, providerTags []string, providerBranch string, selectors []string, consumerVersionTags []string) {
	cURL := C.CString(url)
	defer free(cURL)
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

	C.pactffi_verifier_broker_source_with_selectors(v.handle, cURL, cUsername, cPassword, cToken, boolToCUchar(enablePending), cIncludeWipPactsSince, stringArrayToCStringArray(providerTags), C.ushort(len(providerTags)), cProviderBranch, stringArrayToCStringArray(selectors), C.ushort(len(selectors)), stringArrayToCStringArray(consumerVersionTags), C.ushort(len(consumerVersionTags)))
}

// SetPublishOptions sets the values needed to publish verification
// results back to the Pact Broker: providerVersion is required, the rest
// are optional and pass as "" or nil to omit.
func (v *Verifier) SetPublishOptions(providerVersion string, buildURL string, providerTags []string, providerBranch string) {
	cProviderVersion := C.CString(providerVersion)
	defer free(cProviderVersion)
	cBuildURL := C.CString(buildURL)
	defer free(cBuildURL)
	cProviderBranch := C.CString(providerBranch)
	defer free(cProviderBranch)

	C.pactffi_verifier_set_publish_options(v.handle, cProviderVersion, cBuildURL, stringArrayToCStringArray(providerTags), C.ushort(len(providerTags)), cProviderBranch)
}

// Execute runs the verification against every source and option
// configured on v, returning an error if verification failed or could
// not run.
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
		return fmt.Errorf("an unknown error (%d) occurred when verifying the provider (this indicates a defect in the framework)", int(result))
	}
}

// SetNoPactsIsError controls whether finding no pacts to verify is
// treated as a verification error.
func (v *Verifier) SetNoPactsIsError(isError bool) {
	C.pactffi_verifier_set_no_pacts_is_error(v.handle, boolToCUchar(isError))
}

// SetColoredOutput enables or disables ANSI colour codes in the verifier
// output; colour is enabled by default.
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
