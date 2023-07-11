//go:build consumer
// +build consumer

package avro

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"

	"path/filepath"

	"github.com/stretchr/testify/assert"
)

var dir, _ = os.Getwd()

func TestAvroHTTP(t *testing.T) {
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "AvroConsumer",
		Provider: "AvroProvider",
		PactDir:  filepath.ToSlash(fmt.Sprintf("%s/../pacts", dir)),
	})
	assert.NoError(t, err)

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/user.avsc", dir)

	avroResponse := `{
		"pact:avro": "` + path + `",
		"pact:record-name": "User",
		"pact:content-type": "avro/binary",
		"id": "matching(number, 1)",
		"username": "notEmpty('matt')"
	}`

	// Set up our expected interactions.
	err = mockProvider.
		AddInteraction().
		UponReceiving("A request to do get some Avro stuff").
		UsingPlugin(consumer.PluginConfig{
			Plugin:  "avro",
			Version: "0.0.2",
		}).
		WithRequest("GET", "/avro").
		WillRespondWith(200, func(res *consumer.V4InteractionWithPluginResponseBuilder) {
			res.PluginContents("avro/binary", avroResponse)
		}).
		ExecuteTest(t, func(msc consumer.MockServerConfig) error {
			resp, err := callServiceHTTP(msc)

			assert.Equal(t, int64(1), resp.ID)
			assert.Equal(t, "matt", resp.Username) // ??????!

			return err
		})
	assert.NoError(t, err)
}

func callServiceHTTP(msc consumer.MockServerConfig) (*User, error) {
	client := &http.Client{}
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Host:   fmt.Sprintf("%s:%d", msc.Host, msc.Port),
			Scheme: "http",
			Path:   "/avro",
		},
		Header: make(http.Header),
	}

	req.Header.Set("Content-Type", "avro/binary;record=User")

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	codec := getCodec()
	native, _, err := codec.NativeFromBinary(bytes)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:       native.(map[string]interface{})["id"].(int64),
		Username: native.(map[string]interface{})["username"].(string),
	}

	return user, err
}
