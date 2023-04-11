package native

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func init() {
	Init("")
}

func TestMockServer_CreateAndCleanupMockServer(t *testing.T) {
	m := MockServer{}
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
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	_, err = http.Get(fmt.Sprintf("http://localhost:%d/foobar", port))
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
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
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
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)

	m := MockServer{}
	port, _ := m.CreateMockServer(pactSimple, "0.0.0.0:0", false)
	defer m.CleanupMockServer(port)

	_, err = http.Get(fmt.Sprintf("http://localhost:%d/foobar", port))
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

	t.Log("tls config", config)
}

func TestVersion(t *testing.T) {
	t.Log("version: ", Version())
}

func TestHandleBasedHTTPTests(t *testing.T) {
	tmpPactFolder := "/tmp/"

	for attempt := 0; attempt < 10; attempt++ {
		t.Log("Running attempt", attempt, "============> \n\n")
		for idx := 0; idx < 2; idx++ {
			m := NewHTTPPact(fmt.Sprintf("test-http-consumer%d", idx), "test-http-provider")
			m.WithSpecificationVersion(SPECIFICATION_VERSION_V4)

			i := m.NewInteraction("some interaction")

			i.UponReceiving("some interaction").
				WithRequest("POST", "/products").
				WithRequestHeaders(map[string][]interface{}{
					"content-type": {
						"application/json",
					},
				}).
				WithStatus(200)

			// // Start the mock service
			port, err := m.Start("0.0.0.0:0", false)
			assert.NoError(t, err)
			defer m.CleanupMockServer(port)

			_, err = http.Post(fmt.Sprintf("http://0.0.0.0:%d/products", port), "application/json", nil)
			assert.NoError(t, err)

			mismatches := m.MockServerMismatchedRequests(port)
			if len(mismatches) != 0 {
				t.Fatalf("want 0 mismatches, got '%d'", len(mismatches))
			}

			err = m.WritePactFile(port, tmpPactFolder)
			assert.NoError(t, err)
		}
		cmd := exec.Command("diff", "/tmp/test-http-consumer0-test-http-provider.json", "/tmp/test-http-consumer1-test-http-provider.json")

		var stdout bytes.Buffer
		cmd.Stdout = &stdout

		cmd.Run()

		if strings.Contains(string(stdout.String()), "key") {
			t.Log("Interaction keys changed: ")
			t.Log(string(stdout.String()))
		}

		err := os.Remove("/tmp/test-http-consumer0-test-http-provider.json")
		if err != nil {
			t.Fatal(err.Error())
		}
		err = os.Remove("/tmp/test-http-consumer1-test-http-provider.json")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

}

func TestPluginInteraction(t *testing.T) {
	tmpPactFolder, err := ioutil.TempDir("", "pact-go")
	assert.NoError(t, err)
	log.SetLogLevel("trace")

	m := NewHTTPPact("test-plugin-consumer", "test-plugin-provider")

	// Protobuf plugin test
	m.UsingPlugin("protobuf", "0.0.3")
	m.WithSpecificationVersion(SPECIFICATION_VERSION_V4)

	i := m.NewInteraction("some plugin interaction")

	dir, _ := os.Getwd()
	path := fmt.Sprintf("%s/pact_plugin.proto", dir)

	protobufInteraction := `{
			"pact:proto": "` + path + `",
			"pact:message-type": "InitPluginRequest",
			"pact:content-type": "application/protobuf",
			"implementation": "notEmpty('pact-go-driver')",
			"version": "matching(semver, '0.0.0')"
		}`

	i.UponReceiving("some interaction").
		Given("plugin state").
		WithRequest("GET", "/protobuf").
		WithStatus(200).
		WithPluginInteractionContents(INTERACTION_PART_RESPONSE, "application/protobuf", protobufInteraction)

	port, err := m.Start("0.0.0.0:0", false)
	assert.NoError(t, err)
	defer m.CleanupMockServer(port)

	res, err := http.Get(fmt.Sprintf("http://0.0.0.0:%d/protobuf", port))
	assert.NoError(t, err)

	bytes, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	initPluginRequest := &InitPluginRequest{}
	proto.Unmarshal(bytes, initPluginRequest)
	assert.NoError(t, err)

	assert.Equal(t, "pact-go-driver", initPluginRequest.Implementation)
	assert.Equal(t, "0.0.0", initPluginRequest.Version)

	mismatches := m.MockServerMismatchedRequests(port)
	if len(mismatches) != 0 {
		assert.Len(t, mismatches, 0)
		t.Log(mismatches)
	}

	err = m.WritePactFile(port, tmpPactFolder)
	assert.NoError(t, err)
}

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
