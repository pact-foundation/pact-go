package dsl

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/types"
)

func createMockRemoteServer(valid bool) (*httptest.Server, string) {
	file := createSimplePact(valid)
	dir := filepath.Dir(file.Name())
	path := filepath.Base(file.Name())
	server := httptest.NewServer(http.FileServer(http.Dir(dir)))

	return server, fmt.Sprintf("%s/%s", server.URL, path)
}

func createSimplePact(valid bool) *os.File {
	var data []byte
	if valid {
		data = []byte(`
    {
      "consumer": {
        "name": "Some Consumer"
      },
      "provider": {
        "name": "Some Provider"
      }
    }
  `)
	} else {
		data = []byte(`
    {
      "consumer": {
        "name": "Some Consumer"
      }
    }
  `)
	}

	tmpfile, err := ioutil.TempFile("", "pactgo")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpfile.Write(data); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpfile
}

var checkAuth = func(w http.ResponseWriter, r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	return pair[0] == "foo" && pair[1] == "bar"
}

func createMockRemoteServerWithAuth(valid bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(w, r) {
			w.Write([]byte("Authenticated!"))
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="MY REALM"`)
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	}))

	return ts
}

func TestPublish_Publish(t *testing.T) {
	c := newMockClient()
	p := Publisher{
		pactClient: c,
	}
	r := types.PublishRequest{
		PactURLs:        []string{"/tmp/file.json"},
		PactBroker:      "http://foo.com",
		ConsumerVersion: "1.0.0",
	}
	err := p.Publish(r)

	if err != nil {
		t.Fatal(err)
	}
}
func TestPublish_PublishFail(t *testing.T) {
	c := newMockClient()
	c.PublishPactsError = fmt.Errorf("unable to publish to broker")
	p := Publisher{
		pactClient: c,
	}
	r := types.PublishRequest{
		PactURLs:        []string{"/tmp/file.json"},
		PactBroker:      "http://foo.com",
		ConsumerVersion: "1.0.0",
	}
	err := p.Publish(r)

	if err == nil {
		t.Fatal("want error, got none")
	}
}

func TestPublish_PublishFailIntegration(t *testing.T) {
	p := Publisher{}

	r := types.PublishRequest{
		PactURLs:        []string{"/tmp/file.json"},
		PactBroker:      "a", // this will actually try to publish, should fail
		ConsumerVersion: "1.0.0",
	}
	err := p.Publish(r)

	if err == nil {
		t.Fatal("want error, got none")
	}
}
