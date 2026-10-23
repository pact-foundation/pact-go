//go:build provider

package native

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestStringArrayToCStringArray(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		array, release := stringArrayToCStringArray(nil)
		assert.Nil(t, array)
		assert.NotNil(t, release)
		release()
	})

	t.Run("values", func(t *testing.T) {
		array, release := stringArrayToCStringArray([]string{"consumer1", "consumer2"})
		assert.NotNil(t, array)
		assert.NotNil(t, release)
		release()
	})
}
