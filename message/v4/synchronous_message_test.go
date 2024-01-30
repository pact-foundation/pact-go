package v4

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestSyncTypeSystem(t *testing.T) {
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	_ = log.SetLogLevel("TRACE")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/../../internal/native/pact_plugin.proto", dir)

	grpcInteraction := `{
		"pact:proto": "` + path + `",
		"pact:proto-service": "PactPlugin/InitPlugin",
		"pact:content-type": "application/protobuf",
		"request": {
			"implementation": "notEmpty('pact-go-driver')",
			"version": "matching(semver, '0.0.0')"	
		},
		"response": {
			"catalogue": [
				{
					"type": "INTERACTION",
					"key": "test"
				}
			]
		}
	}`

	// Sync - no plugin
	_ = p.AddSynchronousMessage("some description").
		Given("some state").
		WithRequest(func(r *SynchronousMessageWithRequestBuilder) {
			r.WithJSONContent(map[string]string{"foo": "bar"})
			r.WithMetadata(map[string]string{})
		}).
		WithResponse(func(r *SynchronousMessageWithResponseBuilder) {
			r.WithJSONContent(map[string]string{"foo": "bar"})
			r.WithMetadata(map[string]string{})
		}).
		ExecuteTest(t, func(m SynchronousMessage) error {
			// In this scenario, we have no real transport, so we need to mock/handle both directions

			// e.g. MQ use case -> write to a queue, get a response from another queue

			// What is the user expected to do here?
			// m.Request. // inbound -> send to queue
			// Poll the queue
			// m.Response // response from queue

			// User consumes the request

			return nil
		})

	// Sync - with plugin, but no transport
	csvInteraction := `{
		"request.path": "/reports/report002.csv",
		"response.status": "200",
		"response.contents": {
			"pact:content-type": "text/csv",
			"csvHeaders": true,
			"column:Name": "matching(type,'Name')",
			"column:Number": "matching(number,100)",
			"column:Date": "matching(datetime, 'yyyy-MM-dd','2000-01-01')"
		}
	}`

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	_ = p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "csv",
			Version: "0.0.1",
		}).
		WithContents(csvInteraction, "text/csv").
		ExecuteTest(t, func(m SynchronousMessage) error {
			fmt.Println("Executing the CSV test")
			return nil
		})

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	// Sync - with plugin + transport (pass)
	err := p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.13",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return nil
		})

	assert.Error(t, err)

	p, _ = NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})

	// Sync - with plugin + transport (fail)
	err = p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.13",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return errors.New("bad thing")
		})

	assert.Error(t, err)
}
