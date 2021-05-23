package verifier

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	Init()
}

func TestVerifier_Version(t *testing.T) {
	fmt.Println("version: ", Version())
}

func TestVerifier_Verify(t *testing.T) {
	t.Run("invalid args returns an error", func(t *testing.T) {

		v := Verifier{}
		args := []string{
			"--file",
			"/non/existent/path.json",
			"--hostname",
			"localhost",
			"--port",
			"55827",
			"--state-change-url",
			"http://localhost:55827/__setup/",
			"--loglevel",
			"trace",
		}

		res := v.Verify(args)

		assert.Error(t, res)
	})
}
