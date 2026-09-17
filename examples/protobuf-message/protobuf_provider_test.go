//go:build provider
// +build provider

package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide"
	pactlog "github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/message"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	pactversion "github.com/pact-foundation/pact-go/v2/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPluginMessageProvider(t *testing.T) {
	dir, _ := os.Getwd()
	pactDir := fmt.Sprintf("%s/../pacts", dir)

	err := pactlog.SetLogLevel("INFO")
	require.NoError(t, err)

	pactversion.CheckVersion()

	verifier := provider.NewVerifier()

	functionMappings := message.Handlers{
		"feature message": func([]models.ProviderState) (message.Body, message.Metadata, error) {
			fmt.Println("feature message handler")
			feature, _ := proto.Marshal(&routeguide.Feature{
				Name: "fake feature",
				Location: &routeguide.Point{
					Latitude:  int32(1),
					Longitude: int32(1),
				},
			})
			return feature, message.Metadata{
				"contentType": "application/protobuf;message=.routeguide.Feature", // <- This is required to ensure the correct type is matched
			}, nil
		},
	}

	err = verifier.VerifyProvider(t, provider.VerifyRequest{
		PactFiles: []string{
			filepath.ToSlash(fmt.Sprintf("%s/protobufmessageconsumer-protobufmessageprovider.json", pactDir)),
		},
		Provider:        "protobufmessageprovider",
		MessageHandlers: functionMappings,
	})

	assert.NoError(t, err)
}
