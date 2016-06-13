package main

import "github.com/pact-foundation/pact-go/examples/go-kit/consumer"

func main() {
	client := consumer.Client{
		Host: "http://localhost:8080",
	}
	client.Run()
}
