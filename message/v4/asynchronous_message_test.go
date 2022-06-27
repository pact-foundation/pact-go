package v4

import (
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestAsyncTypeSystem(t *testing.T) {
	p, _ := NewAsynchronousPact(Config{
		Consumer: "asyncconsumer",
		Provider: "asyncprovider",
		PactDir:  "/tmp/",
	})
	log.SetLogLevel("TRACE")

	type foo struct {
		Foo string `json:"foo"`
	}

	// Sync - no plugin
	err := p.AddAsynchronousMessage().
		Given("some state").
		Given("another state").
		ExpectsToReceive("an important json message").
		WithJSONContent(map[string]string{
			"foo": "bar",
		}).
		AsType(&foo{}).
		ConsumedBy(func(mc MessageContents) error {
			fooMessage := *mc.Content.(*foo)
			assert.Equal(t, "bar", fooMessage.Foo)
			return nil
		}).
		Verify(t)

	assert.NoError(t, err)

	// Sync - with plugin, but no transport
	// TODO: ExecuteTest has been disabled for now, because it's not very useful
	// csvInteraction := `{
	// 	"request.path", "/reports/report002.csv",
	// 	"response.status", "200",
	// 	"response.contents": {
	// 		"pact:content-type": "text/csv",                               // Set the content type to CSV
	// 		"csvHeaders": true,                                            // We have a header row
	// 		"column:Name": "matching(type,'Name')",                        // Column with header Name must match by type (which is actually useless with CSV)
	// 		"column:Number", "matching(number,100)",                       // Column with header Number must match a number format
	// 		"column:Date", "matching(datetime, 'yyyy-MM-dd','2000-01-01')" // Column with header Date must match an ISO format yyyy-MM-dd
	// 	}
	// }`

	// TODO: enable when there is a transport for async to test!
	// p.AddAsynchronousMessage().
	// 	Given("some state").
	// 	ExpectsToReceive("some csv content").
	// 	UsingPlugin(PluginConfig{
	// 		Plugin:  "csv",
	// 		Version: "0.0.1",
	// 	}).
	// 	WithContents(csvInteraction, "text/csv").
	// 	StartTransport("notarealtransport", "127.0.0.1", nil).
	// 	ExecuteTest(t, func(tc TransportConfig, m SynchronousMessage) error {
	// 		fmt.Println("Executing the CSV test")
	// 		return nil
	// 	})
}
