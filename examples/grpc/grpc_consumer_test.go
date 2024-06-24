//go:build consumer
// +build consumer

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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGetFeatureSuccess(t *testing.T) {
	p, _ := message.NewSynchronousPact(message.Config{
		Consumer: "grpcconsumer",
		Provider: "grpcprovider",
		PactDir:  filepath.ToSlash(fmt.Sprintf("%s/../pacts", dir)),
	})
	log.SetLogLevel("DEBUG")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/routeguide/route_guide.proto", dir)

	grpcInteraction := `{
		"pact:proto": "` + path + `",
		"pact:proto-service": "RouteGuide/GetFeature",
		"pact:content-type": "application/protobuf",
		"request": {
			"latitude": "matching(number, 180)",
			"longitude": "matching(number, 200)"
		},
		"response": {
			"name": "notEmpty('Big Tree')",
			"location": {
				"latitude": "matching(number, 180)",
				"longitude": "matching(number, 200)"
			}
		}
	}`

	err := p.AddSynchronousMessage("Route guide - GetFeature").
		Given("feature 'Big Tree' exists").
		UsingPlugin(message.PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
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
			c := routeguide.NewRouteGuideClient(conn)

			point := &routeguide.Point{
				Latitude:  180,
				Longitude: 200,
			}

			// Now we can make a normal gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			feature, err := c.GetFeature(ctx, point)

			if err != nil {
				t.Fatal(err.Error())
			}

			feature.GetLocation()
			assert.Equal(t, "Big Tree", feature.GetName())
			assert.Equal(t, int32(180), feature.GetLocation().GetLatitude())

			return nil
		})

	assert.NoError(t, err)
}

func TestGetFeatureError(t *testing.T) {
	log.SetLogLevel("DEBUG")
	p, _ := message.NewSynchronousPact(message.Config{
		Consumer: "grpcconsumer",
		Provider: "grpcprovider",
		PactDir:  filepath.ToSlash(fmt.Sprintf("%s/../pacts", dir)),
	})

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/routeguide/route_guide.proto", dir)

	grpcInteraction := `{
		"pact:proto": "` + path + `",
		"pact:proto-service": "RouteGuide/GetFeature",
		"pact:content-type": "application/protobuf",
		"request": {
			"latitude": "matching(number, -1)",
			"longitude": "matching(number, -1)"
		},
		"responseMetadata": {
			"grpc-status": "NOT_FOUND",
			"grpc-message": "matching(type, 'no feature was found at latitude:-1  longitude:-1')"
		}
	}`

	err := p.AddSynchronousMessage("Route guide - GetFeature - error response").
		Given("feature does not exist at -1, -1").
		UsingPlugin(message.PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(transport message.TransportConfig, m message.SynchronousMessage) error {
			fmt.Println("gRPC transport running on", transport)

			// Establish the gRPC connection
			conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", transport.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer conn.Close()

			// Create the gRPC client
			c := routeguide.NewRouteGuideClient(conn)

			point := &routeguide.Point{
				Latitude:  -1,
				Longitude: -1,
			}

			// Now we can make a normal gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err = c.GetFeature(ctx, point)

			require.Error(t, err)
			// TODO: uncomment once new FFI and new pact-protobuf plugin are released with a fix
			//		 https://github.com/pact-foundation/pact-reference/commit/29b326e59b48a6a78a019b37e378b7742c728da5
			// require.ErrorContains(t, err, "no feature was found at latitude:-1  longitude:-1")

			return nil
		})

	assert.NoError(t, err)
}

func TestSaveFeature(t *testing.T) {
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
		"pact:proto-service": "RouteGuide/SaveFeature",
		"pact:content-type": "application/protobuf",
		"request": {
			"name": "notEmpty('A shed')",
			"location": {
				"latitude": "matching(number, 99)",
				"longitude": "matching(number, 99)"
			}
		},
		"response": {
			"name": "notEmpty('A shed')",
			"location": {
				"latitude": "matching(number, 99)",
				"longitude": "matching(number, 99)"
			}
		}
	}`

	err := p.AddSynchronousMessage("Route guide - SaveFeature").
		Given("feature does not exist at -1, -1").
		UsingPlugin(message.PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
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
			c := routeguide.NewRouteGuideClient(conn)
			feature := &routeguide.Feature{
				Name: "A shed",
				Location: &routeguide.Point{
					Latitude:  99,
					Longitude: 99,
				},
			}

			// Now we can make a normal gRPC request
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			response, err := c.SaveFeature(ctx, feature)

			if err != nil {
				t.Fatal(err.Error())
			}

			assert.Equal(t, feature.GetName(), response.GetName())
			assert.Equal(t, feature.GetLocation().GetLatitude(), feature.GetLocation().GetLatitude())

			return nil
		})

	assert.NoError(t, err)
}
