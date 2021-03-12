#ifndef PACT_MOCK_SERVER_FFI_H
#define PACT_MOCK_SERVER_FFI_H

/* Generated with cbindgen:0.14.1 */

#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <new>

namespace pact_mock_server_ffi {

/// Wraps a Pact model struct
struct InteractionHandle {
  uintptr_t pact;
  uintptr_t interaction;
};

/// Wraps a Pact model struct
struct PactHandle {
  uintptr_t pact;
};

extern "C" {

/// External interface to cleanup a mock server. This function will try terminate the mock server
/// with the given port number and cleanup any memory allocated for it. Returns true, unless a
/// mock server with the given port number does not exist, or the function panics.
///
/// **NOTE:** Although `close()` on the listener for the mock server is called, this does not
/// currently work and the listener will continue handling requests. In this
/// case, it will always return a 404 once the mock server has been cleaned up.
bool cleanup_mock_server(int32_t mock_server_port);

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
int32_t create_mock_server(const char *pact_str,
                           const char *addr_str);

/// Adds a provider state to the Interaction
void given(InteractionHandle interaction, const char *description);

/// Initialise the mock server library
void init();

/// Get self signed certificate for TLS mode
char* get_tls_ca_certificate()

/// Free a string allocated on the Rust heap
void free_string(const char *s)

/// External interface to check if a mock server has matched all its requests. The port number is
/// passed in, and if all requests have been matched, true is returned. False is returned if there
/// is no mock server on the given port, or if any request has not been successfully matched, or
/// the method panics.
bool mock_server_matched(int32_t mock_server_port);

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
char *mock_server_mismatches(int32_t mock_server_port);

/// Creates a new Interaction and returns a handle to it
InteractionHandle new_interaction(PactHandle pact, const char *description);

/// Creates a new Pact model and returns a handle to it
PactHandle new_pact(const char *consumer_name, const char *provider_name);

/// Sets the description for the Interaction
void upon_receiving(InteractionHandle interaction, const char *description);

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
int32_t write_pact_file(int32_t mock_server_port, const char *directory);

} // extern "C"

} // namespace pact_mock_server_ffi

#endif // PACT_MOCK_SERVER_FFI_H
