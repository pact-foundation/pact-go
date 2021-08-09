package dsl

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/types"
)

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
