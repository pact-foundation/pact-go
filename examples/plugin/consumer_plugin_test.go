//go:build consumer
// +build consumer

package plugin

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"

	// "github.com/pact-foundation/pact-go/v2/matchers"
	message "github.com/pact-foundation/pact-go/v2/message/v4"
	"github.com/stretchr/testify/assert"
)

func TestHTTPPlugin(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "MattConsumer",
		Provider: "MattProvider",
	})
	assert.NoError(t, err)

	// MATT is a protocol, where all message start and end with a MATT
	mattRequest := `{"request": {"body": "hello"}}`
	mattResponse := `{"response":{"body":"world"}}`

	// Set up our expected interactions.
	err = mockProvider.
		AddInteraction().
		UponReceiving("A request to do a matt").
		UsingPlugin(consumer.PluginConfig{
			Plugin:  "matt",
			Version: "0.0.1",
		}).
		WithRequest("POST", "/matt", func(req *consumer.V4InteractionWithPluginRequestBuilder) {
			req.PluginContents("application/matt", mattRequest)
		}).
		WillRespondWith(200, func(res *consumer.V4InteractionWithPluginResponseBuilder) {
			res.PluginContents("application/matt", mattResponse)
		}).
		ExecuteTest(t, func(msc consumer.MockServerConfig) error {
			resp, err := callMattServiceHTTP(msc, "hello")

			assert.Equal(t, "world", resp)

			return err
		})
	assert.NoError(t, err)
}

func TestTCPInteraction(t *testing.T) {
	p, _ := message.NewSynchronousPact(message.Config{
		Consumer: "matttcpconsumer",
		Provider: "matttcpprovider",
	})

	// MATT is a protocol, where all message start and end with a MATT
	mattMessage := `{"request": {"body": "hellotcp"}, "response":{"body":"tcpworld"}}`

	err := p.AddSynchronousMessage("Matt message").
		Given("the world exists").
		UsingPlugin(message.PluginConfig{
			Plugin:  "matt",
			Version: "0.0.1",
		}).
		WithContents(mattMessage, "application/matt").
		StartTransport("matt", "127.0.0.1", nil). // For plugin tests, we can't assume if a transport is needed, so this is optional
		ExecuteTest(t, func(transport message.TransportConfig, m message.SynchronousMessage) error {
			fmt.Println("matt TCP transport running on", transport)

			str, err := callMattServiceTCP(transport, "hellotcp!")

			assert.Equal(t, "tcpworld", str)
			return err
		})

	assert.NoError(t, err)
}

func callMattServiceHTTP(msc consumer.MockServerConfig, message string) (string, error) {
	client := &http.Client{}
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Host:   fmt.Sprintf("%s:%d", msc.Host, msc.Port),
			Scheme: "http",
			Path:   "/matt",
		},
		Body:   ioutil.NopCloser(strings.NewReader(generateMattMessage(message))),
		Header: make(http.Header),
	}

	req.Header.Set("Content-Type", "application/matt")

	res, err := client.Do(req)

	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return parseMattMessage(string(bytes)), err
}

func callMattServiceTCP(transport message.TransportConfig, message string) (string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", transport.Address, transport.Port))
	if err != nil {
		return "", err
	}

	conn.Write([]byte(generateMattMessage(message)))

	str, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil {
		return "", err
	}

	return parseMattMessage(str), nil
}

func generateMattMessage(message string) string {
	return fmt.Sprintf("MATT%sMATT\n", message)
}

func parseMattMessage(message string) string {
	return strings.TrimSpace(strings.ReplaceAll(message, "MATT", ""))
}
