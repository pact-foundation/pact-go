//go:build consumer

package avro

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAvroHTTP(t *testing.T) {
	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "AvroConsumer",
		Provider: "AvroProvider",
		PactDir:  filepath.ToSlash(dir + "/../pacts"),
	})
	require.NoError(t, err)

	dir, _ := os.Getwd()
	path := strings.ReplaceAll(dir, "\\", "/") + "/user.avsc"

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
			Version: "0.0.6",
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
		Method: http.MethodGet,
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
		return nil, fmt.Errorf("calling the avro service: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading the avro service response: %w", err)
	}

	codec := getCodec()
	native, _, err := codec.NativeFromBinary(bytes)
	if err != nil {
		return nil, fmt.Errorf("decoding the avro response: %w", err)
	}

	fields, ok := native.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected avro record to decode to map[string]interface{}, got %T", native)
	}

	id, ok := fields["id"].(int64)
	if !ok {
		return nil, fmt.Errorf("expected field \"id\" to be int64, got %T", fields["id"])
	}

	username, ok := fields["username"].(string)
	if !ok {
		return nil, fmt.Errorf("expected field \"username\" to be string, got %T", fields["username"])
	}

	user := &User{
		ID:       id,
		Username: username,
	}

	return user, nil
}
