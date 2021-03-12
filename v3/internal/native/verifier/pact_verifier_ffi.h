#ifndef PACT_VERIFIER_FFI_H
#define PACT_VERIFIER_FFI_H

/* Generated with cbindgen:0.14.3 */

#include <cstdarg>
#include <cstdint>
#include <cstdlib>
#include <new>

namespace pact_verifier_ffi {

extern "C" {

/// Frees the memory allocated to a string by another function
///
/// # Safety
///
/// Exported functions are inherently unsafe.
void free_string(char *s);

/// Initialise the mock server library, can provide an environment variable name to use to
/// set the log levels.
///
/// # Safety
///
/// Exported functions are inherently unsafe.
void init(const char *log_env_var);

/// External interface to verifier a provider
///
/// * `args` - the same as the CLI interface, except newline delimited
///
/// # Errors
///
/// Errors are returned as non-zero numeric values.
///
/// | Error | Description |
/// |-------|-------------|
/// | 1 | The verification process failed, see output for errors |
/// | 2 | A null pointer was received |
/// | 3 | The method panicked |
///
int32_t verify(const char *args);

/// Get the current library version
///
/// # Errors
///
/// An empty string indicates an error determining the current crate version
const char *version();

} // extern "C"

} // namespace pact_verifier_ffi

#endif // PACT_VERIFIER_FFI_H
