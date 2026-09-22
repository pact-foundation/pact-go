package native

// Request is the sub-struct of Mismatch.
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    any               `json:"body,omitempty"`
}

// [
//   {
//     "method": "GET",
//     "mismatches": [
//       {
//         "actual": "",
//         "expected": "\"Bearer 1234\"",
//         "key": "Authorization",
//         "mismatch": "Expected header 'Authorization' but was missing",
//         "type": "HeaderMismatch"
//       }
//     ],
//     "path": "/foobar",
//     "type": "request-mismatch"
//   }
// ]

// MismatchDetail contains the specific assertions that failed during the verification.
type MismatchDetail struct {
	Actual   string `json:"actual"`
	Expected string `json:"expected"`
	Key      string `json:"key"`
	Mismatch string `json:"mismatch"`
	Type     string `json:"type"`
}

// MismatchedRequest contains details of any request mismatches during pact verification.
type MismatchedRequest struct {
	Request

	Mismatches []MismatchDetail `json:"mismatches"`
	Type       string           `json:"type"`
}
