package dsl

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../pacts", dir)
var logDir = fmt.Sprintf("%s/../log", dir)

func TestPact_Integration(t *testing.T) {
	// Enable when running E2E/integration tests before a release
	if os.Getenv("PACT_INTEGRATED_TESTS") != "" {

		// Setup Provider API for verification (later...)
		providerPort := setupProviderAPI()
		pactDaemonPort := 6666

		// Create Pact connecting to local Daemon
		consumerPact := Pact{
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
			_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", consumerPact.Server.Port))
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}
			_, err = http.Get(fmt.Sprintf("http://localhost:%d/bazbat", consumerPact.Server.Port))
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
			}

			return err
		}

		// Setup a complex interaction
		jumper := Like(`"jumper"`)
		shirt := Like(`"shirt"`)
		tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
		size := Like(10)
		colour := Term("red", "red|green|blue")

		body :=
			formatJSON(
				EachLike(
					EachLike(
						fmt.Sprintf(
							`{
            "name": "Jean-Marie de La Beaujardière😀😍",
						"size": %s,
						"colour": %s,
						"tag": %s
					}`, size, colour, tag),
						1),
					1))

		// Set up our interactions. Note we have multiple in this test case!
		consumerPact.
			AddInteraction().
			Given("Some state").
			UponReceiving("Some name for the test").
			WithRequest(Request{
				Method: "GET",
				Path:   "/foobar",
			}).
			WillRespondWith(Response{
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			})
		consumerPact.
			AddInteraction().
			Given("Some state2").
			UponReceiving("Some name for the test").
			WithRequest(Request{
				Method: "GET",
				Path:   "/bazbat",
			}).
			WillRespondWith(Response{
				Status: 200,
				Body:   body,
			})

		// Verify Collaboration Test interactionns (Consumer sid)
		err := consumerPact.Verify(test)
		if err != nil {
			t.Fatalf("Error on Verify: %v", err)
		}

		// Write pact to file `<pact-go>/pacts/my_consumer-my_provider.json`
		consumerPact.WritePact()

		// Publish the Pacts...
		p := Publisher{}
		brokerHost := os.Getenv("PACT_BROKER_HOST")
		err = p.Publish(types.PublishRequest{
			PactURLs:        []string{"../pacts/billy-bobby.json"},
			PactBroker:      brokerHost,
			ConsumerVersion: "1.0.0",
			Tags:            []string{"latest", "sit4"},
			BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
		})

		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		// Verify the Provider - local Pact Files
		providerPact := Pact{
			Port:     pactDaemonPort,
			Consumer: "billy",
			Provider: "bobby",
			LogLevel: "TRACE",
			LogDir:   logDir,
			PactDir:  pactDir,
		}
		err = providerPact.VerifyProvider(types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", providerPort),
			PactURLs:                   []string{fmt.Sprintf("%s/billy-bobby.json", pactDir)},
			ProviderStatesSetupURL:     fmt.Sprintf("http://localhost:%d/setup", providerPort),
			PublishVerificationResults: false, // No HAL links in local pacts
		})

		if err != nil {
			t.Fatal("Error:", err)
		}

		// Verify the Provider - Specific Published Pacts
		err = providerPact.VerifyProvider(types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", providerPort),
			PactURLs:                   []string{fmt.Sprintf("%s/pacts/provider/bobby/consumer/billy/latest/sit4", brokerHost)},
			ProviderStatesSetupURL:     fmt.Sprintf("http://localhost:%d/setup", providerPort),
			BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
			PublishVerificationResults: true,
			ProviderVersion:            "1.0.0",
		})

		if err != nil {
			t.Fatal("Error:", err)
		}

		// Verify the Provider - Latest Published Pacts for any known consumers
		err = providerPact.VerifyProvider(types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", providerPort),
			BrokerURL:                  brokerHost,
			ProviderStatesSetupURL:     fmt.Sprintf("http://localhost:%d/setup", providerPort),
			BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
			BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
			PublishVerificationResults: true,
			ProviderVersion:            "1.0.0",
		})

		if err != nil {
			t.Fatal("Error:", err)
		}

		// Verify the Provider - Tag-based Published Pacts for any known consumers
		err = providerPact.VerifyProvider(types.VerifyRequest{
			ProviderBaseURL:            fmt.Sprintf("http://localhost:%d", providerPort),
			ProviderStatesSetupURL:     fmt.Sprintf("http://localhost:%d/setup", providerPort),
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
}

// Used as the Provider in the verification E2E steps
func setupProviderAPI() int {
	port, _ := utils.GetFreePort()
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
			  [
			    {
            "name": "Jean-Marie de La Baujardière😀😍",
            "size": 10,
			      "colour": "red",
			      "tag": [
			        [
			          "jumper",
			          "shirt"
			        ],
			        [
			          "jumper",
			          "shirt"
			        ]
			      ]
			    }
			  ]
			]`)
	})

	go http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	return port
}
