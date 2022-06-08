package v4

import (
	"testing"
)

func TestTypeSystem(t *testing.T) {
	p, _ := NewSynchronousPact(Config{
		Consumer: "consumer",
		Provider: "provider",
		PactDir:  "/tmp/",
	})

	p.AddSynchronousMessage("some description").
		Given("some state").
		WithRequest(func(r *SynchronousMessageWithRequestBuilder) {
			r.WithJSONContent("")
			r.AsType("")
			r.WithMetadata(map[string]string{})
		}).
		WithResponse(func(r *SynchronousMessageWithResponseBuilder) {
			r.WithJSONContent("")
			r.AsType("")
			r.WithMetadata(map[string]string{})
		}).
		ExecuteTest(t, func(t TransportConfig) error {
			return nil
		})

	p.AddSynchronousMessage("some description").
		Given("some state").
		UsingPlugin(PluginConfig{
			Plugin:  "some plugin",
			Version: "1.0.0",
		}).
		WithContents().
		StartTransport(). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(t TransportConfig) error {
			return nil
		})

}
