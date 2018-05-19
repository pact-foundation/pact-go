# Pact Go

Golang version of [Pact](http://pact.io). Enables consumer driven contract testing, providing a mock service and
DSL for the consumer project, and interaction playback and verification for the service Provider project.

[![Build Status](https://travis-ci.org/pact-foundation/pact-go.svg?branch=master)](https://travis-ci.org/pact-foundation/pact-go)
[![Build status](https://ci.appveyor.com/api/projects/status/lg02mfcmvr3e8w5n?svg=true)](https://ci.appveyor.com/project/mefellows/pact-go)
[![Coverage Status](https://coveralls.io/repos/github/pact-foundation/pact-go/badge.svg?branch=HEAD)](https://coveralls.io/github/pact-foundation/pact-go?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/pact-foundation/pact-go)](https://goreportcard.com/report/github.com/pact-foundation/pact-go)
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)
[![slack](http://slack.pact.io/badge.svg)](http://slack.pact.io)

## Introduction

From the [Pact website](http://docs.pact.io/):

> The Pact family of frameworks provide support for [Consumer Driven Contracts](http://martinfowler.com/articles/consumerDrivenContracts.html) testing.

> A Contract is a collection of agreements between a client (Consumer) and an API (Provider) that describes the interactions that can take place between them.

> Consumer Driven Contracts is a pattern that drives the development of the Provider from its Consumers point of view.

> Pact is a testing tool that guarantees those Contracts are satisfied.

Read [Getting started with Pact](http://dius.com.au/2016/02/03/microservices-pact/) for more information on
how to get going.

Pact Go implements [Pact Specification v2](https://github.com/pact-foundation/pact-specification/tree/version-2),
including [flexible matching](http://docs.pact.io/documentation/matching.html).

<p align="center">
  <a href="https://asciinema.org/a/180671">
    <img width="880" src="https://raw.githubusercontent.com/pact-foundation/pact-go/master/.github/pact-small.svg?sanitize=true">
  </a>
</p>

## Table of Contents

<!-- TOC -->

- [Pact Go](#pact-go)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Installation on \*nix](#installation-on-\nix)
  - [Using Pact](#using-pact)
  - [HTTP API Testing](#http-api-testing)
    - [Consumer Side Testing](#consumer-side-testing)
    - [Provider API Testing](#provider-api-testing)
      - [Provider Verification](#provider-verification)
      - [API with Authorization](#api-with-authorization)
    - [Publishing pacts to a Pact Broker and Tagging Pacts](#publishing-pacts-to-a-pact-broker-and-tagging-pacts)
      - [Publishing from Go code](#publishing-from-go-code)
      - [Publishing Provider Verification Results to a Pact Broker](#publishing-provider-verification-results-to-a-pact-broker)
      - [Publishing from the CLI](#publishing-from-the-cli)
      - [Using the Pact Broker with Basic authentication](#using-the-pact-broker-with-basic-authentication)
  - [Asynchronous API Testing](#asynchronous-api-testing)
    - [Consumer](#consumer)
    - [Provider (Producer)](#provider-producer)
    - [Pact Broker Integration](#pact-broker-integration)
  - [Matching](#matching)
    - [Matching on types](#matching-on-types)
    - [Matching on arrays](#matching-on-arrays)
    - [Matching by regular expression](#matching-by-regular-expression)
    - [Match common formats](#match-common-formats)
      - [Auto-generate matchers from struct tags](#auto-generate-matchers-from-struct-tags)
  - [Examples](#examples)
    - [HTTP APIs](#http-apis)
    - [Asynchronous APIs](#asynchronous-apis)
    - [Integrated examples](#integrated-examples)
  - [Troubleshooting](#troubleshooting)
      - [Splitting tests across multiple files](#splitting-tests-across-multiple-files)
    - [Output Logging](#output-logging)
  - [Contact](#contact)
  - [Documentation](#documentation)
  - [Troubleshooting](#troubleshooting-1)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)

<!-- /TOC -->

## Installation

*NOTE*: This is the new `1.x.x` release branch with the latest API and feature set (including Message support), for the previous stable version, see the [release/0.x.x](https://github.com/pact-foundation/pact-go/tree/release/0.x.x) branch.

1.  Download the [latest release](https://github.com/pact-foundation/pact-ruby-standalone/releases) of the standalone tools and ensure the binaries are on your `PATH`:
1.  Unzip the package into a known location, and ensuring the `pact-go` binary is on the `PATH`.
1.  Run `go get -d github.com/pact-foundation/pact-go` to install the source packages

See below for how to automate this:

### Installation on \*nix

```sh
cd /opt
curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash
export PATH=$PATH:/opt/pact/bin
go get -d github.com/pact-foundation/pact-go
```

Test the installation:

```sh
pact help
```

## Using Pact

Pact supports [synchronous request-response style HTTP interactions](#http-api-testing) and has experimental support for [asynchronous interactions](#asynchronous-api-testing) with JSON-formatted payloads.

Pact Go runs as part of your regular Go tests.

## HTTP API Testing

### Consumer Side Testing

We'll run through a simple example to get an understanding the concepts:

1.  `go get github.com/pact-foundation/pact-go`
1.  `cd $GOPATH/src/github.com/pact-foundation/pact-go/examples/`
1.  `go test -v -run TestConsumer`.

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
			Headers: dsl.MapMatcher{"Content-Type": "application/json"},
			Body:    `{"s":"foo"}`,
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": "application/json"},
			Body:    `{"s":"bar"}`,
		})

	// Verify
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}
}
```

### Provider API Testing

1.  `go get github.com/pact-foundation/pact-go`
1.  `cd $GOPATH/src/github.com/pact-foundation/pact-go/examples/`
1.  `go test -v -run TestProvider`.

Here is the Provider test process broker down:

1.  Start your Provider API:

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

2.  Verify provider API

    You can now tell Pact to read in your Pact files and verify that your API will
    satisfy the requirements of each of your known consumers:

    ```go
    func TestProvider(t *testing.T) {

      // Create Pact connecting to local Daemon
      pact := &dsl.Pact{
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

1.  Use `PactURLs` to specify the exact set of pacts to be replayed:

    ```go
    pact.VerifyProvider(t, types.VerifyRequest{
    	ProviderBaseURL:        "http://myproviderhost",
    	PactURLs:               []string{"http://broker/pacts/provider/them/consumer/me/latest/dev"},
    	ProviderStatesSetupURL: "http://myproviderhost/setup",
    	BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
    	BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
    })
    ```

1.  Use `PactBroker` to automatically find all of the latest consumers:

    ```go
    pact.VerifyProvider(t, types.VerifyRequest{
    	ProviderBaseURL:        "http://myproviderhost",
    	BrokerURL:              "http://brokerHost",
    	ProviderStatesSetupURL: "http://myproviderhost/setup",
    	BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
    	BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
    })
    ```

1.  Use `PactBroker` and `Tags` to automatically find all of the latest consumers:

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

_Important Note_: You should only use this feature for things that can not be persisted in the pact file. By modifying the request, you are potentially modifying the contract from the consumer tests!

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

## Asynchronous API Testing

Modern distributed architectures are increasingly integrated in a decoupled, asynchronous fashion. Message queues such as ActiveMQ, RabbitMQ, SQS, Kafka and Kinesis are common, often integrated via small and frequent numbers of microservices (e.g. lambda).

Furthermore, the web has things like WebSockets which involve bidirectional messaging.

Pact now has experimental support for these use cases, by abstracting away the protocol and focussing on the messages passing between them.

For further reading and introduction into this topic, see this [article](https://dius.com.au/2017/09/22/contract-testing-serverless-and-asynchronous-applications/)
and our [example](https://github.com/pact-foundation/pact-go/tree/feat/matching-rules-daemonless/examples/messages) for a more detailed overview of these concepts.

### Consumer

A Consumer is the system that will be reading a message from a queue or some intermediary - like a Kinesis stream, websocket or S3 bucket -
and be able to handle it.

From a Pact testing point of view, Pact takes the place of the intermediary and confirms whether or not the consumer is able to handle a request.

The following test creates a contract for a Dog API handler:

```go
// 1 Given this handler that accepts a User and returns an error
userHandler := func(u User) error {
	if u.ID == -1 {
		return errors.New("invalid object supplied, missing fields (id)")
	}

	// ... actually consume the message

	return nil
}

// 2 We write a small adapter that will take the incoming dsl.Message
// and call the function with the correct type
var userHandlerWrapper = func(m dsl.Message) error {
	return userHandler(*m.Content.(*User))
}

// 3 Create the Pact Message Consumer
pact := dsl.Pact {
	return dsl.Pact{
		Consumer:                 "PactGoMessageConsumer",
		Provider:                 "PactGoMessageProvider",
		LogDir:                   logDir,
		PactDir:                  pactDir,
	}
}

// 4 Write the consumer test, and call VerifyMessageConsumer
// passing through the function
func TestMessageConsumer_Success(t *testing.T) {
	message := &dsl.Message{}
	message.
		Given("some state").
		ExpectsToReceive("some test case").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"id":   like(127),
			"name": "Baz",
			"access": eachLike(map[string]interface{}{
				"role": term("admin", "admin|controller|user"),
			}, 3),
    })
    AsType(&User{}) // Optional

	pact.VerifyMessageConsumer(t, message, userHandlerWrapper)
}
```

**Explanation**:

1.  The API - a contrived API handler example. Expects a User object and throws an `Error` if it can't handle it.
    * In most applications, some form of transactionality exists and communication with a MQ/broker happens.
    * It's important we separate out the protocol bits from the message handling bits, so that we can test that in isolation.
1.  Creates the MessageConsumer class
1.  Setup the expectations for the consumer - here we expect a `User` object with three fields
1.  Pact will send the message to your message handler. If the handler does not error, the message is saved, otherwise the test fails. There are a few key things to consider:
    * The actual request body that Pact will invoke on your handler will be contained within a `dsl.Message` object along with other context, so the body must be retrieved via `Content` attribute. If you set `Message.AsType(T)` this object will be mapped for you. If you don't want Pact to perform the conversion, you may do so on the object (`dsl.Message.Content`) or on the raw JSON (`dsl.Message.ContentRaw`).
    * All handlers to be tested must be of the shape `func(dsl.Message) error` - that is, they must accept a `Message` and return an `error`. This is how we get around all of the various protocols, and will often require a lightweight adapter function to convert it.
    * In this case, we wrap the actual `userHandler` with `userHandlerWrapper` provided by Pact.

### Provider (Producer)

A Provider (Producer in messaging parlance) is the system that will be putting a message onto the queue.

As per the Consumer case, Pact takes the position of the intermediary (MQ/broker) and checks to see whether or not the Provider sends a message that matches the Consumer's expectations.

```js
	functionMappings := dsl.MessageProviders{
		"some test case": func(m dsl.Message) (interface{}, error) {
			fmt.Println("Calling provider function that is responsible for creating the message")
			res := User{
				ID:   44,
				Name: "Baz",
				Access: []AccessLevel{
					{Role: "admin"},
					{Role: "admin"},
					{Role: "admin"}},
			}

			return res, nil
		},
	}

	// Verify the Provider with local Pact Files
	pact.VerifyMessageProvider(t, types.VerifyMessageRequest{
		PactURLs: []string{filepath.ToSlash(fmt.Sprintf("%s/pactgomessageconsumer-pactgomessageprovider.json", pactDir))},
	}, functionMappings)
```

**Explanation**:

1.  Our API client contains a single function `createDog` which is responsible for generating the message that will be sent to the consumer via some message queue
1.  We configure Pact to stand-in for the queue. The most important bit here is the `handlers` block
    * Similar to the Consumer tests, we map the various interactions that are going to be verified as denoted by their `description` field. In this case, `a request for a dog`, maps to the `createDog` handler. Notice how this matches the original Consumer test.
1.  We can now run the verification process. Pact will read all of the interactions specified by its consumer, and invoke each function that is responsible for generating that message.

### Pact Broker Integration

As per HTTP APIs, you can [publish contracts and verification results to a Broker](#publishing-pacts-to-a-pact-broker-and-tagging-pacts).

## Matching

In addition to verbatim value matching, we have 3 useful matching functions
in the `dsl` package that can increase expressiveness and reduce brittle test
cases.

Rather than use hard-coded values which must then be present on the Provider side,
you can use regular expressions and type matches on objects and arrays to validate the
structure of your APIs.

Matchers can be used on the `Body`, `Headers`, `Path` and `Query` fields of the `dsl.Request`
type, and the `Body` and `Headers` fields of the `dsl.Response` type.

### Matching on types

`dsl.Like(content)` tells Pact that the value itself is not important, as long
as the element _type_ (valid JSON number, string, object etc.) itself matches.

### Matching on arrays

`dsl.EachLike(content, min)` - tells Pact that the value should be an array type,
consisting of elements like those passed in. `min` must be >= 1. `content` may
be a valid JSON value: e.g. strings, numbers and objects.

### Matching by regular expression

`dsl.Term(example, matcher)` - tells Pact that the value should match using
a given regular expression, using `example` in mock responses. `example` must be
a string. \*

_NOTE_: One caveat to note, is that you will need to use valid Ruby
[regular expressions](http://ruby-doc.org/core-2.1.5/Regexp.html) and double
escape backslashes.

_Example:_

Here is a more complex example that shows how all 3 terms can be used together:

```go
	body :=
		Like(map[string]interface{}{
			"response": map[string]interface{}{
				"name": Like("Billy"),
        "type": Term("admin", "admin|user|guest"),
        "items": EachLike("cat", 2)
			},
		})
```

This example will result in a response body from the mock server that looks like:

```json
{
  "response": {
    "name": "Billy",
    "type": "admin",
    "items": ["cat", "cat"]
  }
}
```

### Match common formats

Often times, you find yourself having to re-write regular expressions for common formats. We've created a number of them for you to save you the time:

| method                                                         | description                                                                                     |
| -------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `Identifier()`                                                 | Match an ID (e.g. 42)                                                                           |
| `Integer()`                                                    | Match all numbers that are integers (both ints and longs)                                       |
| `Decimal()`                                                    | Match all real numbers (floating point and decimal)                                             |
| `HexValue()`                                                   | Match all hexadecimal encoded strings                                                           |
| `Date()`                                                       | Match string containing basic ISO8601 dates (e.g. 2016-01-01)                                   |
| `Timestamp()`                                                  | Match a string containing an RFC3339 formatted timestapm (e.g. Mon, 31 Oct 2016 15:21:41 -0400) |
| `Time()`                                                       | Match string containing times in ISO date format (e.g. T22:44:30.652Z)                          |
| `ipIPv4Address | Match string containing IP4 formatted address |
| `IPv6Address()`                                                | Match string containing IP6 formatted address                                                   |
| `UUID()`                                                       | Match strings containing UUIDs                                                                  |

#### Auto-generate matchers from struct tags

Furthermore, if you isolate your Data Transfer Objects (DTOs) to an adapters package so that they exactly reflect the interface between you and your provider, then you can leverage `dsl.Match` to auto-generate the expected response body in your contract tests. Under the hood, `Match` recursively traverses the DTO struct and uses `Term, Like, and EachLike` to create the contract.

This saves the trouble of declaring the contract by hand. It also maintains one source of truth. To change the consumer-provider interface, you only have to update your DTO struct and the contract will automatically follow suit.

_Example:_

```go
type DTO struct {
  ID    string    `json:"id"`
  Title string    `json:"title"`
  Tags  []string  `json:"tags" pact:"min=2"`
  Date  string    `json:"date" pact:"example=2000-01-01,regex=^\\d{4}-\\d{2}-\\d{2}$"`
}
```

then specifying a response body is as simple as:

```go
	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to get foo").
		WithRequest(dsl.Request{
			Method:  "GET",
			Path:    "/foobar",
			Headers: map[string]string{"Content-Type": "application/json"},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: map[string]string{"Content-Type": "application/json"},
			Body:    Match(DTO{}), // That's it!!!
		})
```

The `pact` struct tags shown above are optional. By default, dsl.Match just asserts that the JSON shape matches the struct and that the field types match.

See [dsl.Match](https://github.com/pact-foundation/pact-go/blob/master/dsl/matcher.go) for more information.

See the [matcher tests](https://github.com/pact-foundation/pact-go/blob/master/dsl/matcher_test.go)
for more matching examples.

## Examples

### HTTP APIs

* [API Consumer](https://github.com/pact-foundation/pact-go/tree/master/examples/)
* [Golang ServeMux](https://github.com/pact-foundation/pact-go/tree/master/examples/mux)
* [Go Kit](https://github.com/pact-foundation/pact-go/tree/master/examples/go-kit)
* [Gin](https://github.com/pact-foundation/pact-go/tree/master/examples/gin)

### Asynchronous APIs

* [Message Queue](https://github.com/pact-foundation/pact-go/tree/feat/matching-rules-daemonless/examples/messages)

### Integrated examples

There are number of examples we use as end-to-end integration test prior to releasing a new binary, including publishing to a Pact Broker. To enable them, set the following environment variables

```sh
cd $GOPATH/src/github.com/pact-foundation/pact-go/examples
export PACT_INTEGRATED_TESTS=1
export PACT_BROKER_USERNAME="dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
export PACT_BROKER_PASSWORD="O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"
export PACT_BROKER_HOST="https://test.pact.dius.com.au"
```

Once these variables have been exported, cd into one of the directories containing a test and run `go test -v .`:

## Troubleshooting

#### Splitting tests across multiple files

Pact tests tend to be quite long, due to the need to be specific about request/response payloads. Often times it is nicer to be able to split your tests across multiple files for manageability.

You have two options to achieve this feat:

1.  Set `PactFileWriteMode` to `"merge"` when creating a `Pact` struct:

    This will allow you to have multiple independent tests for a given Consumer-Provider pair, without it clobbering previous interactions.

    See this [PR](https://github.com/pact-foundation/pact-js/pull/48) for background.

    _NOTE_: If using this approach, you _must_ be careful to clear out existing pact files (e.g. `rm ./pacts/*.json`) before you run tests to ensure you don't have left over requests that are no longer relevent.

1.  Create a Pact test helper to orchestrate the setup and teardown of the mock service for multiple tests.

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

## Contact

Join us in slack: [![slack](http://slack.pact.io/badge.svg)](http://slack.pact.io)

or

* Twitter: [@pact_up](https://twitter.com/pact_up)
* Stack Overflow: stackoverflow.com/questions/tagged/pact
* Gitter: https://gitter.im/realestate-com-au/pact
* Gophers #pact [Slack channel](https://gophers.slack.com/messages/pact/)

## Documentation

Additional documentation can be found at the main [Pact website](http://pact.io) and in the [Pact Wiki](https://github.com/pact-foundation/pact-ruby/wiki).

## Troubleshooting

See [TROUBLESHOOTING](https://github.com/pact-foundation/pact-go/wiki/Troubleshooting) for some helpful tips/tricks.

## Roadmap

The [roadmap](https://docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.
Detail on the native Go implementation can be found [here](https://github.com/pact-foundation/pact-go/wiki/Native-implementation-roadmap).

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).
