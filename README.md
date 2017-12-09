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

<!-- TOC -->

- [Pact Go](#pact-go)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Running](#running)
    - [Consumer](#consumer)
      - [Matching (Consumer Tests)](#matching-consumer-tests)
    - [Provider](#provider)
      - [Provider Verification](#provider-verification)
      - [API with Authorization](#api-with-authorization)
    - [Publishing pacts to a Pact Broker and Tagging Pacts](#publishing-pacts-to-a-pact-broker-and-tagging-pacts)
      - [Publishing from Go code](#publishing-from-go-code)
      - [Publishing Provider Verification Results to a Pact Broker](#publishing-provider-verification-results-to-a-pact-broker)
      - [Publishing from the CLI](#publishing-from-the-cli)
      - [Using the Pact Broker with Basic authentication](#using-the-pact-broker-with-basic-authentication)
    - [Troubleshooting](#troubleshooting)
      - [Splitting tests across multiple files](#splitting-tests-across-multiple-files)
      - [Output Logging](#output-logging)
  - [Examples](#examples)
  - [Contact](#contact)
  - [Documentation](#documentation)
  - [Troubleshooting](#troubleshooting-1)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)

<!-- /TOC -->

## Installation

* Download one of the zipped [release](https://github.com/pact-foundation/pact-go/releases) distributions for your OS.
* Unzip the package into a known location, and ensuring the `pact-go` binary is on the `PATH`, next to the `pact` folder.
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

We'll run through a simple example to get an understanding the concepts:

1. Start the daemon with `./pact-go daemon`.
1. `go get github.com/pact-foundation/pact-go`
1. `cd $GOPATH/src/github.com/pact-foundation/pact-go/examples/`
1. `go test -v -run TestConsumer`.

The simple example looks like this:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
)

// Example Pact: How to run me!
// 1. Start the daemon with `./pact-go daemon`
// 2. cd <pact-go>/examples
// 3. go test -v -run TestConsumer
func TestConsumer(t *testing.T) {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Port:     6666, // Ensure this port matches the daemon port!
		Consumer: "MyConsumer",
		Provider: "MyProvider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		u := fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port)
		req, err := http.NewRequest("GET", u, strings.NewReader(`{"s":"foo"}`))

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

		return err
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to get foo").
		WithRequest(dsl.Request{
			Method:  "GET",
			Path:    "/foobar",
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    `{"s":"foo"}`,
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    `{"s":"bar"}`,
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
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
colour := Term("red", "red|green|blue")

match := EachLike(
              EachLike(
                         fmt.Sprintf(`{
                             "size": 10,
                             "colour": %s,
                             "tag": [["jumper", "shirt]]
                         }`, colour)
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

1. Start the daemon with `./pact-go daemon`.
1. `go get github.com/pact-foundation/pact-go`
1. `cd $GOPATH/src/github.com/pact-foundation/pact-go/examples/`
1. `go test -v -run TestProvider`.

Here is the Provider test process broker down:

1. Start your Provider API:

    You need to be able to first start your API in the background as part of your tests
    before you can run the verification process. Here we create `startServer` which can be 
    started in its own goroutine:

    ```go
    func startServer() {
      mux := http.NewServeMux()
      lastName := "billy"

      mux.HandleFunc("/foobar", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        fmt.Fprintf(w, fmt.Sprintf(`{"lastName":"%s"}`, lastName))

        // Break the API by replacing the above and uncommenting one of these
        // w.WriteHeader(http.StatusUnauthorized)
        // fmt.Fprintf(w, `{"s":"baz"}`)
      })

      // This function handles state requests for a particular test
      // In this case, we ensure that the user being requested is available
      // before the Verification process invokes the API.
      mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
        var s *types.ProviderState
        decoder := json.NewDecoder(req.Body)
        decoder.Decode(&s)
        if s.State == "User foo exists" {
          lastName = "bar"
        }

        w.Header().Add("Content-Type", "application/json")
      })
      go http.ListenAndServe(":8000", mux)
    }
    ```

  Note that the server has a `/setup` endpoint that is given a `types.ProviderState` and allows the
  verifier to setup any 
  [provider states](http://docs.pact.io/documentation/provider_states.html) before
  each test is run.

2. Verify provider API

	You can now tell Pact to read in your Pact files and verify that your API will
	satisfy the requirements of each of your known consumers:

	```go
    func TestProvider(t *testing.T) {

      // Create Pact connecting to local Daemon
      pact := &dsl.Pact{
        Port:     6666, // Ensure this port matches the daemon port!
        Consumer: "MyConsumer",
        Provider: "MyProvider",
      }

      // Start provider API in the background
      go startServer()

      // Verify the Provider with local Pact Files
      pact.VerifyProvider(t, types.VerifyRequest{
        ProviderBaseURL:        "http://localhost:8000",
        PactURLs:               []string{filepath.ToSlash(fmt.Sprintf("%s/myconsumer-myprovider.json", pactDir))},
        ProviderStatesSetupURL: "http://localhost:8000/setup",
      })
    }
	```

  The `VerifyProvider` will handle all verifications, treating them as subtests
  and giving you granular test reporting. If you don't like this behaviour, you may call `VerifyProviderRaw` directly and handle the errors manually.

  Note that `PactURLs` may be a list of local pact files or remote based
  urls (e.g. from a
  [Pact Broker](http://docs.pact.io/documentation/sharings_pacts.html)).

  See the `Skip()'ed` [integration tests](https://github.com/pact-foundation/pact-go/blob/master/dsl/pact_test.go)
  for a more complete E2E example.

#### Provider Verification

When validating a Provider, you have 3 options to provide the Pact files:

1. Use `PactURLs` to specify the exact set of pacts to be replayed:

	```go
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        "http://myproviderhost",
		PactURLs:               []string{"http://broker/pacts/provider/them/consumer/me/latest/dev"},		
		ProviderStatesSetupURL: "http://myproviderhost/setup",
		BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
	})
	```
1. Use `PactBroker` to automatically find all of the latest consumers:

	```go
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        "http://myproviderhost",
		BrokerURL:              "http://brokerHost",		
		ProviderStatesSetupURL: "http://myproviderhost/setup",
		BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
	})
	```
1. Use `PactBroker` and `Tags` to automatically find all of the latest consumers:

	```go
	pact.VerifyProvider(t, types.VerifyRequest{
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


For more on provider states, refer to http://docs.pact.io/documentation/provider_states.html.

#### API with Authorization

Sometimes you may need to add things to the requests that can't be persisted in a pact file. Examples of these would be authentication tokens, which have a small life span. e.g. an OAuth bearer token: `Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42`.

For this case, we have a facility that should be carefully used during verification - the ability to specificy custom headers to be sent during provider verification. The property to achieve this is `CustomProviderHeaders`.

For example, to have an `Authorization` header sent as part of the verification request, modify the `VerifyRequest` parameter as per below:

```go
  pact.VerifyProvider(t, types.VerifyRequest{
    ...
    CustomProviderHeaders:  []string{"Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42"},
  })
```

As you can see, this is your opportunity to modify\add to headers being sent to the Provider API, for example to create a valid time-bound token.

*Important Note*: You should only use this feature for things that can not be persisted in the pact file. By modifying the request, you are potentially modifying the contract from the consumer tests!

### Publishing pacts to a Pact Broker and Tagging Pacts

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

#### Publishing Provider Verification Results to a Pact Broker

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

#### Publishing from the CLI

Use a cURL request like the following to PUT the pact to the right location,
specifying your consumer name, provider name and consumer version.

```
curl -v \
  -X PUT \
  -H "Content-Type: application/json" \
  -d@spec/pacts/a_consumer-a_provider.json \
  http://your-pact-broker/pacts/provider/A%20Provider/consumer/A%20Consumer/version/1.0.0
```

#### Using the Pact Broker with Basic authentication

The following flags are required to use basic authentication when
publishing or retrieving Pact files to/from a Pact Broker:

* `BrokerUsername` - the username for Pact Broker basic authentication.
* `BrokerPassword` - the password for Pact Broker basic authentication.

### Troubleshooting

#### Splitting tests across multiple files

Pact tests tend to be quite long, due to the need to be specific about request/response payloads. Often times it is nicer to be able to split your tests across multiple files for manageability.

You have two options to achieve this feat:

1. Set `PactFileWriteMode` to `"update"` when creating a `Pact` struct:

    This will allow you to have multiple independent tests for a given Consumer-Provider pair, without it clobbering previous interactions.

    See this [PR](https://github.com/pact-foundation/pact-js/pull/48) for background.

    _NOTE_: If using this approach, you *must* be careful to clear out existing pact files (e.g. `rm ./pacts/*.json`) before you run tests to ensure you don't have left over requests that are no longer relevent.

1. Create a Pact test helper to orchestrate the setup and teardown of the mock service for multiple tests.

    In larger test bases, this can reduce test suite time and the amount of code you have to manage.

    See the JS [example](https://github.com/tarciosaraiva/pact-melbjs/blob/master/helper.js) and related [issue](https://github.com/pact-foundation/pact-js/issues/11) for more.

#### Output Logging

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

There is a number of examples we use as end-to-end integration test prior to releasing a new binary, including publishing to a Pact Broker. You can run them all by running `make pact` in the project root, or manually (after starting the daemon) as follows:

```sh
cd $GOPATH/src/github.com/pact-foundation/pact-go/examples
export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
```

Once these variables have been exported, cd into one of the directories containing a test and run `go test -v .`:

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

The [roadmap](https://docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.
Detail on the native Go implementation can be found [here](https://github.com/pact-foundation/pact-go/wiki/Native-implementation-roadmap).


## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).
