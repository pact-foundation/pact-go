package provider

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
		{name: "no pacticipant", selector: ConsumerVersionSelector{}, err: false},
		{name: "no pacticipant and all set", selector: ConsumerVersionSelector{All: true}, err: true},
		{name: "all and latest set", selector: ConsumerVersionSelector{All: true}, err: true},
		{name: "pacticipant only", selector: ConsumerVersionSelector{Pacticipant: "foo"}, err: true},
		{name: "pacticipant and tag", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo"}, err: false},
		{name: "pacticipant, tag and all set", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", All: true}, err: false},
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
