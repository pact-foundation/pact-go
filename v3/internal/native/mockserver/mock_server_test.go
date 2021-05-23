package mockserver

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tmpPactFolder = "/var/tmp/pacts/"
var pactSimple = `{
  "consumer": {
    "name": "consumer"
  },
  "provider": {
    "name": "provider"
  },
  "interactions": [
    {
      "description": "Some name for the test",
      "request": {
        "method": "GET",
        "path": "/foobar"
      },
      "response": {
        "status": 200
      },
      "description": "Some name for the test",
      "provider_state": "Some state"
  }]
}`

var pactComplex = `{
  "consumer": {
    "name": "consumer"
  },
  "provider": {
    "name": "provider"
  },
  "interactions": [
    {
    "request": {
      "method": "GET",
      "path": "/foobar",
      "body": {
        "pass": 1234,
        "user": {
          "address": "some address",
          "name": "someusername",
          "phone": 12345678,
          "plaintext": "plaintext"
        }
      }
    },
    "response": {
      "status": 200
    },
    "description": "Some name for the test",
    "provider_state": "Some state",
    "matchingRules": {
      "$.body.pass": {
        "match": "regex",
        "regex": "\\d+"
      },
      "$.body.user.address": {
        "match": "regex",
        "regex": "\\s+"
      },
      "$.body.user.name": {
        "match": "regex",
        "regex": "\\s+"
      },
      "$.body.user.phone": {
        "match": "regex",
        "regex": "\\d+"
      }
    }
  }]
}`

func TestMockServer_CreateAndCleanupMockServer(t *testing.T) {
	m := MockServer{}
	Init()
	port, _ := m.CreateMockServer(pactComplex, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	if port <= 0 {
		t.Fatal("want port > 0, got", port)
	}
}

func TestMockServer_MismatchesSuccess(t *testing.T) {
	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", port))
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("want '200', got '%d'", res.StatusCode)
	}

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 0 {
		t.Fatalf("want 0 mismatches, got '%d'", len(mismatches))
	}
}

func TestMockServer_MismatchesFail(t *testing.T) {
	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 1 {
		t.Fatalf("want 1 mismatch, got '%d'", len(mismatches))
	}
}

func TestMockServer_VerifySuccess(t *testing.T) {
	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", port))
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}

	success, mismatches := m.Verify(port, tmpPactFolder)
	if !success {
		t.Fatalf("want 'true' but got '%v'", success)
	}

	if len(mismatches) != 0 {
		t.Fatalf("want 0 mismatches, got '%d'", len(mismatches))
	}
}

func TestMockServer_VerifyFail(t *testing.T) {
	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)

	success, mismatches := m.Verify(port, tmpPactFolder)
	if success {
		t.Fatalf("want 'false' but got '%v'", success)
	}

	if len(mismatches) != 1 {
		t.Fatalf("want 1 mismatch, got '%d'", len(mismatches))
	}
}

func TestMockServer_WritePactfile(t *testing.T) {
	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", port))
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	err = m.WritePactFile(port, tmpPactFolder)

	if err != nil {
		t.Fatal("error: ", err)
	}
}

func TestMockServer_GetTLSConfig(t *testing.T) {
	config := GetTLSConfig()

	fmt.Println("tls config", config)
}

func TestVersion(t *testing.T) {
	m := MockServer{}
	fmt.Println("version: ", m.Version())
}

func TestHandleBasedHTTPTests(t *testing.T) {
	Init()
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	m := NewHTTPMockServer("test-http-consumer", "test-http-provider")

	fmt.Println("pact struct:", m)

	i := m.NewInteraction("some interaction")
	fmt.Println("pact interaction:", i)

	i.UponReceiving("some interaction").
		Given("some state").
		WithRequest("GET", "/products").
		// withRequestHeader("x-special-header", 0, "header")
		// withQuery("someParam", 0, "someValue")
		WithJSONResponseBody(`{
	  "name": {
      "pact:matcher:type": "type",
      "value": "some name"
    },
	  "age": 23,
	  "alive": true
	}`).
		// withResponseHeader(i, "x-special-header", 0, "header")
		WithStatus(200)

	// // Start the mock service
	// const host = "127.0.0.1"
	port, err := m.Start("0.0.0.0:0", false)
	assert.NoError(t, err)
	defer m.CleanupMockServer(port)

	r, err := http.Get(fmt.Sprintf("http://0.0.0.0:%d/products", port))
	assert.NoError(t, err)

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 0 {
		t.Fatalf("want 0 mismatches, got '%d'", len(mismatches))
	}

	err = m.WritePactFile(port, tmpPactFolder)
	assert.NoError(t, err)

	fmt.Println(r)
}

func TestHandleBasedMessageTestsWithString(t *testing.T) {
	Init()
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
	Init()
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
		WithJSONContents([]byte(`{"some": "json"}`))

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
	Init()
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	s := NewMessageServer("test-binarymessage-consumer", "test-binarymessage-provider").
		WithMetadata("some-namespace", "the-key", "the-value")

	// generate some binary data
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err = zw.Write([]byte("A long time ago in a galaxy far, far away..."))
	assert.NoError(t, err)
	encodedMessage := "H4sIABqQqWAAA3NUyMnPS1coycxNVUhMz1fIzFNIVEhPzEmsqFRISyzSAREKieWJlXp6elwAfbWPLy0AAAA="

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
	log.Println(body)

	var res binaryMessage
	err = json.Unmarshal([]byte(body), &res)
	assert.NoError(t, err)

	assert.Equal(t, "some binary message", res.Description)
	assert.Equal(t, encodedMessage, res.Contents)
	assert.Len(t, res.ProviderStates, 2)
	assert.NotEmpty(t, res.Contents)

	// This is where you would invoke the real function with the message

	err = s.WritePactFile(tmpPactFolder, false)
	assert.NoError(t, err)
}

type binaryMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       []byte                   `json:"contents"`
}
type jsonMessage struct {
	ProviderStates []map[string]interface{} `json:"providerStates"`
	Description    string                   `json:"description"`
	Metadata       map[string]string        `json:"metadata"`
	Contents       interface{}              `json:"contents"`
}

// type messageContent interface {
// 	getStringBody() (string, error)
// 	getObjectBody() (interface{}, error)
// 	getBytesBody() ([]byte, error)
// }

// type binaryMessageContents struct {
// 	Contents []byte `json:"contents"`
// }
// type objectMessageContents struct {
// 	Contents interface{} `json:"contents"`
// }
