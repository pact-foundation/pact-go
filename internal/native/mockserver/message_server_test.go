package mockserver

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleBasedMessageTestsWithString(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	s := NewMessageServer("test-message-consumer", "test-message-provider")

	m := s.NewMessage().
		Given("some state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithContents("text/plain", []byte("some string"))

	body := m.ReifyMessage()

	var res jsonMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	assert.Equal(t, res.Description, "some message")
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithJSON(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	s := NewMessageServer("test-message-consumer", "test-message-provider")

	m := s.NewMessage().
		Given("some state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithJSONContents(map[string]string{
			"some": "json",
		})

	body := m.ReifyMessage()
	log.Println(body) // TODO: JSON is not stringified - probably should be?

	var res jsonMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	assert.Equal(t, res.Description, "some message")
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

func TestHandleBasedMessageTestsWithBinary(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	s := NewMessageServer("test-binarymessage-consumer", "test-binarymessage-provider").
		WithMetadata("some-namespace", "the-key", "the-value")

	// generate some binary data
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	encodedMessage := "A long time ago in a galaxy far, far away..."
	_, err = zw.Write([]byte(encodedMessage))
	assert.NoError(t, err)

	err = zw.Close()
	assert.NoError(t, err)

	m := s.NewMessage().
		Given("some binary state").
		GivenWithParameter("param", map[string]interface{}{
			"foo": "bar",
		}).
		ExpectsToReceive("some binary message").
		WithMetadata(map[string]string{
			"meta": "data",
		}).
		WithBinaryContents(buf.Bytes())

	body := m.ReifyMessage()

	// Check the reified message is good

	var res binaryMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	// Extract binary payload, base 64 decode it, unzip it
	data, err := base64.RawStdEncoding.DecodeString(res.Contents)
	assert.NoError(t, err)
	r, err := gzip.NewReader(bytes.NewReader(data))
	assert.NoError(t, err)
	result, _ := ioutil.ReadAll(r)

	assert.Equal(t, encodedMessage, string(result))
	assert.Equal(t, "some binary message", res.Description)
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message...

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

type binaryMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       string                   `json:"contents"` // base 64 encoded
	// Contents       []byte                   `json:"contents"`
}
type jsonMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       interface{}              `json:"contents"`
}
