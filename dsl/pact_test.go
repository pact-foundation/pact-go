package dsl

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/daemon"
	"github.com/pact-foundation/pact-go/utils"
)

func TestPact_setupLogging(t *testing.T) {
	res := captureOutput(func() {
		(&Pact{LogLevel: "DEBUG"}).setupLogging()
		log.Println("[DEBUG] this should display")
	})

	if !strings.Contains(res, "[DEBUG] this should display") {
		t.Fatalf("Expected log message to contain '[DEBUG] this should display' but got '%s'", res)
	}

	res = captureOutput(func() {
		(&Pact{LogLevel: "INFO"}).setupLogging()
		log.Print("[DEBUG] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}

	res = captureOutput(func() {
		(&Pact{LogLevel: "NONE"}).setupLogging()
		log.Print("[ERROR] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}
}

// Capture output from a log write
func captureOutput(action func()) string {
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	action()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stderr = rescueStderr

	return strings.TrimSpace(string(out))
}

func TestPact_Verify(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()
	testCalled := false
	var testFunc = func() error {
		testCalled = true
		return nil
	}

	pact := &Pact{
		Server: &daemon.PactMockServer{
			Port: getPort(ms.URL),
		},
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})

	err := pact.Verify(testFunc)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if testCalled == false {
		t.Fatalf("Expected test function to be called but it was not")
	}
}

func TestPact_VerifyFail(t *testing.T) {
	ms := setupMockServer(false, t)
	defer ms.Close()
	var testFunc = func() error { return nil }

	pact := &Pact{
		Server: &daemon.PactMockServer{
			Port: getPort(ms.URL),
		},
	}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})

	err := pact.Verify(testFunc)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	if !strings.Contains(err.Error(), "something went wrong") {
		t.Fatalf("Expected response body to contain an error message 'something went wrong' but got '%s'", err.Error())
	}
}

func TestPact_Setup(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port, true)

	pact := &Pact{Port: port, LogLevel: "DEBUG"}
	pact.Setup()
	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
}

func TestPact_Teardown(t *testing.T) {
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, true)

	pact := &Pact{Port: port, LogLevel: "DEBUG"}
	pact.Setup()
	pact.Teardown()
	if pact.Server.Status != 0 {
		t.Fatalf("Expected server exit status to be 0 but got %d", pact.Server.Status)
	}
}

func TestPact_AddInteraction(t *testing.T) {
	pact := &Pact{}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})

	pact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test2").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})

	if len(pact.Interactions) != 2 {
		t.Fatalf("Expected 2 interactions to be added to Pact but got %d", len(pact.Interactions))
	}
}

func TestPact_Integration(t *testing.T) {
	t.Skip() // Enable when running E2E/integration tests before a release

	// Setup Provider API for verification (later...)
	providerPort := setupProviderAPI()
	pactDaemonPort := 6666

	// Create Pact connecting to local Daemon
	pact := &Pact{
		Port:     pactDaemonPort,
		Consumer: "My Consumer",
		Provider: "My Provider",
		LogLevel: "DEBUG",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port))
		if err != nil {
			t.Fatalf("Error sending request: %v", err)
		}
		_, err = http.Get(fmt.Sprintf("http://localhost:%d/bazbat", pact.Server.Port))
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
						"size": %s,
						"colour": %s,
						"tag": %s
					}`, size, colour, tag),
					1),
				1))

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{
			Method: "GET",
			Path:   "/foobar",
		}).
		WillRespondWith(&Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		})
	pact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test").
		WithRequest(&Request{
			Method: "GET",
			Path:   "/bazbat",
		}).
		WillRespondWith(&Response{
			Status: 200,
			Body:   body,
		})

	// Verify
	err := pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

	response := pact.VerifyProvider(&daemon.VerifyRequest{
		ProviderBaseURL:        fmt.Sprintf("http://localhost:%d", providerPort),
		PactURLs:               []string{"./pacts/my_consumer-my_provider.json"},
		ProviderStatesURL:      fmt.Sprintf("http://localhost:%d/states", providerPort),
		ProviderStatesSetupURL: fmt.Sprintf("http://localhost:%d/setup", providerPort),
	})

	fmt.Println(response.Message)
}

// Used as the Provider in the verification E2E steps
func setupProviderAPI() int {
	port, _ := utils.GetFreePort()
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] provider API: states setup")
		w.Header().Add("Content-Type", "application/json")
	})
	mux.HandleFunc("/states", func(w http.ResponseWriter, req *http.Request) {
		log.Println("[DEBUG] provider API: states")
		fmt.Fprintf(w, `{"My Consumer": ["Some state", "Some state2"]}`)
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
