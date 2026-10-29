package native

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyVerifierFailure(t *testing.T) {
	mismatch := `{"interaction": "a request for a book", "mismatch": {"type": "mismatches", "mismatches": [{"type": "StatusMismatch"}], "interactionId": "1"}}`
	loadError := `{"interaction": "Failed to load pact", "mismatch": {"type": "error", "message": "Failed to load pact 'http://localhost/pacts/1' - task 5 panicked", "interactionId": ""}}`
	interactionError := `{"interaction": "a request for a pet", "mismatch": {"type": "error", "message": "Request to provider failed", "interactionId": "2"}}`

	tests := []struct {
		name     string
		json     string
		want     error
		contains string
	}{
		{name: "only mismatches", json: `{"result": false, "errors": [` + mismatch + `]}`, want: ErrVerifierFailed},
		{name: "no errors listed", json: `{"result": false, "errors": []}`, want: ErrVerifierFailed},
		{name: "pact failed to load", json: `{"result": false, "errors": [` + loadError + `]}`, want: ErrVerifierFailedToRun, contains: "task 5 panicked"},
		{name: "interaction could not be evaluated", json: `{"result": false, "errors": [` + interactionError + `]}`, want: ErrVerifierFailedToRun, contains: "Request to provider failed"},
		{name: "mismatch alongside a load error", json: `{"result": false, "errors": [` + mismatch + `, ` + loadError + `]}`, want: ErrVerifierFailedToRun, contains: "Failed to load pact"},
		{name: "unknown error type", json: `{"result": false, "errors": [{"interaction": "x", "mismatch": {"type": "something-new"}}]}`, want: ErrVerifierFailedToRun},
		{name: "empty result", json: "", want: ErrVerifierFailedToRun},
		{name: "unreadable result", json: "not json", want: ErrVerifierFailedToRun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyVerifierFailure(tt.json)

			require.ErrorIs(t, err, tt.want)
			assert.False(t, errors.Is(err, ErrVerifierFailed) && errors.Is(err, ErrVerifierFailedToRun), "errors must be mutually exclusive")
			if tt.contains != "" {
				assert.ErrorContains(t, err, tt.contains)
			}
		})
	}
}
