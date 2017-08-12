package e2e

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bytes"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

var dir, _ = os.Getwd()
var brokerHost = os.Getenv("PACT_BROKER_HOST")
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/../log", dir)

// var name = "Jean-Marie de La Beaujardière😀😍"
var name = "billy"

var like = dsl.Like
var eachLike = dsl.EachLike
var term = dsl.Term

func TestPactIntegration_Consumer(t *testing.T) {
	pactDaemonPort := 6666

	// Create Pact connecting to local Daemon
	consumerPact := dsl.Pact{
		Port:     pactDaemonPort,
		Consumer: "billy",
		Provider: "bobby",
		LogLevel: "TRACE",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
	defer consumerPact.Teardown()

	// Pass in test case
	var test = func() error {
		// Get request /foobar
		_, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foobar", consumerPact.Server.Port))
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}

		// Post request /bazbat
		bodyRequest := bytes.NewBufferString(fmt.Sprintf(`{"name": "%s"}`, name))
		_, err = http.Post(fmt.Sprintf("http://127.0.0.1:%d/bazbat", consumerPact.Server.Port), "application/json", bodyRequest)
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}

		return err
	}

	// Setup a complex interaction
	colour := term("red", "red|green|blue")

	body :=
		eachLike(
			fmt.Sprintf(
				`{
            "name": "jumper",
						"size": 10,
						"colour": %s,
            "tag": ["jumper", "shirt"],
            "price": 1.07
					}`, colour),
			1)

	// Set up our interactions. Note we have multiple in this test case!
	consumerPact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   "/foobar",
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})
	consumerPact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test").
		WithRequest(dsl.Request{
			Method: "POST",
			Path:   "/bazbat",
			Body: fmt.Sprintf(`
          {
            "name": "%s"
          }`, name),
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body:   body,
		})

	// Verify Collaboration Test interactions (Consumer side)
	err := consumerPact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

	// Write pact to file `<pact-go>/pacts/my_consumer-my_provider.json`
	consumerPact.WritePact()

}

func TestPactIntegration_Publish(t *testing.T) {
	// Publish the Pacts...
	p := dsl.Publisher{}
	err := p.Publish(types.PublishRequest{
		PactURLs:        []string{"../pacts/billy-bobby.json"},
		PactBroker:      brokerHost,
		ConsumerVersion: "1.0.0",
		Tags:            []string{"latest", "sit4"},
		BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPactIntegration_Provider(t *testing.T) {
	// Setup Provider API for verification (later...)
	//providerPort, _ := utils.GetFreePort()
	providerPort := 80
	go setupProviderAPI(providerPort)
	pactDaemonPort := 6666

	// Wait for Provider to come up
	waitForPortInTest(providerPort, 5*time.Second, t)

	// Verify the Provider - local Pact Files
	providerPact := dsl.Pact{
		Port:     pactDaemonPort,
		Consumer: "billy",
		Provider: "bobby",
		LogLevel: "TRACE",
		LogDir:   logDir,
		PactDir:  pactDir,
	}
	err := providerPact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", providerPort),
		PactURLs:                   []string{filepath.ToSlash(fmt.Sprintf("%s/billy-bobby.json", pactDir))},
		ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", providerPort),
		PublishVerificationResults: false, // No HAL links in local pacts
		Verbose:                    true,
	})

	if err != nil {
		t.Fatal("Error:", err)
	}

	// Verify the Provider - Specific Published Pacts
	err = providerPact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", providerPort),
		PactURLs:                   []string{fmt.Sprintf("%s/pacts/provider/bobby/consumer/billy/latest/sit4", brokerHost)},
		ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", providerPort),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		Verbose:                    true,
	})

	if err != nil {
		t.Fatal("Error:", err)
	}

	// Verify the Provider - Latest Published Pacts for any known consumers
	err = providerPact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", providerPort),
		BrokerURL:                  brokerHost,
		ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", providerPort),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		Verbose:                    true,
	})

	if err != nil {
		t.Fatal("Error:", err)
	}

	// Verify the Provider - Tag-based Published Pacts for any known consumers
	err = providerPact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", providerPort),
		ProviderStatesSetupURL:     fmt.Sprintf("http://127.0.0.1:%d/setup", providerPort),
		BrokerURL:                  brokerHost,
		Tags:                       []string{"latest", "sit4"},
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

// Used as the Provider in the verification E2E steps
func setupProviderAPI(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] provider API: states setup")
		w.Header().Add("Content-Type", "application/json")
	})
	mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] provider API: /foobar")
		w.Header().Add("Content-Type", "application/json")
	})
	mux.HandleFunc("/bazbat", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] provider API: /bazbat")
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
      [
        {
          "size": 10,
          "name": "%s",
          "colour": "red",
          "price": 17.01,
          "tag": [
              "jumper",
              "shirt"
          ]
        }
			]`, name)
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", port, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))
}

// Use this to wait for a daemon to be running prior
// to running tests
func waitForPortInTest(port int, wait time.Duration, t *testing.T) {
	timeout := time.After(wait)
	for {
		select {
		case <-timeout:
			t.Fatalf("Expected server to start < 1s.")
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				return
			}
		}
	}
}
