//go:build consumer
// +build consumer

package protobuf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	message "github.com/pact-foundation/pact-go/v2/message/v4"
	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()

func TestPluginMessageConsumer(t *testing.T) {
	p, _ := message.NewAsynchronousPact(message.Config{
		Consumer: "protobufmessageconsumer",
		Provider: "protobufmessageprovider",
		PactDir:  filepath.ToSlash(fmt.Sprintf("%s/../pacts", dir)),
	})

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/../grpc/routeguide/route_guide.proto", dir)

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
			Version: "0.3.13",
		}).
		WithContents(protoMessage, "application/protobuf").
		ExecuteTest(t, func(m message.AsynchronousMessage) error {
			// TODO: normally would actually read/consume the message
			return nil
		})

	assert.NoError(t, err)
}
