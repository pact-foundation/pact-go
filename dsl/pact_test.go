package dsl

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/types"
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

func TestPact_Verify(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()
	testCalled := false
	var testFunc = func() error {
		testCalled = true
		return nil
	}

	pact := &Pact{
		Server: &types.MockServer{
			Port: getPort(ms.URL),
		},
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	err := pact.Verify(testFunc)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if testCalled == false {
		t.Fatalf("Expected test function to be called but it was not")
	}
}
func TestPact_VerifyMockServerFail(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()
	var testFunc = func() error { return nil }

	pact := &Pact{Server: &types.MockServer{Port: 1}}
	err := pact.Verify(testFunc)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_WritePact(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	pact := &Pact{
		Server: &types.MockServer{
			Port: getPort(ms.URL),
		},
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	err := pact.WritePact()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPact_WritePactFail(t *testing.T) {
	ms := setupMockServer(false, t)
	defer ms.Close()

	pact := &Pact{
		Server: &types.MockServer{
			Port: getPort(ms.URL),
		},
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	err := pact.WritePact()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_VerifyFail(t *testing.T) {
	ms := setupMockServer(false, t)
	defer ms.Close()
	var testFunc = func() error { return nil }

	pact := &Pact{
		Server: &types.MockServer{
			Port: getPort(ms.URL),
		},
	}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	err := pact.Verify(testFunc)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	if !strings.Contains(err.Error(), "something went wrong") {
		t.Fatalf("Expected response body to contain an error message 'something went wrong' but got '%s'", err.Error())
	}
}

func TestPact_Setup(t *testing.T) {
	pact := &Pact{LogLevel: "DEBUG"}
	defer stubPorts()()
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
	pact2 := &Pact{LogLevel: "DEBUG"}
	pact2.Setup(false)
	if pact2.Server != nil {
		t.Fatalf("Expected server to be nil")
	}
	if pact2.pactClient == nil {
		t.Fatalf("Needed to still have a client")
	}
}

func TestPact_SetupWithMockServerPort(t *testing.T) {
	c, _ := createClient(true)
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768", pactClient: c}
	defer stubPorts()()
	pact.Setup(true)

	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatalf("Expected mock daemon to be started on specific port")
	}
}

func TestPact_SetupWithMockServerPortCSV(t *testing.T) {
	c, _ := createClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768,32769", pactClient: c}
	pact.Setup(true)

	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatalf("Expected mock daemon to be started on specific port")
	}
}

func TestPact_SetupWithMockServerPortRange(t *testing.T) {
	c, _ := createClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768-32770", pactClient: c}
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatal("Expected mock daemon to be started on specific port, got", 32768)
	}
}

func TestPact_Invalidrange(t *testing.T) {
	c, _ := createClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "abc-32770", pactClient: c}
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
	if pact.Server.Port != 0 {
		t.Fatalf("Expected mock daemon to be started on specific port")
	}
}

func TestPact_Teardown(t *testing.T) {
	c, _ := createClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	pact.Setup(true)
	pact.Teardown()
	if pact.Server.Error != nil {
		t.Fatal("got error:", pact.Server.Error)
	}
}

func TestPact_VerifyProvider(t *testing.T) {
	c, _ := createClient(true)
	defer stubPorts()()

	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProviderBroker(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	c, _ := createClient(true)
	defer stubPorts()()

	pact := &Pact{LogLevel: "DEBUG", pactClient: c, Provider: "bobby"}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL:            "http://www.foo.com",
		BrokerURL:                  s.URL,
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProviderBrokerNoConsumers(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	c, _ := createClient(true)

	pact := &Pact{LogLevel: "DEBUG", pactClient: c, Provider: "providernotexist"}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		BrokerURL:       s.URL,
	})

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_VerifyProviderFail(t *testing.T) {
	c, _ := createClient(false)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_AddInteraction(t *testing.T) {
	pact := &Pact{}
	defer stubPorts()()

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	pact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test2").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	if len(pact.Interactions) != 2 {
		t.Fatalf("Expected 2 interactions to be added to Pact but got %d", len(pact.Interactions))
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

func stubPorts() func() {
	log.Println("Stubbing port timeout")
	old := waitForPort
	waitForPort = func(int, string, string, string) error {
		return nil
	}
	return func() { waitForPort = old }
}
