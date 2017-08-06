# Pact Go

Golang version of [Pact](http://pact.io). Enables consumer driven contract testing, providing a mock service and
DSL for the consumer project, and interaction playback and verification for the service Provider project.

[![wercker status](https://app.wercker.com/status/fad2d928716c629e641cae515ac58547/s/master "wercker status")](https://app.wercker.com/project/bykey/fad2d928716c629e641cae515ac58547)
[![Coverage Status](https://coveralls.io/repos/github/pact-foundation/pact-go/badge.svg?branch=HEAD)](https://coveralls.io/github/pact-foundation/pact-go?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/pact-foundation/pact-go)](https://goreportcard.com/report/github.com/pact-foundation/pact-go)
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)
[![Build status](https://ci.appveyor.com/api/projects/status/lg02mfcmvr3e8w5n?svg=true)](https://ci.appveyor.com/project/mefellows/pact-go)

## Introduction

From the [Pact website](http://docs.pact.io/):

>The Pact family of frameworks provide support for [Consumer Driven Contracts](http://martinfowler.com/articles/consumerDrivenContracts.html) testing.

>A Contract is a collection of agreements between a client (Consumer) and an API (Provider) that describes the interactions that can take place between them.

>Consumer Driven Contracts is a pattern that drives the development of the Provider from its Consumers point of view.

>Pact is a testing tool that guarantees those Contracts are satisfied.

Read [Getting started with Pact](http://dius.com.au/2016/02/03/microservices-pact/) for more information on
how to get going.

Pact Go implements [Pact Specification v2](https://github.com/pact-foundation/pact-specification/tree/version-2),
including [flexible matching](http://docs.pact.io/documentation/matching.html).

[![asciicast](https://asciinema.org/a/121445.png)](https://asciinema.org/a/121445)

## Table of Contents

<!-- TOC depthFrom:2 depthTo:6 withLinks:1 updateOnSave:1 orderedList:1 -->

1. [Table of Contents](#table-of-contents)
2. [Installation](#installation)
3. [Running](#running)
	1. [Consumer](#consumer)
		1. [Matching (Consumer Tests)](#matching-consumer-tests)
	2. [Provider](#provider)
		1. [Provider States](#provider-states)
	3. [Publishing Pacts to a Broker and Tagging Pacts](#publishing-pacts-to-a-broker-and-tagging-pacts)
		1. [Publishing from Go code](#publishing-from-go-code)
		2. [Publishing from the CLI](#publishing-from-the-cli)
        4. [Publishing Verification Results to a Pact Broker](#publishing-verification-results-to-a-pact-broker)
	5. [Using the Pact Broker with Basic authentication](#using-the-pact-broker-with-basic-authentication)
	6. [Output Logging](#output-logging)
4. [Examples](#examples)
5. [Contact](#contact)
6. [Documentation](#documentation)
7. [Roadmap](#roadmap)
8. [Contributing](#contributing)

<!-- /TOC -->

## Installation

* Download one of the zipped [release](https://github.com/pact-foundation/pact-go/releases) distributions for your OS.
* Unzip the package into a known location, and rename `pact-go_<os/arch>` to `pact-go` and ensuring it is on the `PATH`.
* Run `pact-go` to see what options are available.
* Run `go get -d github.com/pact-foundation/pact-go` to install the source packages

## Running

Pact Go runs a two-step process:

1. Run `pact-go daemon` in a separate process/shell. The Consumer and Provider
DSLs communicate over a local (RPC) connection, and is transparent to clients.
1. Create your Pact Consumer/Provider Tests. It defaults to run on port `6666`.

*NOTE: The daemon is thread safe and it is normal to leave it
running for long periods (e.g. on a CI server).*

### Consumer
1. Start the daemon with `./pact-go daemon`.
1. go get github.com/pact-foundation/pact-go
1. cd $GOPATH/src/github.com/pact-foundation/pact-go/examples
1. `cd <pact-go>/examples`.
1. `go run -v consumer.go`.

Here is a simple example (`consumer_test.go`) you can run with `go test -v .`:

```go
package somepackage

import (
	"fmt"
	"github.com/pact-foundation/pact-go/dsl"
	"net/http"
	"testing"
)

func TestLogin(t *testing.T) {

	// Create Pact, connecting to local Daemon
	// Ensure the port matches the daemon port!
	pact := dsl.Pact{
		Port:     6666,
		Consumer: "MyConsumer",
		Provider: "MyProvider",
	}
	// Shuts down Mock Service when done
	defer pact.Teardown()

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("User Matt exists").           // Provider State
		UponReceiving("A request to login"). // Test Case Name
		WithRequest(dsl.Request{
			Method: "GET",
			Path:   "/login",
			Body: `{"username":"matt"}`,
		}).
		WillRespondWith(dsl.Response{
			Status: 200,
			Body: `{"username":"matt", "id":1234}`,
		})

	// Run the test and verify the interactions.
	if err := pact.Verify(func() error {
		u := fmt.Sprintf("http://localhost:%d/login", pact.Server.Port)
		req, err := http.NewRequest("GET", u, strings.NewReader(`{"username":"matt"}`))

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			return err
		}
		if _, err = http.DefaultClient.Do(req); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// Write pact to file `<pact-go>/pacts/my_consumer-my_provider.json`
	pact.WritePact()
}

```


#### Matching (Consumer Tests)

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

match := EachLike(
              EachLike(
                         fmt.Sprintf(`{
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


### Provider

1. Start your Provider API:

	```go
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
	})
	mux.HandleFunc("/someapi", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
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
			]`)
	})
	go http.ListenAndServe(":8000", mux)
	```

	Note that the server has and endpoint:  `/setup` that allows the
	verifier to setup any 
	[provider states](http://docs.pact.io/documentation/provider_states.html) before
	each test is run.

2. Verify provider API

	You can now tell Pact to read in your Pact files and verify that your API will
	satisfy the requirements of each of your known consumers:

	```go
	response := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        "http://localhost:8000",
		PactURLs:               []string{"./pacts/my_consumer-my_provider.json"},		
		ProviderStatesSetupURL: "http://localhost:8000/setup",
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
	```

	Note that `PactURLs` is a list of local pact files or remote based
	urls (e.g. from a
	[Pact Broker](http://docs.pact.io/documentation/sharings_pacts.html)).


	See the `Skip()'ed` [integration tests](https://github.com/pact-foundation/pact-go/blob/master/dsl/pact_test.go)
	for a more complete E2E example.

#### Provider Verification

When validating a Provider, you have 3 options to provide the Pact files:

1. Use `PactURLs` to specify the exact set of pacts to be replayed:

	```go
	response = pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        "http://myproviderhost",
		PactURLs:               []string{"http://broker/pacts/provider/them/consumer/me/latest/dev"},		
		ProviderStatesSetupURL: "http://myproviderhost/setup",
		BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
	})
	```
1. Use `PactBroker` to automatically find all of the latest consumers:

	```go
	response = pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        "http://myproviderhost",
		BrokerURL:              "http://brokerHost",		
		ProviderStatesSetupURL: "http://myproviderhost/setup",
		BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
	})
	```
1. Use `PactBroker` and `Tags` to automatically find all of the latest consumers:

	```go
	response = pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL:        "http://myproviderhost",
		BrokerURL:              "http://brokerHost",
		Tags:                   []string{"latest", "sit4"},		
		ProviderStatesSetupURL: "http://myproviderhost/setup",
		BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
	})
	```

Options 2 and 3 are particularly useful when you want to validate that your
Provider is able to meet the contracts of what's in Production and also the latest
in development.

See this [article](http://rea.tech/enter-the-pact-matrix-or-how-to-decouple-the-release-cycles-of-your-microservices/)
for more on this strategy.
e
```

See the examples or read more at http://docs.pact.io/documentation/provider_states.html.

#### Publishing Verification Results to a Pact Broker

If you're using a Pact Broker (e.g. a hosted one at pact.dius.com.au), you can
publish your verification results so that consumers can query if they are safe
to release.

It looks like this:

![screenshot of verification result](https://cloud.githubusercontent.com/assets/53900/25884085/2066d98e-3593-11e7-82af-3b41a20af8e5.png)

You need to specify the following:

```go
PublishVerificationResults: true,
ProviderVersion:            "1.0.0",
```

_NOTE_: You need to be already pulling pacts from the broker for this feature to work.
### Publishing Pacts to a Broker and Tagging Pacts

See the [Pact Broker](http://docs.pact.io/documentation/sharings_pacts.html)
documentation for more details on the Broker and this [article](http://rea.tech/enter-the-pact-matrix-or-how-to-decouple-the-release-cycles-of-your-microservices/)
on how to make it work for you.

#### Publishing from Go code

```go
p := Publisher{}
err := p.Publish(types.PublishRequest{
	PactURLs:	[]string{"./pacts/my_consumer-my_provider.json"},
	PactBroker:	"http://pactbroker:8000",
	ConsumerVersion: "1.0.0",
	Tags:		[]string{"latest", "dev"},
})
```

#### Publishing from the CLI

Use a cURL request like the following to PUT the pact to the right location,
specifying your consumer name, provider name and consumer version.

```
curl -v -XPUT \-H "Content-Type: application/json" \
-d@spec/pacts/a_consumer-a_provider.json \
http://your-pact-broker/pacts/provider/A%20Provider/consumer/A%20Consumer/version/1.0.0
```

### Using the Pact Broker with Basic authentication

The following flags are required to use basic authentication when
publishing or retrieving Pact files to/from a Pact Broker:

* `BrokerUsername` - the username for Pact Broker basic authentication.
* `BrokerPassword` - the password for Pact Broker basic authentication.

### Splitting tests across multiple files

Pact tests tend to be quite long, due to the need to be specific about request/response payloads. Often times it is nicer to be able to split your tests across multiple files for manageability.

You have two options to achieve this feat:

1. Set `PactFileWriteMode` to `"update"` when creating a `Pact` struct:

    This will allow you to have multiple independent tests for a given Consumer-Provider pair, without it clobbering previous interactions.

    See this [PR](https://github.com/pact-foundation/pact-js/pull/48) for background.

    _NOTE_: If using this approach, you *must* be careful to clear out existing pact files (e.g. `rm ./pacts/*.json`) before you run tests to ensure you don't have left over requests that are no longer relevent.

1. Create a Pact test helper to orchestrate the setup and teardown of the mock service for multiple tests.

    In larger test bases, this can reduce test suite time and the amount of code you have to manage.

    See the JS [example](https://github.com/tarciosaraiva/pact-melbjs/blob/master/helper.js) and related [issue](https://github.com/pact-foundation/pact-js/issues/11) for more.

### Output Logging

Pact Go uses a simple log utility ([logutils](https://github.com/hashicorp/logutils))
to filter log messages. The CLI already contains flags to manage this,
should you want to control log level in your tests, you can set it like so:

```go
pact := Pact{
  ...
	LogLevel: "DEBUG", // One of DEBUG, INFO, ERROR, NONE
}
```

## Examples

There is a single file, end-to-end integration test we use as a smoke test before releasing a new binary, including publishing to a broker, you can run it (after starting the daemon) as follows:

```sh
cd dsl
export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
go test -run TestPact_Integration
```

*Other examples:*

* [API Consumer](https://github.com/pact-foundation/pact-go/tree/master/examples/)
* [Golang ServeMux](https://github.com/pact-foundation/pact-go/tree/master/examples/mux)
* [Go Kit](https://github.com/pact-foundation/pact-go/tree/master/examples/go-kit)
* [Gin](https://github.com/pact-foundation/pact-go/tree/master/examples/gin)

## Contact

* Twitter: [@pact_up](https://twitter.com/pact_up)
* Stack Overflow: stackoverflow.com/questions/tagged/pact
* Gitter: https://gitter.im/realestate-com-au/pact
* Gophers #pact [Slack channel](https://gophers.slack.com/messages/pact/)

## Documentation

Additional documentation can be found at the main [Pact website](http://pact.io) and in the [Pact Wiki](https://github.com/realestate-com-au/pact/wiki).

## Troubleshooting

See [TROUBLESHOOTING](https://github.com/pact-foundation/pact-go/wiki/Troubleshooting) for some helpful tips/tricks.

## Roadmap

The [roadmap](docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.
Detail on the native Go implementation can be found [here](https://github.com/pact-foundation/pact-go/wiki/Native-implementation-roadmap).


## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).
