//go:build consumer
// +build consumer

// Package main contains a runnable Consumer Pact test example.
package main

import (
	"fmt"
	"testing"
	// message "github.com/pact-foundation/pact-go/v2/message/v4"
	// "github.com/stretchr/testify/assert"
)

func TestSynchronousMessageConsumer(t *testing.T) {
	fmt.Println("test")
	// p, _ := message.NewSynchronousPact(message.Config{
	// 	Consumer: "consumer",
	// 	Provider: "provider",
	// 	PactDir:  "/tmp/",
	// })
}
