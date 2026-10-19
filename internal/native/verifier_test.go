//go:build provider

package native

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	Init("INFO")
}

func TestVerifier_Version(_ *testing.T) {
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

func TestVerifier_ExecuteReportsResultToObserver(t *testing.T) {
	var got string
	ResultObserver = func(resultJSON string) { got = resultJSON }
	defer func() { ResultObserver = nil }()

	v := NewVerifier("pact-go", "test")
	defer v.Shutdown()
	require.NoError(t, v.Execute())

	assert.Contains(t, got, `"result":true`)
}

func TestVerifier_Shutdown(_ *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.Shutdown()
}

func TestVerifier_SetProviderInfo(_ *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.SetProviderInfo("name", "http", "localhost", 1234, "/")
}

func TestVerifier_SetConsumerFilters(_ *testing.T) {
	v := NewVerifier("pact-go", "test")
	v.SetConsumerFilters([]string{"consumer1", "consumer2"})
}
