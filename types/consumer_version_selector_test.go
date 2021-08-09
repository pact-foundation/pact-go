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
		{name: "no pacticipant", selector: ConsumerVersionSelector{}, err: false},
		{name: "no pacticipant and all set", selector: ConsumerVersionSelector{All: true}, err: false},
		{name: "all and latest set", selector: ConsumerVersionSelector{All: true, Latest: true}, err: true},
		{name: "pacticipant only", selector: ConsumerVersionSelector{Pacticipant: "foo"}, err: false},
		{name: "pacticipant and tag", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo"}, err: false},
		{name: "pacticipant, tag and all set", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", All: true}, err: false},
		{name: "pacticipant and consumer", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", Consumer: "bar"}, err: true},
		{name: "pacticipant, tag, consumer", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo"}, err: false},
		{name: "pacticipant, tag, deployedOrReleased", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", DeployedOrReleased: true}, err: false},
		{name: "pacticipant, tag, deployed", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", Deployed: true}, err: false},
		{name: "pacticipant, tag, released", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", Released: true}, err: false},
		{name: "pacticipant, tag, environment", selector: ConsumerVersionSelector{Pacticipant: "foo", Tag: "foo", Environment: "dev"}, err: false},
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
