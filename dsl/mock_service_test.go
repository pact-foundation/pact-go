package dsl

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Simple mock server for testing. This is getting confusing...
func setupMockServer(success bool, t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		t.Logf("%v\n", r)
		t.Logf("Request Body: %s\n", request)

		if success {
			fmt.Fprintln(w, "Hello, client")
		} else {
			http.Error(w, "something went wrong\n", http.StatusInternalServerError)
		}
	}))

	return ts
}

func TestMockService_AddInteraction(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})
	err := mockService.AddInteraction(i)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMockService_AddInteractionFail(t *testing.T) {
	ms := setupMockServer(false, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&Request{}).
		WillRespondWith(&Response{})
	err := mockService.AddInteraction(i)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestMockService_DeleteInteractions(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}
	err := mockService.DeleteInteractions()

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMockService_WritePact(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL:  ms.URL,
		Consumer: "Foo Consumer",
		Provider: "Bar Provider",
	}

	err := mockService.WritePact()

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMockService_WritePactFail(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}

	err := mockService.WritePact()

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestMockService_Verify(t *testing.T) {
	ms := setupMockServer(true, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}

	err := mockService.Verify()

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestMockService_VerifyFail(t *testing.T) {
	ms := setupMockServer(false, t)
	defer ms.Close()

	mockService := &MockService{
		BaseURL: ms.URL,
	}

	err := mockService.Verify()

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}
