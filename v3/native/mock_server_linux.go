package native

/*
#cgo LDFLAGS: ${SRCDIR}/../../libs/libpact_mock_server.dylib

// Library headers
int create_mock_server(char* pact, int port);
*/
import "C"
import "fmt"

// CreateMockServer creates a new Mock Server from a given Pact file
func CreateMockServer(pact string) int {
	res := C.create_mock_server(C.CString(pact), 0)
	fmt.Println("Mock Server running on port:", res)
	return int(res)
}
