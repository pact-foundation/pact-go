package main

import "fmt"
import "github.com/mefellows/pact-go/dsl"

func main() {
	fmt.Printf("Testing")
	pact := dsl.Pact{}
	pact.
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(dsl.Request{}).
		WillRespondWith(dsl.Response{})
}
