package grpc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/examples/grpc/routeguide"
	"github.com/pact-foundation/pact-go/v2/log"
	message "github.com/pact-foundation/pact-go/v2/message/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var dir, _ = os.Getwd()

func TestGrpcInteraction(t *testing.T) {
	p, _ := message.NewSynchronousPact(message.Config{
		Consumer: "grpcconsumer",
		Provider: "grpcprovider",
		PactDir:  filepath.ToSlash(fmt.Sprintf("%s/../pacts", dir)),
	})
	log.SetLogLevel("INFO")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/routeguide/route_guide.proto", dir)

	grpcInteraction := `{
		"pact:proto": "` + path + `",
		"pact:proto-service": "RouteGuide/GetFeature",
		"pact:content-type": "application/protobuf",
		"pact:protobuf-config": {
			"additionalIncludes": ["` + dir + `/routeguide"]
		},
		"request": {
			"stuff": {
				"id": "foo"
			}
		},
		"response": {
			"result_code": "RESULT_CODE_OK"
		}
	}`

	err := p.AddSynchronousMessage("Route guide - GetFeature").
		Given("feature 'Big Tree' exists").
		UsingPlugin(message.PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.4",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(transport message.TransportConfig, m message.SynchronousMessage) error {
			fmt.Println("gRPC transport running on", transport)

			// Establish the gRPC connection
			conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", transport.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				t.Fatal("unable to communicate to grpc server", err)
			}
			defer conn.Close()

			// Create the gRPC client
			c := routeguide_v2.NewRouteGuideClient(conn)

			feature_id := &routeguide_v2.IdentitySpec_Id{
				Id: "foo",
			}
			feature_spec := &routeguide_v2.IdentitySpec{
				Identity: feature_id,
			}
			point := &routeguide_v2.FeatureRequest{
				Stuff: feature_spec,
			}

			// Now we can make a normal gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			feature, err := c.GetFeature(ctx, point)

			if err != nil {
				t.Fatal(err.Error())
			}

			assert.Equal(t, routeguide_v2.ResultCode(1), feature.GetResultCode())

			return nil
		})

	assert.NoError(t, err)
}
