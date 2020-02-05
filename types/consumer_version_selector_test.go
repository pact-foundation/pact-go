package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumerVersionSelectorValidate(t *testing.T) {
	tests := []struct {
		name     string
		selector ConsumerVersionSelector
		err      bool
	}{
		{name: "no pacticipant", selector: ConsumerVersionSelector{}, err: true},
		{name: "pacticipant only", selector: ConsumerVersionSelector{Pacticipant: "foo"}, err: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.selector.Validate()
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
