//go:build provider
// +build provider

package native

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	Init("INFO")
}

func TestVerifier_Version(t *testing.T) {
	fmt.Println("version: ", Version())
}

func TestVerifier_NewForApplication(t *testing.T) {
	v := NewVerifier("pact-go", "test")

	assert.NotNil(t, v.handle)
}

func TestVerifier_Execute(t *testing.T) {
	v := NewVerifier("pact-go", "test")
	err := v.Execute()

	assert.NoError(t, err)
}

func TestVerifier_Shutdown(t *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.Shutdown()
}

func TestVerifier_SetProviderInfo(t *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.SetProviderInfo("name", "http", "localhost", 1234, "/")
}

func TestVerifier_SetConsumerFilters(t *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.SetConsumerFilters([]string{"consumer1", "consumer2"})
}
