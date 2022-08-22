package v4

import (
	"fmt"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestAsyncTypeSystem(t *testing.T) {
	t.Skip()
	p, _ := NewAsynchronousPact(Config{
		Consumer: "asyncconsumer",
		Provider: "asyncprovider",
		PactDir:  "/tmp/",
	})
	log.SetLogLevel("TRACE")

	// type foo struct {
	// 	Foo string `json:"foo"`
	// }

	// // Sync - no plugin
	// err := p.AddAsynchronousMessage().
	// 	Given("some state").
	// 	Given("another state").
	// 	ExpectsToReceive("an important json message").
	// 	WithJSONContent(map[string]string{
	// 		"foo": "bar",
	// 	}).
	// 	AsType(&foo{}).
	// 	ConsumedBy(func(mc AsynchronousMessage) error {
	// 		fooMessage := mc.Body.(*foo)
	// 		assert.Equal(t, "bar", fooMessage.Foo)
	// 		return nil
	// 	}).
	// 	Verify(t)

	// assert.NoError(t, err)

	// Sync - with plugin, but no transport
	// TODO: ExecuteTest has been disabled for now, because it's not very useful
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

	// TODO: enable when there is a transport for async to test!
	err := p.AddAsynchronousMessage().
		Given("some state").
		ExpectsToReceive("some csv content").
		UsingPlugin(PluginConfig{
			Plugin:  "csv",
			Version: "0.0.1",
		}).
		WithContents(csvInteraction, "text/csv").
		// StartTransport("notarealtransport", "127.0.0.1", nil).
		ExecuteTest(t, func(m AsynchronousMessage) error {

			fmt.Println("Executing the CSV test", string(m.Contents))
			return nil
		})
	assert.NoError(t, err)
}
