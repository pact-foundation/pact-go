# Pact Go

Golang version of [Pact](http://pact.io). Enables consumer driven contract testing, providing a mock service and
DSL for the consumer project, and interaction playback and verification for the service Provider project.

Implements [Pact Specification v2](https://github.com/pact-foundation/pact-specification/tree/version-2),
including [flexible matching](http://docs.pact.io/documentation/matching.html).

From the [Ruby Pact website](https://github.com/realestate-com-au/pact):

> Define a pact between service consumers and providers, enabling "consumer driven contract" testing.
>
>Pact provides an RSpec DSL for service consumers to define the HTTP requests they will make to a service provider and the HTTP responses they expect back.
>These expectations are used in the consumers specs to provide a mock service provider. The interactions are recorded, and played back in the service provider
>specs to ensure the service provider actually does provide the response the consumer expects.
>
>This allows testing of both sides of an integration point using fast unit tests.
>
>This gem is inspired by the concept of "Consumer driven contracts". See http://martinfowler.com/articles/consumerDrivenContracts.html for more information.


Read [Getting started with Pact](http://dius.com.au/2016/02/03/microservices-pact/) for more information on
how to get going.


[![wercker status](https://app.wercker.com/status/fad2d928716c629e641cae515ac58547/s/master "wercker status")](https://app.wercker.com/project/bykey/fad2d928716c629e641cae515ac58547)
[![Coverage Status](https://coveralls.io/repos/github/pact-foundation/pact-go/badge.svg?branch=HEAD)](https://coveralls.io/github/pact-foundation/pact-go?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/pact-foundation/pact-go)](https://goreportcard.com/report/github.com/pact-foundation/pact-go)
[![Code Climate](https://codeclimate.com/github/pact-foundation/pact-go/badges/gpa.svg)](https://codeclimate.com/github/pact-foundation/pact-go)
[![Issue Count](https://codeclimate.com/github/pact-foundation/pact-go/badges/issue_count.svg)](https://codeclimate.com/github/pact-foundation/pact-go)
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)

## Installation

* Download a [release](https://github.com/pact-foundation/pact-go/releases) for your OS.
* Unzip the package into a known location, and add to the `PATH`.
* Run `pact-go` to see what options are available

## Running

Due to some design constraints, Pact Go runs a two-step process

1. Run `pact-go daemon` in a separate process/shell. The Consumer and Provider
DSLs communicate over a local (RPC) connection, and is transparent to clients.
1. Create your Pact Consumer/Provider Tests. It defaults to run on port `6666`.

NOTE: The daemon is completely thread safe and it is safe to leave the daemon
running for long periods (e.g. on a CI server).

### Example
1. Start the daemon with `./pact-go daemon`
1. `cd <pact-go>/examples`
1. `go run consumer.go`

```go
import "github.com/pact-foundation/pact-go/dsl"
import ...

func TestSomeApi(t *testing.T) {

	// Create Pact, connecting to local Daemon
	// Ensure the port matches the daemon port!
	pact := &dsl.Pact{
		Port:     6666, 				
		Consumer: "My Consumer",
		Provider: "My Provider",
	}
	// Shuts down Mock Service when done
	defer pact.Teardown()

	// Pass in your test case as a function to Verify()
	var test = func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/", pact.Server.Port))
		return err
	}

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("User Matt exists"). // Provider State
		UponReceiving("A request to login"). // Test Case Name
		WithRequest(&dsl.Request{
			Method: "GET",
			Path:   "/login",
		}).
		WillRespondWith(&dsl.Response{
			Status: 200,
		})

	// Run the test and verify the interactions.
	err := pact.Verify(test)
	if err != nil {
		t.Fatalf("Error on Verify: %v", err)
	}

    // You should now have a pact file in the file `<pact-go>/pacts/my_consumer-my_provider.json`
}
```

### Matching

In addition to verbatim value matching, you have 3 useful matching functions
in the `dsl` package that can increase expressiveness and reduce brittle test
cases.

* `dsl.Term(example, matcher)` - tells Pact that the value should match using
a given regular expression, using `example` in mock responses. `example` must be
a string.
* `dsl.Like(content)` - tells Pact that the value itself is not important, as long
as the element _type_ (valid JSON number, string, object etc.) itself matches.
* `dsl.EachLike(content, min)` - tells Pact that the value should be an array type,
consisting of elements like those passed in. `min` must be >= 1. `content` may
be a valid JSON value: e.g. strings, numbers and objects.

*Example:*

Here is a complex example that shows how all 3 terms can be used together:

```go
jumper := Like(`"jumper"`)
shirt := Like(`"shirt"`)
tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
size := Like(10)
colour := Term("red", "red|green|blue")

match := formatJSON(
	EachLike(
		EachLike(
			fmt.Sprintf(
				`{
					"size": %s,
					"colour": %s,
					"tag": %s
				}`, size, colour, tag),
			1),
		1))
```

This example will result in a response body from the mock server that looks like:
```json
[
  [
    {
      "size": 10,
      "colour": "red",
      "tag": [
        [
          "jumper",
          "shirt"
        ],
        [
          "jumper",
          "shirt"
        ]
      ]
    }
  ]
]
```

See the [matcher tests](https://github.com/pact-foundation/pact-go/blob/master/dsl/matcher_test.go)
for more matching examples.

*NOTE*: One caveat to note, is that you will need to use valid Ruby
[regular expressions](http://ruby-doc.org/core-2.1.5/Regexp.html) and double
escape backslashes.

Read more about [flexible matching](https://github.com/realestate-com-au/pact/wiki/Regular-expressions-and-type-matching-with-Pact).

### Output Logging

Pact Go uses a simple log utility ([logutils](https://github.com/hashicorp/logutils))
to filter log messages. The CLI already contains flags to manage this,
should you want to control log level in your tests, you can set it like so:

```go
pact := &Pact{
  ...
	LogLevel: "DEBUG", // One of DEBUG, INFO, ERROR, NONE
}
```

## Contact

* Twitter: [@pact_up](https://twitter.com/pact_up)
* Google users group: https://groups.google.com/forum/#!forum/pact-support

## Documentation

Additional documentation can be found at the main [Pact website](http://pact.io) and in the [Pact Wiki](https://github.com/realestate-com-au/pact/wiki).

## Developing

For full integration testing locally, Ruby 2.1.5 must be installed. Under the
hood, Pact Go bundles the
[Pact Mock Service](https://github.com/bethesque/pact-mock_service) and
[Pact Provider Verifier](https://github.com/pact-foundation/pact-provider-verifier)
projects to implement up to v2.0 of the Pact Specification. This is only
temporary, until [Pact Reference](https://github.com/pact-foundation/pact-reference/)
work is completed.

* Git clone https://github.com/pact-foundation/pact-go.git
* Run `make dev` to build the package and setup the Ruby 'binaries' locally

## Roadmap

The [roadmap](docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).
