package v2

import (
	"encoding/json"
	"fmt"

	"github.com/pact-foundation/pact-go/v2/compatibility-suite/model"
)

type ffiVerifierResult struct {
	Errors []struct {
		Interaction string `json:"interaction"`
		Mismatch    struct {
			Type       string `json:"type"`
			Message    string `json:"message"`
			Mismatches []struct {
				Type string `json:"type"`
			} `json:"mismatches"`
		} `json:"mismatch"`
	} `json:"errors"`
}

// parseVerifierResult reads the failures out of pact_ffi's verification
// result (pactffi_verifier_json).
func parseVerifierResult(data []byte) ([]model.VerificationError, error) {
	var result ffiVerifierResult

	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("parsing the verifier result: %w", err)
	}

	out := make([]model.VerificationError, 0, len(result.Errors))
	for _, e := range result.Errors {
		failure := model.VerificationError{Interaction: e.Interaction, Message: e.Mismatch.Message}
		for _, m := range e.Mismatch.Mismatches {
			failure.MismatchTypes = append(failure.MismatchTypes, m.Type)
		}
		out = append(out, failure)
	}

	return out, nil
}
