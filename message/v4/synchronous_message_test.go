package v4

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestSyncTypeSystem_NoPlugin(t *testing.T) {
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	// Sync - no plugin
	err := p.AddSynchronousMessage("some description").
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
	assert.NoError(t, err)
}

// Sync - with plugin, but no transport
func TestSyncTypeSystem_CsvPlugin_Matcher(t *testing.T) {
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})

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

	err := p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "csv",
			Version: "0.0.6",
		}).
		WithContents(csvInteraction, "text/csv").
		ExecuteTest(t, func(m SynchronousMessage) error {
			fmt.Println("Executing the CSV test")
			return nil
		})

	assert.NoError(t, err)
}
func TestSyncTypeSystem_ProtobufPlugin_Matcher_Transport(t *testing.T) {
	_ = log.SetLogLevel("INFO")
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/../../internal/native/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

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

	// Sync - with plugin + transport (pass)
	err := p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return nil
		})

	assert.Error(t, err)
	// assert.Equal(t, "Did not receive any requests for path 'PactPlugin/InitPlugin'", err)
	// TODO:- Work out why we get the following error message

	// Error
	// synchronous_message_test.go:130:
	// Error Trace:    /Users/saf/dev/pact-foundation/pact-go/message/v4/synchronous_message_test.go:130
	// Error:          Not equal:
	// 				expected: string("Did not receive any requests for path 'PactPlugin/InitPlugin'")
	// 				actual  : *errors.errorString(&errors.errorString{s:"pact validation failed: [{Request:{Method: Path:PactPlugin/InitPlugin Query: Headers:map[] Body:<nil>} Mismatches:[] Type:}]"})
	// Logs
	// 2024-07-04T01:36:33.745740Z DEBUG ThreadId(01) pact_plugin_driver::plugin_manager: Got response: ShutdownMockServerResponse { ok: false, results: [MockServerResult { path: "PactPlugin/InitPlugin", error: "Did not receive any requests for path 'PactPlugin/InitPlugin'", mismatches: [] }] }
}

// Sync - with plugin + transport (fail)
func TestSyncTypeSystem_ProtobufPlugin_Matcher_Transport_Fail(t *testing.T) {
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})
	_ = log.SetLogLevel("INFO")
	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/../../internal/native/pact_plugin.proto", strings.ReplaceAll(dir, "\\", "/"))

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

	err := p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "protobuf",
			Version: "0.3.15",
		}).
		WithContents(grpcInteraction, "application/protobuf").
		StartTransport("grpc", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig, m SynchronousMessage) error {
			fmt.Println("Executing a test - this is where you would normally make the gRPC call")

			return errors.New("bad thing")
		})

	assert.Error(t, err)
}
