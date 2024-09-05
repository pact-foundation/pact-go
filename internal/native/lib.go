// Package native contains the c bindings into the Pact Reference types.
package native

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

func getSystemLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libpact_ffi.dylib"
	case "linux":
		return "libpact_ffi.so"
	case "windows":
		return "pact_ffi.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

type (
	size_t uintptr
)

var pactffi_version func() string
var pactffi_init func(string)
var pactffi_init_with_log_level func(string)
var pactffi_enable_ansi_support func()
var pactffi_log_message func(string, string, string)
var pactffi_match_message func(uintptr, uintptr) uintptr
var pactffi_mismatches_get_iter func(uintptr) uintptr
var pactffi_mismatches_delete func(uintptr)
var pactffi_mismatches_iter_next func(uintptr) uintptr
var pactffi_mismatches_iter_delete func(uintptr)
var pactffi_mismatch_to_json func(uintptr) string
var pactffi_mismatch_type func(uintptr) string
var pactffi_mismatch_summary func(uintptr) string
var pactffi_mismatch_description func(uintptr) string
var pactffi_mismatch_ansi_description func(uintptr) string
var pactffi_get_error_message func(string, int32) int32
var pactffi_log_to_stdout func(int32) int32
var pactffi_log_to_stderr func(int32) int32
var pactffi_log_to_file func(string, int32) int32
var pactffi_log_to_buffer func(int32) int32
var pactffi_logger_init func()
var pactffi_logger_attach_sink func(string, int32) int32
var pactffi_logger_apply func() int32
var pactffi_fetch_log_buffer func(string) string
var pactffi_parse_pact_json func(string) uintptr
var pactffi_pact_model_delete func(uintptr)
var pactffi_pact_model_interaction_iterator func(uintptr) uintptr
var pactffi_pact_spec_version func(uintptr) int32
var pactffi_pact_interaction_delete func(uintptr)
var pactffi_async_message_new func() uintptr
var pactffi_async_message_delete func(uintptr)
var pactffi_async_message_get_contents func(uintptr) uintptr
var pactffi_async_message_get_contents_str func(uintptr) string
var pactffi_async_message_set_contents_str func(uintptr, string, string)
var pactffi_async_message_get_contents_length func(uintptr) size_t
var pactffi_async_message_get_contents_bin func(uintptr) uintptr
var pactffi_async_message_set_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_async_message_get_description func(uintptr) string
var pactffi_async_message_set_description func(uintptr, string) int32
var pactffi_async_message_get_provider_state func(uintptr, uint32) uintptr
var pactffi_async_message_get_provider_state_iter func(uintptr) uintptr
var pactffi_consumer_get_name func(uintptr) string
var pactffi_pact_get_consumer func(uintptr) uintptr
var pactffi_pact_consumer_delete func(uintptr)
var pactffi_message_contents_get_contents_str func(uintptr) string
var pactffi_message_contents_set_contents_str func(uintptr, string, string)
var pactffi_message_contents_get_contents_length func(uintptr) size_t
var pactffi_message_contents_get_contents_bin func(uintptr) uintptr
var pactffi_message_contents_set_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_message_contents_get_metadata_iter func(uintptr) uintptr
var pactffi_message_contents_get_matching_rule_iter func(uintptr, int32) uintptr
var pactffi_request_contents_get_matching_rule_iter func(uintptr, int32) uintptr
var pactffi_response_contents_get_matching_rule_iter func(uintptr, int32) uintptr
var pactffi_message_contents_get_generators_iter func(uintptr, int32) uintptr
var pactffi_request_contents_get_generators_iter func(uintptr, int32) uintptr
var pactffi_response_contents_get_generators_iter func(uintptr, int32) uintptr
var pactffi_parse_matcher_definition func(string) uintptr
var pactffi_matcher_definition_error func(uintptr) string
var pactffi_matcher_definition_value func(uintptr) string
var pactffi_matcher_definition_delete func(uintptr)
var pactffi_matcher_definition_generator func(uintptr) uintptr
var pactffi_matcher_definition_value_type func(uintptr) int32
var pactffi_matching_rule_iter_delete func(uintptr)
var pactffi_matcher_definition_iter func(uintptr) uintptr
var pactffi_matching_rule_iter_next func(uintptr) uintptr
var pactffi_matching_rule_id func(uintptr) uint16
var pactffi_matching_rule_value func(uintptr) string
var pactffi_matching_rule_pointer func(uintptr) uintptr
var pactffi_matching_rule_reference_name func(uintptr) string
var pactffi_validate_datetime func(string, string) int32
var pactffi_generator_to_json func(uintptr) string
var pactffi_generator_generate_string func(uintptr, string) string
var pactffi_generator_generate_integer func(uintptr, string) uint16
var pactffi_generators_iter_delete func(uintptr)
var pactffi_generators_iter_next func(uintptr) uintptr
var pactffi_generators_iter_pair_delete func(uintptr)
var pactffi_sync_http_new func() uintptr
var pactffi_sync_http_delete func(uintptr)
var pactffi_sync_http_get_request func(uintptr) uintptr
var pactffi_sync_http_get_request_contents func(uintptr) string
var pactffi_sync_http_set_request_contents func(uintptr, string, string)
var pactffi_sync_http_get_request_contents_length func(uintptr) size_t
var pactffi_sync_http_get_request_contents_bin func(uintptr) uintptr
var pactffi_sync_http_set_request_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_sync_http_get_response func(uintptr) uintptr
var pactffi_sync_http_get_response_contents func(uintptr) string
var pactffi_sync_http_set_response_contents func(uintptr, string, string)
var pactffi_sync_http_get_response_contents_length func(uintptr) size_t
var pactffi_sync_http_get_response_contents_bin func(uintptr) uintptr
var pactffi_sync_http_set_response_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_sync_http_get_description func(uintptr) string
var pactffi_sync_http_set_description func(uintptr, string) int32
var pactffi_sync_http_get_provider_state func(uintptr, uint32) uintptr
var pactffi_sync_http_get_provider_state_iter func(uintptr) uintptr
var pactffi_pact_interaction_as_synchronous_http func(uintptr) uintptr
var pactffi_pact_interaction_as_message func(uintptr) uintptr
var pactffi_pact_interaction_as_asynchronous_message func(uintptr) uintptr
var pactffi_pact_interaction_as_synchronous_message func(uintptr) uintptr
var pactffi_pact_message_iter_delete func(uintptr)
var pactffi_pact_message_iter_next func(uintptr) uintptr
var pactffi_pact_sync_message_iter_next func(uintptr) uintptr
var pactffi_pact_sync_message_iter_delete func(uintptr)
var pactffi_pact_sync_http_iter_next func(uintptr) uintptr
var pactffi_pact_sync_http_iter_delete func(uintptr)
var pactffi_pact_interaction_iter_next func(uintptr) uintptr
var pactffi_pact_interaction_iter_delete func(uintptr)
var pactffi_matching_rule_to_json func(uintptr) string
var pactffi_matching_rules_iter_delete func(uintptr)
var pactffi_matching_rules_iter_next func(uintptr) uintptr
var pactffi_matching_rules_iter_pair_delete func(uintptr)
var pactffi_message_new func() uintptr
var pactffi_message_new_from_json func(uint32, string, int32) uintptr
var pactffi_message_new_from_body func(string, string) uintptr
var pactffi_message_delete func(uintptr)
var pactffi_message_get_contents func(uintptr) string
var pactffi_message_set_contents func(uintptr, string, string)
var pactffi_message_get_contents_length func(uintptr) size_t
var pactffi_message_get_contents_bin func(uintptr) uintptr
var pactffi_message_set_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_message_get_description func(uintptr) string
var pactffi_message_set_description func(uintptr, string) int32
var pactffi_message_get_provider_state func(uintptr, uint32) uintptr
var pactffi_message_get_provider_state_iter func(uintptr) uintptr
var pactffi_provider_state_iter_next func(uintptr) uintptr
var pactffi_provider_state_iter_delete func(uintptr)
var pactffi_message_find_metadata func(uintptr, string) string
var pactffi_message_insert_metadata func(uintptr, string, string) int32
var pactffi_message_metadata_iter_next func(uintptr) uintptr
var pactffi_message_get_metadata_iter func(uintptr) uintptr
var pactffi_message_metadata_iter_delete func(uintptr)
var pactffi_message_metadata_pair_delete func(uintptr)
var pactffi_message_pact_new_from_json func(string, string) uintptr
var pactffi_message_pact_delete func(uintptr)
var pactffi_message_pact_get_consumer func(uintptr) uintptr
var pactffi_message_pact_get_provider func(uintptr) uintptr
var pactffi_message_pact_get_message_iter func(uintptr) uintptr
var pactffi_message_pact_message_iter_next func(uintptr) uintptr
var pactffi_message_pact_message_iter_delete func(uintptr)
var pactffi_message_pact_find_metadata func(uintptr, string, string) string
var pactffi_message_pact_get_metadata_iter func(uintptr) uintptr
var pactffi_message_pact_metadata_iter_next func(uintptr) uintptr
var pactffi_message_pact_metadata_iter_delete func(uintptr)
var pactffi_message_pact_metadata_triple_delete func(uintptr)
var pactffi_provider_get_name func(uintptr) string
var pactffi_pact_get_provider func(uintptr) uintptr
var pactffi_pact_provider_delete func(uintptr)
var pactffi_provider_state_get_name func(uintptr) string
var pactffi_provider_state_get_param_iter func(uintptr) uintptr
var pactffi_provider_state_param_iter_next func(uintptr) uintptr
var pactffi_provider_state_delete func(uintptr)
var pactffi_provider_state_param_iter_delete func(uintptr)
var pactffi_provider_state_param_pair_delete func(uintptr)
var pactffi_sync_message_new func() uintptr
var pactffi_sync_message_delete func(uintptr)
var pactffi_sync_message_get_request_contents_str func(uintptr) string
var pactffi_sync_message_set_request_contents_str func(uintptr, string, string)
var pactffi_sync_message_get_request_contents_length func(uintptr) size_t
var pactffi_sync_message_get_request_contents_bin func(uintptr) uintptr
var pactffi_sync_message_set_request_contents_bin func(uintptr, uintptr, size_t, string)
var pactffi_sync_message_get_request_contents func(uintptr) uintptr
var pactffi_sync_message_get_number_responses func(uintptr) size_t
var pactffi_sync_message_get_response_contents_str func(uintptr, size_t) string
var pactffi_sync_message_set_response_contents_str func(uintptr, size_t, string, string)
var pactffi_sync_message_get_response_contents_length func(uintptr, size_t) size_t
var pactffi_sync_message_get_response_contents_bin func(uintptr, size_t) uintptr
var pactffi_sync_message_set_response_contents_bin func(uintptr, size_t, uintptr, size_t, string)
var pactffi_sync_message_get_response_contents func(uintptr, size_t) uintptr
var pactffi_sync_message_get_description func(uintptr) string
var pactffi_sync_message_set_description func(uintptr, string) int32
var pactffi_sync_message_get_provider_state func(uintptr, uint32) uintptr
var pactffi_sync_message_get_provider_state_iter func(uintptr) uintptr
var pactffi_string_delete func(string)
var pactffi_create_mock_server func(string, string, bool) int32
var pactffi_get_tls_ca_certificate func() string
var pactffi_create_mock_server_for_pact func(uintptr, string, bool) int32
var pactffi_create_mock_server_for_transport func(uintptr, string, uint16, string, string) int32
var pactffi_mock_server_matched func(int32) bool
var pactffi_mock_server_mismatches func(int32) string
var pactffi_cleanup_mock_server func(int32) bool
var pactffi_write_pact_file func(int32, string, bool) int32
var pactffi_mock_server_logs func(int32) string
var pactffi_generate_datetime_string func(string) uintptr
var pactffi_check_regex func(string, string) bool
var pactffi_generate_regex_value func(string) uintptr
var pactffi_free_string func(uintptr)
var pactffi_new_pact func(string, string) uintptr
var pactffi_pact_handle_to_pointer func(uint16) uintptr
var pactffi_new_interaction func(uintptr, string) uintptr
var pactffi_new_message_interaction func(uintptr, string) uintptr
var pactffi_new_sync_message_interaction func(uintptr, string) uintptr
var pactffi_upon_receiving func(uintptr, string) bool
var pactffi_given func(uintptr, string) bool
var pactffi_interaction_test_name func(uint32, string) uint32
var pactffi_given_with_param func(uintptr, string, string, string) bool
var pactffi_given_with_params func(uintptr, string, string) int32
var pactffi_with_request func(uintptr, string, string) bool
var pactffi_with_query_parameter func(uintptr, string, int, string) bool
var pactffi_with_query_parameter_v2 func(uintptr, string, int, string) bool
var pactffi_with_specification func(uintptr, int32) bool
var pactffi_handle_get_pact_spec_version func(uint16) int32
var pactffi_with_pact_metadata func(uintptr, string, string, string) bool
var pactffi_with_header func(uintptr, int32, string, int, string) bool
var pactffi_with_header_v2 func(uintptr, int32, string, int, string) bool
var pactffi_set_header func(uint32, int32, string, string) bool
var pactffi_response_status func(uintptr, uint16) bool
var pactffi_response_status_v2 func(uintptr, string) bool
var pactffi_with_body func(uintptr, int32, string, string) bool
var pactffi_with_binary_body func(uint32, int32, string, string, size_t) bool
var pactffi_with_binary_file func(uintptr, int32, string, string, size_t) bool
var pactffi_with_matching_rules func(uint32, int32, string) bool
var pactffi_with_multipart_file_v2 func(uint32, int32, string, string, string, string) uintptr
var pactffi_with_multipart_file func(uintptr, int32, string, string, string) uintptr
var pactffi_pact_handle_get_message_iter func(uintptr) uintptr
var pactffi_pact_handle_get_sync_message_iter func(uintptr) uintptr
var pactffi_pact_handle_get_sync_http_iter func(uint16) uintptr
var pactffi_new_message_pact func(string, string) uintptr
var pactffi_new_message func(uint16, string) uint32
var pactffi_with_metadata func(uintptr, string, string, int) bool
var pactffi_message_expects_to_receive func(uintptr, string)
var pactffi_message_given func(uintptr, string)
var pactffi_message_given_with_param func(uintptr, string, string, string)
var pactffi_message_with_contents func(uintptr, string, uintptr, size_t)
var pactffi_message_with_metadata func(uintptr, string, string)
var pactffi_message_with_metadata_v2 func(uint32, string, string)
var pactffi_message_reify func(uint32) string
var pactffi_write_message_pact_file func(uintptr, string, bool) int32
var pactffi_with_message_pact_metadata func(uintptr, string, string, string)
var pactffi_pact_handle_write_file func(uint16, string, bool) int32
var pactffi_new_async_message func(uint16, string) uint32
var pactffi_free_pact_handle func(uint16) uint32
var pactffi_free_message_pact_handle func(uint16) uint32
var pactffi_verify func(string) int32
var pactffi_verifier_new func() uintptr
var pactffi_verifier_new_for_application func(string, string) uintptr
var pactffi_verifier_shutdown func(uintptr)
var pactffi_verifier_set_provider_info func(uintptr, string, string, string, uint16, string)
var pactffi_verifier_add_provider_transport func(uintptr, string, uint16, string, string)
var pactffi_verifier_set_filter_info func(uintptr, string, string, uint8)
var pactffi_verifier_set_provider_state func(uintptr, string, uint8, uint8)
var pactffi_verifier_set_verification_options func(uintptr, uint8, uint64) int32
var pactffi_verifier_set_coloured_output func(uintptr, uint8) int32
var pactffi_verifier_set_no_pacts_is_error func(uintptr, uint8) int32
var pactffi_verifier_set_publish_options func(uintptr, string, string, []*byte, uint16, string) int32
var pactffi_verifier_set_consumer_filters func(uintptr, []*byte, uint16)
var pactffi_verifier_add_custom_header func(uintptr, string, string)
var pactffi_verifier_add_file_source func(uintptr, string)
var pactffi_verifier_add_directory_source func(uintptr, string)
var pactffi_verifier_url_source func(uintptr, string, string, string, string)
var pactffi_verifier_broker_source func(uintptr, string, string, string, string)
var pactffi_verifier_broker_source_with_selectors func(uintptr, string, string, string, string, uint8, string, []*byte, uint16, string, []*byte, uint16, []*byte, uint16)
var pactffi_verifier_execute func(uintptr) int32
var pactffi_verifier_cli_args func() string
var pactffi_verifier_logs func(uintptr) string
var pactffi_verifier_logs_for_provider func(string) string
var pactffi_verifier_output func(uintptr, uint8) string
var pactffi_verifier_json func(uintptr) string
var pactffi_using_plugin func(uintptr, string, string) uint32
var pactffi_cleanup_plugins func(uintptr)
var pactffi_interaction_contents func(uintptr, int32, string, string) uint32
var pactffi_matches_string_value func(uintptr, string, string, uint8) string
var pactffi_matches_u64_value func(uintptr, uint64, uint64, uint8) string
var pactffi_matches_i64_value func(uintptr, int64, int64, uint8) string
var pactffi_matches_f64_value func(uintptr, float64, float64, uint8) string
var pactffi_matches_bool_value func(uintptr, uint8, uint8, uint8) string
var pactffi_matches_binary_value func(uintptr, uintptr, uint64, uintptr, uint64, uint8) string
var pactffi_matches_json_value func(uintptr, string, string, uint8) string

func init() {
	libpact_ffi, err := openLibrary(filepath.Join(os.Getenv("PACT_LD_LIBRARY_PATH"), getSystemLibrary()))
	if err != nil {
		panic(err)
	}
	purego.RegisterLibFunc(&pactffi_version, libpact_ffi, "pactffi_version")
	purego.RegisterLibFunc(&pactffi_init, libpact_ffi, "pactffi_init")
	purego.RegisterLibFunc(&pactffi_init_with_log_level, libpact_ffi, "pactffi_init_with_log_level")
	purego.RegisterLibFunc(&pactffi_enable_ansi_support, libpact_ffi, "pactffi_enable_ansi_support")
	purego.RegisterLibFunc(&pactffi_log_message, libpact_ffi, "pactffi_log_message")
	purego.RegisterLibFunc(&pactffi_match_message, libpact_ffi, "pactffi_match_message")
	purego.RegisterLibFunc(&pactffi_mismatches_get_iter, libpact_ffi, "pactffi_mismatches_get_iter")
	purego.RegisterLibFunc(&pactffi_mismatches_delete, libpact_ffi, "pactffi_mismatches_delete")
	purego.RegisterLibFunc(&pactffi_mismatches_iter_next, libpact_ffi, "pactffi_mismatches_iter_next")
	purego.RegisterLibFunc(&pactffi_mismatches_iter_delete, libpact_ffi, "pactffi_mismatches_iter_delete")
	purego.RegisterLibFunc(&pactffi_mismatch_to_json, libpact_ffi, "pactffi_mismatch_to_json")
	purego.RegisterLibFunc(&pactffi_mismatch_type, libpact_ffi, "pactffi_mismatch_type")
	purego.RegisterLibFunc(&pactffi_mismatch_summary, libpact_ffi, "pactffi_mismatch_summary")
	purego.RegisterLibFunc(&pactffi_mismatch_description, libpact_ffi, "pactffi_mismatch_description")
	purego.RegisterLibFunc(&pactffi_mismatch_ansi_description, libpact_ffi, "pactffi_mismatch_ansi_description")
	purego.RegisterLibFunc(&pactffi_get_error_message, libpact_ffi, "pactffi_get_error_message")
	purego.RegisterLibFunc(&pactffi_log_to_stdout, libpact_ffi, "pactffi_log_to_stdout")
	purego.RegisterLibFunc(&pactffi_log_to_stderr, libpact_ffi, "pactffi_log_to_stderr")
	purego.RegisterLibFunc(&pactffi_log_to_file, libpact_ffi, "pactffi_log_to_file")
	purego.RegisterLibFunc(&pactffi_log_to_buffer, libpact_ffi, "pactffi_log_to_buffer")
	purego.RegisterLibFunc(&pactffi_logger_init, libpact_ffi, "pactffi_logger_init")
	purego.RegisterLibFunc(&pactffi_logger_attach_sink, libpact_ffi, "pactffi_logger_attach_sink")
	purego.RegisterLibFunc(&pactffi_logger_apply, libpact_ffi, "pactffi_logger_apply")
	purego.RegisterLibFunc(&pactffi_fetch_log_buffer, libpact_ffi, "pactffi_fetch_log_buffer")
	purego.RegisterLibFunc(&pactffi_parse_pact_json, libpact_ffi, "pactffi_parse_pact_json")
	purego.RegisterLibFunc(&pactffi_pact_model_delete, libpact_ffi, "pactffi_pact_model_delete")
	purego.RegisterLibFunc(&pactffi_pact_model_interaction_iterator, libpact_ffi, "pactffi_pact_model_interaction_iterator")
	purego.RegisterLibFunc(&pactffi_pact_spec_version, libpact_ffi, "pactffi_pact_spec_version")
	purego.RegisterLibFunc(&pactffi_pact_interaction_delete, libpact_ffi, "pactffi_pact_interaction_delete")
	purego.RegisterLibFunc(&pactffi_async_message_new, libpact_ffi, "pactffi_async_message_new")
	purego.RegisterLibFunc(&pactffi_async_message_delete, libpact_ffi, "pactffi_async_message_delete")
	purego.RegisterLibFunc(&pactffi_async_message_get_contents, libpact_ffi, "pactffi_async_message_get_contents")
	purego.RegisterLibFunc(&pactffi_async_message_get_contents_str, libpact_ffi, "pactffi_async_message_get_contents_str")
	purego.RegisterLibFunc(&pactffi_async_message_set_contents_str, libpact_ffi, "pactffi_async_message_set_contents_str")
	purego.RegisterLibFunc(&pactffi_async_message_get_contents_length, libpact_ffi, "pactffi_async_message_get_contents_length")
	purego.RegisterLibFunc(&pactffi_async_message_get_contents_bin, libpact_ffi, "pactffi_async_message_get_contents_bin")
	purego.RegisterLibFunc(&pactffi_async_message_set_contents_bin, libpact_ffi, "pactffi_async_message_set_contents_bin")
	purego.RegisterLibFunc(&pactffi_async_message_get_description, libpact_ffi, "pactffi_async_message_get_description")
	purego.RegisterLibFunc(&pactffi_async_message_set_description, libpact_ffi, "pactffi_async_message_set_description")
	purego.RegisterLibFunc(&pactffi_async_message_get_provider_state, libpact_ffi, "pactffi_async_message_get_provider_state")
	purego.RegisterLibFunc(&pactffi_async_message_get_provider_state_iter, libpact_ffi, "pactffi_async_message_get_provider_state_iter")
	purego.RegisterLibFunc(&pactffi_consumer_get_name, libpact_ffi, "pactffi_consumer_get_name")
	purego.RegisterLibFunc(&pactffi_pact_get_consumer, libpact_ffi, "pactffi_pact_get_consumer")
	purego.RegisterLibFunc(&pactffi_pact_consumer_delete, libpact_ffi, "pactffi_pact_consumer_delete")
	purego.RegisterLibFunc(&pactffi_message_contents_get_contents_str, libpact_ffi, "pactffi_message_contents_get_contents_str")
	purego.RegisterLibFunc(&pactffi_message_contents_set_contents_str, libpact_ffi, "pactffi_message_contents_set_contents_str")
	purego.RegisterLibFunc(&pactffi_message_contents_get_contents_length, libpact_ffi, "pactffi_message_contents_get_contents_length")
	purego.RegisterLibFunc(&pactffi_message_contents_get_contents_bin, libpact_ffi, "pactffi_message_contents_get_contents_bin")
	purego.RegisterLibFunc(&pactffi_message_contents_set_contents_bin, libpact_ffi, "pactffi_message_contents_set_contents_bin")
	purego.RegisterLibFunc(&pactffi_message_contents_get_metadata_iter, libpact_ffi, "pactffi_message_contents_get_metadata_iter")
	purego.RegisterLibFunc(&pactffi_message_contents_get_matching_rule_iter, libpact_ffi, "pactffi_message_contents_get_matching_rule_iter")
	purego.RegisterLibFunc(&pactffi_request_contents_get_matching_rule_iter, libpact_ffi, "pactffi_request_contents_get_matching_rule_iter")
	purego.RegisterLibFunc(&pactffi_response_contents_get_matching_rule_iter, libpact_ffi, "pactffi_response_contents_get_matching_rule_iter")
	purego.RegisterLibFunc(&pactffi_message_contents_get_generators_iter, libpact_ffi, "pactffi_message_contents_get_generators_iter")
	purego.RegisterLibFunc(&pactffi_request_contents_get_generators_iter, libpact_ffi, "pactffi_request_contents_get_generators_iter")
	purego.RegisterLibFunc(&pactffi_response_contents_get_generators_iter, libpact_ffi, "pactffi_response_contents_get_generators_iter")
	purego.RegisterLibFunc(&pactffi_parse_matcher_definition, libpact_ffi, "pactffi_parse_matcher_definition")
	purego.RegisterLibFunc(&pactffi_matcher_definition_error, libpact_ffi, "pactffi_matcher_definition_error")
	purego.RegisterLibFunc(&pactffi_matcher_definition_value, libpact_ffi, "pactffi_matcher_definition_value")
	purego.RegisterLibFunc(&pactffi_matcher_definition_delete, libpact_ffi, "pactffi_matcher_definition_delete")
	purego.RegisterLibFunc(&pactffi_matcher_definition_generator, libpact_ffi, "pactffi_matcher_definition_generator")
	purego.RegisterLibFunc(&pactffi_matcher_definition_value_type, libpact_ffi, "pactffi_matcher_definition_value_type")
	purego.RegisterLibFunc(&pactffi_matching_rule_iter_delete, libpact_ffi, "pactffi_matching_rule_iter_delete")
	purego.RegisterLibFunc(&pactffi_matcher_definition_iter, libpact_ffi, "pactffi_matcher_definition_iter")
	purego.RegisterLibFunc(&pactffi_matching_rule_iter_next, libpact_ffi, "pactffi_matching_rule_iter_next")
	purego.RegisterLibFunc(&pactffi_matching_rule_id, libpact_ffi, "pactffi_matching_rule_id")
	purego.RegisterLibFunc(&pactffi_matching_rule_value, libpact_ffi, "pactffi_matching_rule_value")
	purego.RegisterLibFunc(&pactffi_matching_rule_pointer, libpact_ffi, "pactffi_matching_rule_pointer")
	purego.RegisterLibFunc(&pactffi_matching_rule_reference_name, libpact_ffi, "pactffi_matching_rule_reference_name")
	purego.RegisterLibFunc(&pactffi_validate_datetime, libpact_ffi, "pactffi_validate_datetime")
	purego.RegisterLibFunc(&pactffi_generator_to_json, libpact_ffi, "pactffi_generator_to_json")
	purego.RegisterLibFunc(&pactffi_generator_generate_string, libpact_ffi, "pactffi_generator_generate_string")
	purego.RegisterLibFunc(&pactffi_generator_generate_integer, libpact_ffi, "pactffi_generator_generate_integer")
	purego.RegisterLibFunc(&pactffi_generators_iter_delete, libpact_ffi, "pactffi_generators_iter_delete")
	purego.RegisterLibFunc(&pactffi_generators_iter_next, libpact_ffi, "pactffi_generators_iter_next")
	purego.RegisterLibFunc(&pactffi_generators_iter_pair_delete, libpact_ffi, "pactffi_generators_iter_pair_delete")
	purego.RegisterLibFunc(&pactffi_sync_http_new, libpact_ffi, "pactffi_sync_http_new")
	purego.RegisterLibFunc(&pactffi_sync_http_delete, libpact_ffi, "pactffi_sync_http_delete")
	purego.RegisterLibFunc(&pactffi_sync_http_get_request, libpact_ffi, "pactffi_sync_http_get_request")
	purego.RegisterLibFunc(&pactffi_sync_http_get_request_contents, libpact_ffi, "pactffi_sync_http_get_request_contents")
	purego.RegisterLibFunc(&pactffi_sync_http_set_request_contents, libpact_ffi, "pactffi_sync_http_set_request_contents")
	purego.RegisterLibFunc(&pactffi_sync_http_get_request_contents_length, libpact_ffi, "pactffi_sync_http_get_request_contents_length")
	purego.RegisterLibFunc(&pactffi_sync_http_get_request_contents_bin, libpact_ffi, "pactffi_sync_http_get_request_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_http_set_request_contents_bin, libpact_ffi, "pactffi_sync_http_set_request_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_http_get_response, libpact_ffi, "pactffi_sync_http_get_response")
	purego.RegisterLibFunc(&pactffi_sync_http_get_response_contents, libpact_ffi, "pactffi_sync_http_get_response_contents")
	purego.RegisterLibFunc(&pactffi_sync_http_set_response_contents, libpact_ffi, "pactffi_sync_http_set_response_contents")
	purego.RegisterLibFunc(&pactffi_sync_http_get_response_contents_length, libpact_ffi, "pactffi_sync_http_get_response_contents_length")
	purego.RegisterLibFunc(&pactffi_sync_http_get_response_contents_bin, libpact_ffi, "pactffi_sync_http_get_response_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_http_set_response_contents_bin, libpact_ffi, "pactffi_sync_http_set_response_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_http_get_description, libpact_ffi, "pactffi_sync_http_get_description")
	purego.RegisterLibFunc(&pactffi_sync_http_set_description, libpact_ffi, "pactffi_sync_http_set_description")
	purego.RegisterLibFunc(&pactffi_sync_http_get_provider_state, libpact_ffi, "pactffi_sync_http_get_provider_state")
	purego.RegisterLibFunc(&pactffi_sync_http_get_provider_state_iter, libpact_ffi, "pactffi_sync_http_get_provider_state_iter")
	purego.RegisterLibFunc(&pactffi_pact_interaction_as_synchronous_http, libpact_ffi, "pactffi_pact_interaction_as_synchronous_http")
	purego.RegisterLibFunc(&pactffi_pact_interaction_as_message, libpact_ffi, "pactffi_pact_interaction_as_message")
	purego.RegisterLibFunc(&pactffi_pact_interaction_as_asynchronous_message, libpact_ffi, "pactffi_pact_interaction_as_asynchronous_message")
	purego.RegisterLibFunc(&pactffi_pact_interaction_as_synchronous_message, libpact_ffi, "pactffi_pact_interaction_as_synchronous_message")
	purego.RegisterLibFunc(&pactffi_pact_message_iter_delete, libpact_ffi, "pactffi_pact_message_iter_delete")
	purego.RegisterLibFunc(&pactffi_pact_message_iter_next, libpact_ffi, "pactffi_pact_message_iter_next")
	purego.RegisterLibFunc(&pactffi_pact_sync_message_iter_next, libpact_ffi, "pactffi_pact_sync_message_iter_next")
	purego.RegisterLibFunc(&pactffi_pact_sync_message_iter_delete, libpact_ffi, "pactffi_pact_sync_message_iter_delete")
	purego.RegisterLibFunc(&pactffi_pact_sync_http_iter_next, libpact_ffi, "pactffi_pact_sync_http_iter_next")
	purego.RegisterLibFunc(&pactffi_pact_sync_http_iter_delete, libpact_ffi, "pactffi_pact_sync_http_iter_delete")
	purego.RegisterLibFunc(&pactffi_pact_interaction_iter_next, libpact_ffi, "pactffi_pact_interaction_iter_next")
	purego.RegisterLibFunc(&pactffi_pact_interaction_iter_delete, libpact_ffi, "pactffi_pact_interaction_iter_delete")
	purego.RegisterLibFunc(&pactffi_matching_rule_to_json, libpact_ffi, "pactffi_matching_rule_to_json")
	purego.RegisterLibFunc(&pactffi_matching_rules_iter_delete, libpact_ffi, "pactffi_matching_rules_iter_delete")
	purego.RegisterLibFunc(&pactffi_matching_rules_iter_next, libpact_ffi, "pactffi_matching_rules_iter_next")
	purego.RegisterLibFunc(&pactffi_matching_rules_iter_pair_delete, libpact_ffi, "pactffi_matching_rules_iter_pair_delete")
	purego.RegisterLibFunc(&pactffi_message_new, libpact_ffi, "pactffi_message_new")
	purego.RegisterLibFunc(&pactffi_message_new_from_json, libpact_ffi, "pactffi_message_new_from_json")
	purego.RegisterLibFunc(&pactffi_message_new_from_body, libpact_ffi, "pactffi_message_new_from_body")
	purego.RegisterLibFunc(&pactffi_message_delete, libpact_ffi, "pactffi_message_delete")
	purego.RegisterLibFunc(&pactffi_message_get_contents, libpact_ffi, "pactffi_message_get_contents")
	purego.RegisterLibFunc(&pactffi_message_set_contents, libpact_ffi, "pactffi_message_set_contents")
	purego.RegisterLibFunc(&pactffi_message_get_contents_length, libpact_ffi, "pactffi_message_get_contents_length")
	purego.RegisterLibFunc(&pactffi_message_get_contents_bin, libpact_ffi, "pactffi_message_get_contents_bin")
	purego.RegisterLibFunc(&pactffi_message_set_contents_bin, libpact_ffi, "pactffi_message_set_contents_bin")
	purego.RegisterLibFunc(&pactffi_message_get_description, libpact_ffi, "pactffi_message_get_description")
	purego.RegisterLibFunc(&pactffi_message_set_description, libpact_ffi, "pactffi_message_set_description")
	purego.RegisterLibFunc(&pactffi_message_get_provider_state, libpact_ffi, "pactffi_message_get_provider_state")
	purego.RegisterLibFunc(&pactffi_message_get_provider_state_iter, libpact_ffi, "pactffi_message_get_provider_state_iter")
	purego.RegisterLibFunc(&pactffi_provider_state_iter_next, libpact_ffi, "pactffi_provider_state_iter_next")
	purego.RegisterLibFunc(&pactffi_provider_state_iter_delete, libpact_ffi, "pactffi_provider_state_iter_delete")
	purego.RegisterLibFunc(&pactffi_message_find_metadata, libpact_ffi, "pactffi_message_find_metadata")
	purego.RegisterLibFunc(&pactffi_message_insert_metadata, libpact_ffi, "pactffi_message_insert_metadata")
	purego.RegisterLibFunc(&pactffi_message_metadata_iter_next, libpact_ffi, "pactffi_message_metadata_iter_next")
	purego.RegisterLibFunc(&pactffi_message_get_metadata_iter, libpact_ffi, "pactffi_message_get_metadata_iter")
	purego.RegisterLibFunc(&pactffi_message_metadata_iter_delete, libpact_ffi, "pactffi_message_metadata_iter_delete")
	purego.RegisterLibFunc(&pactffi_message_metadata_pair_delete, libpact_ffi, "pactffi_message_metadata_pair_delete")
	purego.RegisterLibFunc(&pactffi_message_pact_new_from_json, libpact_ffi, "pactffi_message_pact_new_from_json")
	purego.RegisterLibFunc(&pactffi_message_pact_delete, libpact_ffi, "pactffi_message_pact_delete")
	purego.RegisterLibFunc(&pactffi_message_pact_get_consumer, libpact_ffi, "pactffi_message_pact_get_consumer")
	purego.RegisterLibFunc(&pactffi_message_pact_get_provider, libpact_ffi, "pactffi_message_pact_get_provider")
	purego.RegisterLibFunc(&pactffi_message_pact_get_message_iter, libpact_ffi, "pactffi_message_pact_get_message_iter")
	purego.RegisterLibFunc(&pactffi_message_pact_message_iter_next, libpact_ffi, "pactffi_message_pact_message_iter_next")
	purego.RegisterLibFunc(&pactffi_message_pact_message_iter_delete, libpact_ffi, "pactffi_message_pact_message_iter_delete")
	purego.RegisterLibFunc(&pactffi_message_pact_find_metadata, libpact_ffi, "pactffi_message_pact_find_metadata")
	purego.RegisterLibFunc(&pactffi_message_pact_get_metadata_iter, libpact_ffi, "pactffi_message_pact_get_metadata_iter")
	purego.RegisterLibFunc(&pactffi_message_pact_metadata_iter_next, libpact_ffi, "pactffi_message_pact_metadata_iter_next")
	purego.RegisterLibFunc(&pactffi_message_pact_metadata_iter_delete, libpact_ffi, "pactffi_message_pact_metadata_iter_delete")
	purego.RegisterLibFunc(&pactffi_message_pact_metadata_triple_delete, libpact_ffi, "pactffi_message_pact_metadata_triple_delete")
	purego.RegisterLibFunc(&pactffi_provider_get_name, libpact_ffi, "pactffi_provider_get_name")
	purego.RegisterLibFunc(&pactffi_pact_get_provider, libpact_ffi, "pactffi_pact_get_provider")
	purego.RegisterLibFunc(&pactffi_pact_provider_delete, libpact_ffi, "pactffi_pact_provider_delete")
	purego.RegisterLibFunc(&pactffi_provider_state_get_name, libpact_ffi, "pactffi_provider_state_get_name")
	purego.RegisterLibFunc(&pactffi_provider_state_get_param_iter, libpact_ffi, "pactffi_provider_state_get_param_iter")
	purego.RegisterLibFunc(&pactffi_provider_state_param_iter_next, libpact_ffi, "pactffi_provider_state_param_iter_next")
	purego.RegisterLibFunc(&pactffi_provider_state_delete, libpact_ffi, "pactffi_provider_state_delete")
	purego.RegisterLibFunc(&pactffi_provider_state_param_iter_delete, libpact_ffi, "pactffi_provider_state_param_iter_delete")
	purego.RegisterLibFunc(&pactffi_provider_state_param_pair_delete, libpact_ffi, "pactffi_provider_state_param_pair_delete")
	purego.RegisterLibFunc(&pactffi_sync_message_new, libpact_ffi, "pactffi_sync_message_new")
	purego.RegisterLibFunc(&pactffi_sync_message_delete, libpact_ffi, "pactffi_sync_message_delete")
	purego.RegisterLibFunc(&pactffi_sync_message_get_request_contents_str, libpact_ffi, "pactffi_sync_message_get_request_contents_str")
	purego.RegisterLibFunc(&pactffi_sync_message_set_request_contents_str, libpact_ffi, "pactffi_sync_message_set_request_contents_str")
	purego.RegisterLibFunc(&pactffi_sync_message_get_request_contents_length, libpact_ffi, "pactffi_sync_message_get_request_contents_length")
	purego.RegisterLibFunc(&pactffi_sync_message_get_request_contents_bin, libpact_ffi, "pactffi_sync_message_get_request_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_message_set_request_contents_bin, libpact_ffi, "pactffi_sync_message_set_request_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_message_get_request_contents, libpact_ffi, "pactffi_sync_message_get_request_contents")
	purego.RegisterLibFunc(&pactffi_sync_message_get_number_responses, libpact_ffi, "pactffi_sync_message_get_number_responses")
	purego.RegisterLibFunc(&pactffi_sync_message_get_response_contents_str, libpact_ffi, "pactffi_sync_message_get_response_contents_str")
	purego.RegisterLibFunc(&pactffi_sync_message_set_response_contents_str, libpact_ffi, "pactffi_sync_message_set_response_contents_str")
	purego.RegisterLibFunc(&pactffi_sync_message_get_response_contents_length, libpact_ffi, "pactffi_sync_message_get_response_contents_length")
	purego.RegisterLibFunc(&pactffi_sync_message_get_response_contents_bin, libpact_ffi, "pactffi_sync_message_get_response_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_message_set_response_contents_bin, libpact_ffi, "pactffi_sync_message_set_response_contents_bin")
	purego.RegisterLibFunc(&pactffi_sync_message_get_response_contents, libpact_ffi, "pactffi_sync_message_get_response_contents")
	purego.RegisterLibFunc(&pactffi_sync_message_get_description, libpact_ffi, "pactffi_sync_message_get_description")
	purego.RegisterLibFunc(&pactffi_sync_message_set_description, libpact_ffi, "pactffi_sync_message_set_description")
	purego.RegisterLibFunc(&pactffi_sync_message_get_provider_state, libpact_ffi, "pactffi_sync_message_get_provider_state")
	purego.RegisterLibFunc(&pactffi_sync_message_get_provider_state_iter, libpact_ffi, "pactffi_sync_message_get_provider_state_iter")
	purego.RegisterLibFunc(&pactffi_string_delete, libpact_ffi, "pactffi_string_delete")
	purego.RegisterLibFunc(&pactffi_create_mock_server, libpact_ffi, "pactffi_create_mock_server")
	purego.RegisterLibFunc(&pactffi_get_tls_ca_certificate, libpact_ffi, "pactffi_get_tls_ca_certificate")
	purego.RegisterLibFunc(&pactffi_create_mock_server_for_pact, libpact_ffi, "pactffi_create_mock_server_for_pact")
	purego.RegisterLibFunc(&pactffi_create_mock_server_for_transport, libpact_ffi, "pactffi_create_mock_server_for_transport")
	purego.RegisterLibFunc(&pactffi_mock_server_matched, libpact_ffi, "pactffi_mock_server_matched")
	purego.RegisterLibFunc(&pactffi_mock_server_mismatches, libpact_ffi, "pactffi_mock_server_mismatches")
	purego.RegisterLibFunc(&pactffi_cleanup_mock_server, libpact_ffi, "pactffi_cleanup_mock_server")
	purego.RegisterLibFunc(&pactffi_write_pact_file, libpact_ffi, "pactffi_write_pact_file")
	purego.RegisterLibFunc(&pactffi_mock_server_logs, libpact_ffi, "pactffi_mock_server_logs")
	purego.RegisterLibFunc(&pactffi_generate_datetime_string, libpact_ffi, "pactffi_generate_datetime_string")
	purego.RegisterLibFunc(&pactffi_check_regex, libpact_ffi, "pactffi_check_regex")
	purego.RegisterLibFunc(&pactffi_generate_regex_value, libpact_ffi, "pactffi_generate_regex_value")
	purego.RegisterLibFunc(&pactffi_free_string, libpact_ffi, "pactffi_free_string")
	purego.RegisterLibFunc(&pactffi_new_pact, libpact_ffi, "pactffi_new_pact")
	purego.RegisterLibFunc(&pactffi_pact_handle_to_pointer, libpact_ffi, "pactffi_pact_handle_to_pointer")
	purego.RegisterLibFunc(&pactffi_new_interaction, libpact_ffi, "pactffi_new_interaction")
	purego.RegisterLibFunc(&pactffi_new_message_interaction, libpact_ffi, "pactffi_new_message_interaction")
	purego.RegisterLibFunc(&pactffi_new_sync_message_interaction, libpact_ffi, "pactffi_new_sync_message_interaction")
	purego.RegisterLibFunc(&pactffi_upon_receiving, libpact_ffi, "pactffi_upon_receiving")
	purego.RegisterLibFunc(&pactffi_given, libpact_ffi, "pactffi_given")
	purego.RegisterLibFunc(&pactffi_interaction_test_name, libpact_ffi, "pactffi_interaction_test_name")
	purego.RegisterLibFunc(&pactffi_given_with_param, libpact_ffi, "pactffi_given_with_param")
	purego.RegisterLibFunc(&pactffi_given_with_params, libpact_ffi, "pactffi_given_with_params")
	purego.RegisterLibFunc(&pactffi_with_request, libpact_ffi, "pactffi_with_request")
	purego.RegisterLibFunc(&pactffi_with_query_parameter, libpact_ffi, "pactffi_with_query_parameter")
	purego.RegisterLibFunc(&pactffi_with_query_parameter_v2, libpact_ffi, "pactffi_with_query_parameter_v2")
	purego.RegisterLibFunc(&pactffi_with_specification, libpact_ffi, "pactffi_with_specification")
	purego.RegisterLibFunc(&pactffi_handle_get_pact_spec_version, libpact_ffi, "pactffi_handle_get_pact_spec_version")
	purego.RegisterLibFunc(&pactffi_with_pact_metadata, libpact_ffi, "pactffi_with_pact_metadata")
	purego.RegisterLibFunc(&pactffi_with_header, libpact_ffi, "pactffi_with_header")
	purego.RegisterLibFunc(&pactffi_with_header_v2, libpact_ffi, "pactffi_with_header_v2")
	purego.RegisterLibFunc(&pactffi_set_header, libpact_ffi, "pactffi_set_header")
	purego.RegisterLibFunc(&pactffi_response_status, libpact_ffi, "pactffi_response_status")
	purego.RegisterLibFunc(&pactffi_response_status_v2, libpact_ffi, "pactffi_response_status_v2")
	purego.RegisterLibFunc(&pactffi_with_body, libpact_ffi, "pactffi_with_body")
	purego.RegisterLibFunc(&pactffi_with_binary_body, libpact_ffi, "pactffi_with_binary_body")
	purego.RegisterLibFunc(&pactffi_with_binary_file, libpact_ffi, "pactffi_with_binary_file")
	purego.RegisterLibFunc(&pactffi_with_matching_rules, libpact_ffi, "pactffi_with_matching_rules")
	purego.RegisterLibFunc(&pactffi_with_multipart_file_v2, libpact_ffi, "pactffi_with_multipart_file_v2")
	purego.RegisterLibFunc(&pactffi_with_multipart_file, libpact_ffi, "pactffi_with_multipart_file")
	purego.RegisterLibFunc(&pactffi_pact_handle_get_message_iter, libpact_ffi, "pactffi_pact_handle_get_message_iter")
	purego.RegisterLibFunc(&pactffi_pact_handle_get_sync_message_iter, libpact_ffi, "pactffi_pact_handle_get_sync_message_iter")
	purego.RegisterLibFunc(&pactffi_pact_handle_get_sync_http_iter, libpact_ffi, "pactffi_pact_handle_get_sync_http_iter")
	purego.RegisterLibFunc(&pactffi_new_message_pact, libpact_ffi, "pactffi_new_message_pact")
	purego.RegisterLibFunc(&pactffi_new_message, libpact_ffi, "pactffi_new_message")
	purego.RegisterLibFunc(&pactffi_with_metadata, libpact_ffi, "pactffi_with_metadata")
	purego.RegisterLibFunc(&pactffi_message_expects_to_receive, libpact_ffi, "pactffi_message_expects_to_receive")
	purego.RegisterLibFunc(&pactffi_message_given, libpact_ffi, "pactffi_message_given")
	purego.RegisterLibFunc(&pactffi_message_given_with_param, libpact_ffi, "pactffi_message_given_with_param")
	purego.RegisterLibFunc(&pactffi_message_with_contents, libpact_ffi, "pactffi_message_with_contents")
	purego.RegisterLibFunc(&pactffi_message_with_metadata, libpact_ffi, "pactffi_message_with_metadata")
	purego.RegisterLibFunc(&pactffi_message_with_metadata_v2, libpact_ffi, "pactffi_message_with_metadata_v2")
	purego.RegisterLibFunc(&pactffi_message_reify, libpact_ffi, "pactffi_message_reify")
	purego.RegisterLibFunc(&pactffi_write_message_pact_file, libpact_ffi, "pactffi_write_message_pact_file")
	purego.RegisterLibFunc(&pactffi_with_message_pact_metadata, libpact_ffi, "pactffi_with_message_pact_metadata")
	purego.RegisterLibFunc(&pactffi_pact_handle_write_file, libpact_ffi, "pactffi_pact_handle_write_file")
	purego.RegisterLibFunc(&pactffi_new_async_message, libpact_ffi, "pactffi_new_async_message")
	purego.RegisterLibFunc(&pactffi_free_pact_handle, libpact_ffi, "pactffi_free_pact_handle")
	purego.RegisterLibFunc(&pactffi_free_message_pact_handle, libpact_ffi, "pactffi_free_message_pact_handle")
	purego.RegisterLibFunc(&pactffi_verify, libpact_ffi, "pactffi_verify")
	purego.RegisterLibFunc(&pactffi_verifier_new, libpact_ffi, "pactffi_verifier_new")
	purego.RegisterLibFunc(&pactffi_verifier_new_for_application, libpact_ffi, "pactffi_verifier_new_for_application")
	purego.RegisterLibFunc(&pactffi_verifier_shutdown, libpact_ffi, "pactffi_verifier_shutdown")
	purego.RegisterLibFunc(&pactffi_verifier_set_provider_info, libpact_ffi, "pactffi_verifier_set_provider_info")
	purego.RegisterLibFunc(&pactffi_verifier_add_provider_transport, libpact_ffi, "pactffi_verifier_add_provider_transport")
	purego.RegisterLibFunc(&pactffi_verifier_set_filter_info, libpact_ffi, "pactffi_verifier_set_filter_info")
	purego.RegisterLibFunc(&pactffi_verifier_set_provider_state, libpact_ffi, "pactffi_verifier_set_provider_state")
	purego.RegisterLibFunc(&pactffi_verifier_set_verification_options, libpact_ffi, "pactffi_verifier_set_verification_options")
	purego.RegisterLibFunc(&pactffi_verifier_set_coloured_output, libpact_ffi, "pactffi_verifier_set_coloured_output")
	purego.RegisterLibFunc(&pactffi_verifier_set_no_pacts_is_error, libpact_ffi, "pactffi_verifier_set_no_pacts_is_error")
	purego.RegisterLibFunc(&pactffi_verifier_set_publish_options, libpact_ffi, "pactffi_verifier_set_publish_options")
	purego.RegisterLibFunc(&pactffi_verifier_set_consumer_filters, libpact_ffi, "pactffi_verifier_set_consumer_filters")
	purego.RegisterLibFunc(&pactffi_verifier_add_custom_header, libpact_ffi, "pactffi_verifier_add_custom_header")
	purego.RegisterLibFunc(&pactffi_verifier_add_file_source, libpact_ffi, "pactffi_verifier_add_file_source")
	purego.RegisterLibFunc(&pactffi_verifier_add_directory_source, libpact_ffi, "pactffi_verifier_add_directory_source")
	purego.RegisterLibFunc(&pactffi_verifier_url_source, libpact_ffi, "pactffi_verifier_url_source")
	purego.RegisterLibFunc(&pactffi_verifier_broker_source, libpact_ffi, "pactffi_verifier_broker_source")
	purego.RegisterLibFunc(&pactffi_verifier_broker_source_with_selectors, libpact_ffi, "pactffi_verifier_broker_source_with_selectors")
	purego.RegisterLibFunc(&pactffi_verifier_execute, libpact_ffi, "pactffi_verifier_execute")
	purego.RegisterLibFunc(&pactffi_verifier_cli_args, libpact_ffi, "pactffi_verifier_cli_args")
	purego.RegisterLibFunc(&pactffi_verifier_logs, libpact_ffi, "pactffi_verifier_logs")
	purego.RegisterLibFunc(&pactffi_verifier_logs_for_provider, libpact_ffi, "pactffi_verifier_logs_for_provider")
	purego.RegisterLibFunc(&pactffi_verifier_output, libpact_ffi, "pactffi_verifier_output")
	purego.RegisterLibFunc(&pactffi_verifier_json, libpact_ffi, "pactffi_verifier_json")
	purego.RegisterLibFunc(&pactffi_using_plugin, libpact_ffi, "pactffi_using_plugin")
	purego.RegisterLibFunc(&pactffi_cleanup_plugins, libpact_ffi, "pactffi_cleanup_plugins")
	purego.RegisterLibFunc(&pactffi_interaction_contents, libpact_ffi, "pactffi_interaction_contents")
	purego.RegisterLibFunc(&pactffi_matches_string_value, libpact_ffi, "pactffi_matches_string_value")
	purego.RegisterLibFunc(&pactffi_matches_u64_value, libpact_ffi, "pactffi_matches_u64_value")
	purego.RegisterLibFunc(&pactffi_matches_i64_value, libpact_ffi, "pactffi_matches_i64_value")
	purego.RegisterLibFunc(&pactffi_matches_f64_value, libpact_ffi, "pactffi_matches_f64_value")
	purego.RegisterLibFunc(&pactffi_matches_bool_value, libpact_ffi, "pactffi_matches_bool_value")
	purego.RegisterLibFunc(&pactffi_matches_binary_value, libpact_ffi, "pactffi_matches_binary_value")
	purego.RegisterLibFunc(&pactffi_matches_json_value, libpact_ffi, "pactffi_matches_json_value")
}

// func main() {

// 	print(pactffi_version())
// 	pactffi_log_to_stdout(4)
// 	// pactffi_log_to_file("pact.log", 5)

// 	var verifier_handle = pactffi_verifier_new_for_application("pact-purego", "0.0.1")
// 	pactffi_verifier_set_provider_info(verifier_handle, "grpc-provider", "http", "localhost", 1234, "/")
// 	pactffi_verifier_add_file_source(verifier_handle, "no_pact.json")
// 	// pactffi_verifier_add_file_source(verifier_handle, "pact.json")
// 	// pactffi_verifier_add_file_source(verifier_handle, "../pacts/Consumer-Alice Service.json")
// 	// pactffi_verifier_add_file_source(verifier_handle, "../pacts/grpc-consumer-perl-area-calculator-provider.json")
// 	pactffi_verifier_execute(verifier_handle)
// 	pactffi_verifier_shutdown(verifier_handle)
// }

// hasSuffix tests whether the string s ends with suffix.
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// CString converts a go string to *byte that can be passed to C code.
func CString(name string) *byte {
	if hasSuffix(name, "\x00") {
		return &(*(*[]byte)(unsafe.Pointer(&name)))[0]
	}
	b := make([]byte, len(name)+1)
	copy(b, name)
	return &b[0]
}
