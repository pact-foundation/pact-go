//go:build consumer

package protobuf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	message "github.com/pact-foundation/pact-go/v2/message/v4"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()

func TestPluginMessageConsumer(t *testing.T) {
	p, _ := message.NewAsynchronousPact(message.Config{
		Consumer: "protobufmessageconsumer",
		Provider: "protobufmessageprovider",
		PactDir:  filepath.ToSlash(dir + "/../pacts"),
	})

	dir, _ := os.Getwd()
	path := strings.ReplaceAll(dir, "\\", "/") + "/../grpc/routeguide/route_guide.proto"

	protoMessage := `{
		"pact:proto": "` + path + `",
		"pact:message-type": "Feature",
		"pact:content-type": "application/protobuf",

		"name": "notEmpty('Big Tree')",
		"location": {
			"latitude": "matching(number, 180)",
			"longitude": "matching(number, 200)"
		}
	}`

	err := p.AddAsynchronousMessage().
		Given("the world exists").
		ExpectsToReceive("feature message").
		UsingPlugin(message.PluginConfig{
			Plugin:  "protobuf",
			Version: "0.5.4",
		}).
		WithContents(protoMessage, "application/protobuf").
		ExecuteTest(t, func(m message.AsynchronousMessage) error {
			// TODO: normally would actually read/consume the message
			return nil
		})

	assert.NoError(t, err)
}
