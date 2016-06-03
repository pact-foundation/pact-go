package dsl

import (
	"errors"
	"testing"

	"github.com/mefellows/pact-go/utils"
)

func simplePact() (pact *PactMock) {
	pact = &PactMock{}
	pact.
		UponReceiving("Some name for the test").
		WithRequest(&PactRequest{}).
		WillRespondWith(&PactResponse{})
	return
}

func TestPact_Before(t *testing.T) {
	port, _ := utils.GetFreePort()
	createDaemon(port)

	// pact := &PactConsumer{Port: port}
	// pact.Before()
	// <-time.After(1 * time.Second)
	// pact.After()

	// Can I hit stuff?

	//

}

//

func providerStatesPact() (pact *PactMock) {
	pact = &PactMock{}
	pact.
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&PactRequest{}).
		WillRespondWith(&PactResponse{})
	return
}

func TestPactDSL(t *testing.T) {
	pact := &PactMock{}
	pact.
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&PactRequest{}).
		WillRespondWith(&PactResponse{})
}

func TestPactVerify_NoState(t *testing.T) {
	pact := simplePact()
	err := pact.Verify()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPactVerify_NoState_Fail(t *testing.T) {
	pact := simplePact()
	pact.VerifyResponse = errors.New("Pact failure!")
	err := pact.Verify()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPactVerify_State(t *testing.T) {
	pact := providerStatesPact()
	err := pact.Verify()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPactVerify_State_Fail(t *testing.T) {
	pact := providerStatesPact()
	pact.VerifyResponse = errors.New("Pact failure!")
	err := pact.Verify()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}
