package v3

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

var pactFileWithInteractions = newPactFileV3("consumer", "provider", []*InteractionV3{
	{
		Interaction{
			Request: Request{
				Method: "GET",
				Path:   String("/foo"),
			},
			Response: Response{
				Status: 200,
			},
			Description: "some interaction",
		},
		[]ProviderStateV3{},
	},
}, nil)

var pactFileWithInteractions2 = newPactFileV3("consumer", "provider", []*InteractionV3{
	{
		Interaction{
			Request: Request{
				Method: "GET",
				Path:   String("/bar"),
			},
			Response: Response{
				Status: 400,
			},
			Description: "another interaction",
		},
		[]ProviderStateV3{},
	},
}, nil)

var pactFileWithMessages = newPactFileV3("consumer", "provider", nil, []*Message{
	{
		Description: "some event",
		Content:     nil,
	},
})

func TestPactFileV3ReaderWriter(t *testing.T) {
	mockFs := afero.NewMemMapFs()

	p := &pactFileV3ReaderWriter{
		fs: mockFs,
	}

	t.Run("write", func(t *testing.T) {
		err := p.write("/tmp/", pactFileWithInteractions)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("update", func(t *testing.T) {
		err := p.update("/tmp/", pactFileWithInteractions2)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("read", func(t *testing.T) {
		f, err := p.read("/tmp/consumer-provider.json")
		if err != nil {
			t.Error(err)
		}
		assert.Len(t, f.Interactions, 2)
	})
}

func TestMergePactFiles(t *testing.T) {
	tests := []struct {
		name    string
		orig    pactFileV3
		updated pactFileV3
		want    pactFileV3
		wantErr bool
	}{
		{
			name:    "pact file with interactions merged with interactions",
			wantErr: false,
			orig:    pactFileWithInteractions,
			updated: pactFileWithInteractions,
		},
		{
			name:    "pact file with messages merged with messages",
			wantErr: false,
			orig:    pactFileWithMessages,
			updated: pactFileWithMessages,
		},
		{
			name:    "error case - pact file with messages merged with interactions",
			wantErr: true,
			orig:    pactFileWithMessages,
			updated: pactFileWithInteractions,
		},
		{
			name:    "error case - pact file with interactions merged with messages",
			wantErr: true,
			orig:    pactFileWithInteractions,
			updated: pactFileWithMessages,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mergePactFiles(tt.orig, tt.updated)
			if tt.wantErr {
				assert.Error(t, err, "wanted error for %s but did not get one", tt.name)
			}
		})
	}
}

func TestPactFilePath(t *testing.T) {
	assert.Equal(t, "/tmp/consumer-provider.json", pactFilePath("/tmp/", pactFileWithInteractions))
}
