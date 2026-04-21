package native

/*
#include "pact.h"
*/
import "C"

// interactionGiven wraps pactffi_given so that it is only referenced from a
// single CGo compilation unit. Go 1.24+ performs cross-file consistency checks
// for CGo symbols, and having the same C function (especially one returning
// bool) referenced from multiple files can trigger "inconsistent definitions"
// errors. By centralising the call here, mock_server.go and message_server.go
// can both invoke this Go-level wrapper without directly referencing the C
// symbol.
func interactionGiven(handle C.InteractionHandle, state string) {
	cState := C.CString(state)
	defer free(cState)

	C.pactffi_given(handle, cState)
}

// interactionGivenWithParams wraps pactffi_given_with_param for the same
// reason as interactionGiven above. It handles the full params map in one call
// so that the state C string is allocated only once, matching the performance
// of the original inline implementation.
func interactionGivenWithParams(handle C.InteractionHandle, state string, params map[string]interface{}) {
	cState := C.CString(state)
	defer free(cState)

	for k, v := range params {
		cKey := C.CString(k)
		cValue := C.CString(stringFromInterface(v))

		C.pactffi_given_with_param(handle, cState, cKey, cValue)

		free(cValue)
		free(cKey)
	}
}
