package verifier

import (
	"fmt"
	"testing"
)

func TestVerifier_Version(t *testing.T) {
	v := Verifier{}
	fmt.Println("version: ", v.Version())
}

func TestVerifier_Verify(t *testing.T) {
	v := Verifier{}
	v.Init()
	args := []string{
		"--afile",
		"/Users/matthewfellows/go/src/github.com/pact-foundation/pact-go/examples/v3/pacts/V3Consumer-V3Provider.json",
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
	fmt.Println("result: ", res)
}

func TestFoo(t *testing.T) {
	s := `error: Found argument \'--afile\' which wasn\'t expected, or isn\'t valid in this context\n\tDid you mean \u{1b}[32m--\u{1b}[0m\u{1b}[32mfile\u{1b}[0m?\n\nUSAGE:\n    pact_verifier_cli [FLAGS] [OPTIONS] --broker-url <broker-url>... --dir <dir>... --file <file>... --provider-name <provider-name> --url <url>...\n\nFor more information try --help`
	fmt.Println(s)
}
