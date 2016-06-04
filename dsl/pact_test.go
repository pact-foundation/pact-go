package dsl

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/daemon"
	"github.com/pact-foundation/pact-go/utils"
)

func createPact() {

}

func getPort(url string) int {
	i, _ := strconv.Atoi(strings.Split(url, ":")[2])
	return i
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
	createDaemon(port)

	pact := &Pact{Port: port}
	pact.Setup()
	if pact.Server == nil {
		t.Fatalf("Expected server to be created")
	}
}

func TestPact_Teardown(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)

	pact := &Pact{Port: port}
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

	// Create Pact connecting to local Daemon
	pact := &Pact{
		Port:     6666,
		Consumer: "My Consumer",
		Provider: "My Provider",
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
		})

	// Verify
	err := pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}
}
