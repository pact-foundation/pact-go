package dsl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/types"
)

func init() {
	// mock out this function
	checkCliCompatibility = func() {}
}

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
		t.Fatalf("want error but got none")
	}

	if !strings.Contains(err.Error(), "something went wrong") {
		t.Fatalf("expected response body to contain an error message 'something went wrong' but got '%s'", err.Error())
	}
}

func TestPact_Setup(t *testing.T) {
	pact := &Pact{LogLevel: "DEBUG"}
	defer stubPorts()()
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("expected server to be created")
	}

	pact2 := &Pact{LogLevel: "DEBUG"}
	pact2.Setup(false)
	if pact2.Server != nil {
		t.Fatalf("expected server to be nil")
	}
	if pact2.pactClient == nil {
		t.Fatalf("expected non-nil client")
	}
}

func TestPact_SetupWithMockServerPort(t *testing.T) {
	c, _ := createMockClient(true)
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768", pactClient: c}
	defer stubPorts()()
	pact.Setup(true)

	if pact.Server == nil {
		t.Fatalf("expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatalf("expected mock daemon to be started on specific port")
	}
}

func TestPact_SetupWithMockServerPortCSV(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768,32769", pactClient: c}
	pact.Setup(true)

	if pact.Server == nil {
		t.Fatalf("expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatalf("expected mock daemon to be started on specific port")
	}
}

func TestPact_SetupWithMockServerPortRange(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "32768-32770", pactClient: c}
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("expected server to be created")
	}
	if pact.Server.Port != 32768 {
		t.Fatal("expected mock daemon to be started on specific port, got", 32768)
	}
}

func TestPact_Invalidrange(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", AllowedMockServerPorts: "abc-32770", pactClient: c}
	pact.Setup(true)
	if pact.Server == nil {
		t.Fatalf("expected server to be created")
	}
	if pact.Server.Port != 0 {
		t.Fatalf("expected mock daemon to be started on specific port")
	}
}

func TestPact_Teardown(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	pact.Setup(true)
	pact.Teardown()
	if pact.Server.Error != nil {
		t.Fatal("got error:", pact.Server.Error)
	}
}

func TestPact_TeardownFail(t *testing.T) {
	c := &mockClient{}

	pact := &Pact{LogLevel: "DEBUG", pactClient: c, Server: &types.MockServer{}}
	pact.Teardown()
}

func TestPact_VerifyProviderRaw(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()

	dummyMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
		RequestFilter:   dummyMiddleware,
		BeforeHook: func() error {
			fmt.Println("aeuaoseu")
			return nil
		},
		AfterHook: func() error {
			fmt.Println("aeuaoseu")
			return nil
		},
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProvider(t *testing.T) {
	c, _ := createMockClient(true)
	defer stubPorts()()
	exampleTest := &testing.T{}
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}

	_, err := pact.VerifyProvider(exampleTest, types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProviderFail(t *testing.T) {
	c, _ := createMockClient(false)
	defer stubPorts()()
	exampleTest := &testing.T{}
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}

	_, err := pact.VerifyProvider(exampleTest, types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestPact_VerifyProviderFailBadURL(t *testing.T) {
	c, _ := createMockClient(false)
	defer stubPorts()()
	exampleTest := &testing.T{}
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}

	_, err := pact.VerifyProvider(exampleTest, types.VerifyRequest{
		ProviderBaseURL: "http.",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestPact_VerifyProviderBroker(t *testing.T) {
	s := setupMockBroker(false)
	defer s.Close()
	c, _ := createMockClient(true)
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

func TestPact_VerifyProviderRawFail(t *testing.T) {
	c, _ := createMockClient(false)
	defer stubPorts()()
	pact := &Pact{LogLevel: "DEBUG", pactClient: c}
	_, err := pact.VerifyProviderRaw(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err == nil {
		t.Fatalf("expected error but got none")
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
		t.Fatalf("expected 2 interactions to be added to Pact but got %d", len(pact.Interactions))
	}
}

func TestPact_BeforeHook(t *testing.T) {
	var called bool

	req, err := http.NewRequest("GET", "/__setup", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := beforeHookMiddleware(func() error {
		called = true
		return nil
	})
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect hook to be called
	if !called {
		t.Error("expected state handler to have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Error("expected http handler to be invoked")
	}
}
func TestPact_BeforeHookNotSetupPath(t *testing.T) {
	var called bool

	req, err := http.NewRequest("GET", "/blah", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := beforeHookMiddleware(func() error {
		called = true
		return nil
	})
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect hook not to be called
	if called {
		t.Error("expected state handler to not have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Error("expected http handler to be invoked")
	}
}
func TestPact_AfterHook(t *testing.T) {
	var called bool

	req, err := http.NewRequest("GET", "/blah", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := afterHookMiddleware(func() error {
		called = true
		return nil
	})
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect hook to be called
	if !called {
		t.Error("expected state handler to have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Error("expected http handler to be invoked")
	}
}
func TestPact_AfterHookSetupPath(t *testing.T) {
	var called bool

	req, err := http.NewRequest("GET", "/__setup", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := afterHookMiddleware(func() error {
		called = true
		return nil
	})

	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect state handler
	if called {
		t.Error("expected state handler to not have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Error("expected http handler to be invoked")
	}
}

func TestPact_StateHandlerMiddlewareStateHandlerExists(t *testing.T) {
	var called bool

	handlers := map[string]types.StateHandler{
		"state x": func() error {
			called = true

			return nil
		},
	}

	req, err := http.NewRequest("POST", "/__setup", strings.NewReader(`{
		"states": ["state x"],
		"consumer": "test",
		"provider": "provider"
		}`))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := stateHandlerMiddleware(handlers)
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect state handler
	if !called {
		t.Error("expected state handler to have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h == "true" {
		t.Error("expected http handler to not be invoked")
	}
}

func TestPact_StateHandlerMiddlewareStateHandlerNotExists(t *testing.T) {
	var called bool

	handlers := map[string]types.StateHandler{}

	req, err := http.NewRequest("POST", "/__setup", strings.NewReader(`{
		"states": ["state x"],
		"consumer": "test",
		"provider": "provider"
		}`))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := stateHandlerMiddleware(handlers)
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// Expect state handler
	if called {
		t.Error("expected state handler to not have been called")
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h == "true" {
		t.Error("expected http handler to not be invoked")
	}
}

func TestPact_StateHandlerMiddlewareStateHandlerError(t *testing.T) {
	handlers := map[string]types.StateHandler{
		"state x": func() error {
			return errors.New("handler error")
		},
	}

	req, err := http.NewRequest("POST", "/__setup", strings.NewReader(`{
		"states": ["state x"],
		"consumer": "test",
		"provider": "provider"
		}`))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := stateHandlerMiddleware(handlers)
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// expect 500
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("want statusCode to be 500, goto %v", status)
	}

	// expect http handler to NOT have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h == "true" {
		t.Error("expected http handler to not be invoked")
	}
}

func TestPact_StateHandlerMiddlewarePassThroughInvalidPath(t *testing.T) {
	handlers := map[string]types.StateHandler{}

	req, err := http.NewRequest("POST", "/someotherpath", strings.NewReader(`{ }`))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mw := stateHandlerMiddleware(handlers)
	mw(dummyHandler("X-Dummy-Handler")).ServeHTTP(rr, req)

	// expect http handler to have been called
	if h := rr.HeaderMap.Get("X-Dummy-Handler"); h != "true" {
		t.Errorf("expected target http handler to be invoked")
	}
}

func dummyHandler(header string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header, "true")
	})
}

func TestPact_AddMessage(t *testing.T) {
	// AddMessage is a fairly useless method, currently,
	// but is reserved for future API usage.
	pact := &Pact{}
	pact.AddMessage()

	if len(pact.MessageInteractions) != 1 {
		t.Errorf("expected pact to have 1 Message Interaction but got %v", len(pact.MessageInteractions))
	}

}

var message = `{
	"providerStates": [
		{
			"name": "state x"
		}
	],
	"metadata": {
		"content-type": "application-json"
	},
	"content": {"foo": "bar"},
	"description": "a message"
	}`

func createMessageHandlers(invoked *int, err error) MessageHandlers {
	return map[string]MessageHandler{
		"a message": func(m Message) (interface{}, error) {
			*invoked++
			fmt.Println("message handler")

			return nil, err
		},
	}
}

func createStateHandlers(invoked *int, err error) StateHandlers {
	return map[string]StateHandler{
		"state x": func(s State) error {
			fmt.Println("state handler")
			*invoked++

			return err
		},
	}
}

func TestPact_MessageVerificationHandler(t *testing.T) {
	var called = 0

	req, err := http.NewRequest("POST", "/", strings.NewReader(message))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := messageVerificationHandler(createMessageHandlers(&called, nil), createStateHandlers(&called, nil))
	h.ServeHTTP(rr, req)

	// Expect state handler
	if called != 2 {
		t.Error("expected state handler and message handler to have been called", called)
	}
}

func TestPact_MessageVerificationHandlerInvalidMessage(t *testing.T) {
	var called = 0

	req, err := http.NewRequest("POST", "/", strings.NewReader("{broken"))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := messageVerificationHandler(createMessageHandlers(&called, nil), createStateHandlers(&called, nil))
	h.ServeHTTP(rr, req)

	// Expect state handler
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 500 but got %v", rr.Code)
	}
}

func TestPact_MessageVerificationHandlerMessageNotFound(t *testing.T) {
	var called = 0
	var badMessage = `{
		"content": {"foo": "bar"},
		"description": "a message not found"
	}`

	req, err := http.NewRequest("POST", "/", strings.NewReader(badMessage))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := messageVerificationHandler(createMessageHandlers(&called, nil), createStateHandlers(&called, nil))
	h.ServeHTTP(rr, req)

	// Expect state handler
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 but got %v", rr.Code)
	}
}

func TestPact_MessageVerificationHandlerStateHandlerFail(t *testing.T) {
	var called = 0

	req, err := http.NewRequest("POST", "/", strings.NewReader(message))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := messageVerificationHandler(createMessageHandlers(&called, nil), createStateHandlers(&called, errors.New("state handler failed")))
	h.ServeHTTP(rr, req)

	// Expect state handler
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 but got %v", rr.Code)
	}
}

func TestPact_MessageVerificationHandlerMessageHandlerFail(t *testing.T) {
	var called = 0

	req, err := http.NewRequest("POST", "/", strings.NewReader(message))

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := messageVerificationHandler(createMessageHandlers(&called, errors.New("message handler failed")), createStateHandlers(&called, nil))
	h.ServeHTTP(rr, req)

	// Expect state handler
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 but got %v", rr.Code)
	}
}

func TestPact_VerifyMessageConsumerRawFailToReify(t *testing.T) {
	pact := &Pact{}

	message := pact.AddMessage()
	message.
		Given("user with id 1 exists").
		ExpectsToReceive("a user").
		WithContent(map[string]interface{}{
			"id": Like(1),
		})

	client, _ := createMockClient(false)

	pact.pactClient = client

	var invoked bool
	h := func(m Message) error {
		invoked = true
		return nil
	}
	err := pact.VerifyMessageConsumerRaw(message, h)

	if err == nil {
		t.Fatalf("expected error but got none")
	}

	if invoked {
		t.Fatalf("expected handler not to be invoked")
	}
}

func TestPact_VerifyMessageConsumerRawSuccess(t *testing.T) {
	pact := &Pact{}

	message := pact.AddMessage()
	message.
		Given("user with id 1 exists").
		ExpectsToReceive("a user").
		WithContent(map[string]interface{}{
			"foo": "bar",
		})
		// TODO: replace the createMockClient below and mock out properly
		// AsType(idType{})

	c, _ := createMockClient(true)

	pact.pactClient = c

	var invoked bool
	h := func(m Message) error {
		invoked = true
		return nil
	}

	err := pact.VerifyMessageConsumerRaw(message, h)

	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if !invoked {
		t.Fatalf("expected handler to be invoked")
	}
}

func TestPact_VerifyMessageConsumerFail(t *testing.T) {
	pact := &Pact{}

	message := pact.AddMessage()
	message.
		Given("user with id 1 exists").
		ExpectsToReceive("a user").
		WithContent(map[string]interface{}{
			"foo": "bar",
		})

	c, _ := createMockClient(true)

	pact.pactClient = c

	h := func(m Message) error {
		return errors.New("message handler failed")
	}
	exampleTest := &testing.T{}
	err := pact.VerifyMessageConsumer(exampleTest, message, h)

	if err == nil {
		t.Fatalf("expected error but got none")
	}
}

func TestPact_VerifyMessageConsumerSuccess(t *testing.T) {
	pact := &Pact{}

	message := pact.AddMessage()
	message.
		Given("user with id 1 exists").
		ExpectsToReceive("a user").
		WithContent(map[string]interface{}{
			"foo": "bar",
		})

	c, _ := createMockClient(true)

	pact.pactClient = c

	var invoked bool
	h := func(m Message) error {
		invoked = true
		return nil
	}
	exampleTest := &testing.T{}
	err := pact.VerifyMessageConsumer(exampleTest, message, h)

	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if !invoked {
		t.Fatalf("expected handler to be invoked")
	}
}

func TestPact_VerifyMessageProviderSuccess(t *testing.T) {
	c, _ := createMockClient(true)
	var called = 0
	defer stubPorts()()
	exampleTest := &testing.T{}

	pact := &Pact{LogLevel: "DEBUG", pactClient: c}

	_, err := pact.VerifyMessageProvider(exampleTest, VerifyMessageRequest{
		PactURLs:        []string{"foo.json", "bar.json"},
		MessageHandlers: createMessageHandlers(&called, nil),
		StateHandlers:   createStateHandlers(&called, nil),
	})

	if err != nil {
		t.Fatal("Error:", err)
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
	waitForPort = func(int, string, string, time.Duration, string) error {
		return nil
	}
	return func() { waitForPort = old }
}
