package dsl

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

func TestPact_findConsumersNoTags(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	request := types.VerifyRequest{
		BrokerURL: s.URL,
	}
	err := findConsumers("bobby", &request)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	if len(request.PactURLs) != 2 {
		t.Fatalf("Expected 2 PactURLs but got: %d, %s", len(request.PactURLs), request.PactURLs)
	}
}

func TestPact_findConsumersWithTags(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:      []string{"dev", "prod"},
		BrokerURL: s.URL,
	}
	err := findConsumers("bobby", &request)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	if len(request.PactURLs) != 2 {
		t.Fatalf("Expected 2 PactURLs but got: %d", len(request.PactURLs))
	}
}

func TestPact_findConsumersBrokerDown(t *testing.T) {
	port, _ := utils.GetFreePort()
	request := types.VerifyRequest{
		Tags:      []string{"dev", "prod"},
		BrokerURL: fmt.Sprintf("http://localhost:%d", port),
	}
	err := findConsumers("idontexist", &request)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_findConsumersInvalidResponse(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:      []string{"broken"},
		BrokerURL: s.URL,
	}
	err := findConsumers("bobby", &request)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_findConsumersInvalidURL(t *testing.T) {
	request := types.VerifyRequest{
		BrokerURL: "%%%",
	}
	err := findConsumers("broken", &request)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_findConsumersErrorResponse(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:      []string{"dev"},
		BrokerURL: s.URL,
	}
	err := findConsumers("broken", &request)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_findConsumersNoConsumers(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:      []string{"dev", "prod"},
		BrokerURL: s.URL,
	}
	err := findConsumers("idontexist", &request)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_findConsumersAuthenticated(t *testing.T) {
	s := setupMockBroker(true)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:           []string{"dev", "prod"},
		BrokerURL:      s.URL,
		BrokerUsername: "foo",
		BrokerPassword: "bar",
	}
	err := findConsumers("bobby", &request)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
}

func TestPact_findConsumersAuthenticatedFail(t *testing.T) {
	s := setupMockBroker(true)
	defer s.Close()
	request := types.VerifyRequest{
		Tags:      []string{"dev", "prod"},
		BrokerURL: s.URL,
	}
	err := findConsumers("bobby", &request)

	switch err {
	case ErrUnauthorized:
	default:
		t.Fatalf("Expected error to be 'ErrNotAuthorized' but got %s", err)
	}
}

// Pretend to be a Broker for fetching Pacts
func setupMockBroker(auth bool) *httptest.Server {
	mux := http.NewServeMux()
	var authFunc func(inner http.HandlerFunc) http.HandlerFunc

	if auth {
		// Use the foo/bar basic authentication middleware in publish_test.go
		authFunc = func(inner http.HandlerFunc) http.HandlerFunc {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if checkAuth(w, r) {
					log.Println("[DEBUG] broker - authenticated!")
					inner.ServeHTTP(w, r)
					return
				}

				w.Header().Set("WWW-Authenticate", `Basic realm="Broker Authentication Required"`)
				w.WriteHeader(401)
				w.Write([]byte("401 Unauthorized\n"))
			})
		}
	} else {
		// Create a do-nothing authentication middleware
		authFunc = func(inner http.HandlerFunc) http.HandlerFunc {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println("[DEBUG] broker - no authentication")
				inner.ServeHTTP(w, r)
			})
		}
	}

	server := httptest.NewServer(mux)

	// Find latest 'bobby' consumers (no tag)
	// curl --user pactuser:pact -H "accept: application/hal+json" "http://pact.onegeek.com.au/pacts/provider/bobby/latest"
	mux.HandleFunc("/pacts/provider/bobby/latest", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] get pacts for provider 'bobby'")
		fmt.Fprintf(w, `{"_links":{"self":{"href":"%s/pacts/provider/bobby/latest","title":"Latest pact versions for the provider bobby"},"provider":{"href":"%s/pacticipants/bobby","title":"bobby"},"pacts":[{"href":"%s/pacts/provider/bobby/consumer/jessica/version/2.0.0","title":"Pact between jessica (v2.0.0) and bobby","name":"jessica"},{"href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.0","title":"Pact between billy (v1.0.0) and bobby","name":"billy"}]}}`, server.URL, server.URL, server.URL, server.URL)
		w.Header().Add("Content-Type", "application/hal+json")
	}))

	// Find 'bobby' consumers for tag 'prod'
	// curl --user pactuser:pact -H "accept: application/hal+json" "http://pact.onegeek.com.au/pacts/provider/bobby/latest/sit4"
	mux.Handle("/pacts/provider/bobby/latest/prod", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] get all pacts for provider 'bobby' where the tag 'prod' exists")
		fmt.Fprintf(w, `{"_links":{"self":{"href":"%s/pacts/provider/bobby/latest/dev","title":"Latest pact versions for the provider bobby with tag 'dev'"},"provider":{"href":"%s/pacticipants/bobby","title":"bobby"},"pacts":[{"href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.0","title":"Pact between billy (v1.0.0) and bobby","name":"billy"}]}}`, server.URL, server.URL, server.URL)
		w.Header().Add("Content-Type", "application/hal+json")
	}))

	// Broken response
	mux.Handle("/pacts/provider/bobby/latest/broken", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] broken broker")
		fmt.Fprintf(w, `broken response`)
		w.Header().Add("Content-Type", "application/hal+json")
	}))

	// 50x response
	// curl --user pactuser:pact -H "accept: application/hal+json" "http://pact.onegeek.com.au/pacts/provider/bobby/latest/sit4"
	mux.Handle("/pacts/provider/broken/latest/dev", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] broker broker response")
		w.WriteHeader(500)
		w.Write([]byte("500 Server Error\n"))
	}))

	// Find 'bobby' consumers for tag 'dev'
	// curl --user pactuser:pact -H "accept: application/hal+json" "http://pact.onegeek.com.au/pacts/provider/bobby/latest/sit4"
	mux.Handle("/pacts/provider/bobby/latest/dev", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] get all pacts for provider 'bobby' where the tag 'dev' exists")
		fmt.Fprintf(w, `{"_links":{"self":{"href":"%s/pacts/provider/bobby/latest/dev","title":"Latest pact versions for the provider bobby with tag 'dev'"},"provider":{"href":"%s/pacticipants/bobby","title":"bobby"},"pacts":[{"href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.1","title":"Pact between billy (v1.0.1) and bobby","name":"billy"}]}}`, server.URL, server.URL, server.URL)
		w.Header().Add("Content-Type", "application/hal+json")
	}))

	// Actual Consumer Pact
	// curl -v --user pactuser:pact -H "accept: application/json" http://pact.onegeek.com.au/pacts/provider/bobby/consumer/billy/version/1.0.0
	mux.Handle("/pacts/provider/bobby/consumer/billy/version/", authFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] get all pacts for provider 'bobby' where any tag exists")
		fmt.Fprintf(w, `{"consumer":{"name":"billy"},"provider":{"name":"bobby"},"interactions":[{"description":"Some name for the test","provider_state":"Some state","request":{"method":"GET","path":"/foobar"},"response":{"status":200,"headers":{"Content-Type":"application/json"}}},{"description":"Some name for the test","provider_state":"Some state2","request":{"method":"GET","path":"/bazbat"},"response":{"status":200,"headers":{},"body":[[{"colour":"red","size":10,"tag":[["jumper","shirt"],["jumper","shirt"]]}]],"matchingRules":{"$.body":{"min":1},"$.body[*].*":{"match":"type"},"$.body[*]":{"min":1},"$.body[*][*].*":{"match":"type"},"$.body[*][*].colour":{"match":"regex","regex":"red|green|blue"},"$.body[*][*].size":{"match":"type"},"$.body[*][*].tag":{"min":2},"$.body[*][*].tag[*].*":{"match":"type"},"$.body[*][*].tag[*][0]":{"match":"type"},"$.body[*][*].tag[*][1]":{"match":"type"}}}}],"metadata":{"pactSpecificationVersion":"2.0.0"},"updatedAt":"2016-06-11T13:11:33+00:00","createdAt":"2016-06-09T12:46:42+00:00","_links":{"self":{"title":"Pact","name":"Pact between billy (v1.0.0) and bobby","href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.0"},"pb:consumer":{"title":"Consumer","name":"billy","href":"%s/pacticipants/billy"},"pb:provider":{"title":"Provider","name":"bobby","href":"%s/pacticipants/bobby"},"pb:latest-pact-version":{"title":"Pact","name":"Latest version of this pact","href":"%s/pacts/provider/bobby/consumer/billy/latest"},"pb:previous-distinct":{"title":"Pact","name":"Previous distinct version of this pact","href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.0/previous-distinct"},"pb:diff-previous-distinct":{"title":"Diff","name":"Diff with previous distinct version of this pact","href":"%s/pacts/provider/bobby/consumer/billy/version/1.0.0/diff/previous-distinct"},"pb:pact-webhooks":{"title":"Webhooks for the pact between billy and bobby","href":"%s/webhooks/provider/bobby/consumer/billy"},"pb:tag-prod-version":{"title":"Tag this version as 'production'","href":"%s/pacticipants/billy/versions/1.0.0/tags/prod"},"pb:tag-version":{"title":"Tag version","href":"%s/pacticipants/billy/versions/1.0.0/tags/{tag}"},"curies":[{"name":"pb","href":"%s/doc/{rel}","templated":true}]}}`, server.URL, server.URL, server.URL, server.URL, server.URL, server.URL, server.URL, server.URL, server.URL, server.URL)
		w.Header().Add("Content-Type", "application/hal+json")
	}))

	return server
}
