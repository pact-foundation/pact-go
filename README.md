<!-- Please use absolute URLs for all links so that they work when the content of this page is rendered in docs.pact.io -->

# Pact Go

Golang version of [Pact](http://pact.io). Pact is a contract testing framework for HTTP APIs and non-HTTP asynchronous messaging systems.

Enables consumer driven contract testing, providing a mock service and
DSL for the consumer project, and interaction playback and verification for the service Provider project.

[![Build Status](https://travis-ci.org/pact-foundation/pact-go.svg?branch=master)](https://travis-ci.org/pact-foundation/pact-go)
[![Build status](https://ci.appveyor.com/api/projects/status/lg02mfcmvr3e8w5n?svg=true)](https://ci.appveyor.com/project/mefellows/pact-go)
[![Coverage Status](https://coveralls.io/repos/github/pact-foundation/pact-go/badge.svg?branch=HEAD)](https://coveralls.io/github/pact-foundation/pact-go?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/pact-foundation/pact-go)](https://goreportcard.com/report/github.com/pact-foundation/pact-go)
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)
[![slack](https://slack.pact.io/badge.svg)](https://slack.pact.io)
[![Known Vulnerabilities](https://snyk.io/test/github/pact-foundation/pact-go/badge.svg?targetFile=Gopkg.lock)](https://snyk.io/test/github/pact-foundation/pact-go?targetFile=Gopkg.lock)

## Introduction

From the [Pact website]:

> The Pact family of frameworks provide support for [Consumer Driven Contracts](http://martinfowler.com/articles/consumerDrivenContracts.html) testing.

> A Contract is a collection of agreements between a client (Consumer) and an API (Provider) that describes the interactions that can take place between them.

> Consumer Driven Contracts is a pattern that drives the development of the Provider from its Consumers point of view.

> Pact is a testing tool that guarantees those Contracts are satisfied.

Read [Getting started with Pact] for more information for beginners.

<p align="center">
  <a href="https://asciinema.org/a/180671">
    <img width="880" src="https://raw.githubusercontent.com/pact-foundation/pact-go/master/.github/pact-small.svg?sanitize=true"></img>
  </a>
</p>

## Table of Contents

<!-- TOC -->

- [Pact Go](#pact-go)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [Versions](#versions)
  - [Installation](#installation)
    - [Go get](#go-get)
    - [Installation on \*nix](#installation-on-nix)
  - [V3 Beta](#v3-beta)
  - [Using Pact](#using-pact)
  - [HTTP API Testing](#http-api-testing)
    - [Consumer Side Testing](#consumer-side-testing)
    - [Provider API Testing](#provider-api-testing)
      - [Provider Verification](#provider-verification)
      - [Provider States](#provider-states)
      - [Before and After Hooks](#before-and-after-hooks)
      - [Request Filtering](#request-filtering)
        - [Example: API with Authorization](#example-api-with-authorization)
      - [Pending Pacts](#pending-pacts)
      - [WIP Pacts](#wip-pacts)
      - [Lifecycle of a provider verification](#lifecycle-of-a-provider-verification)
    - [Publishing pacts to a Pact Broker and Tagging Pacts](#publishing-pacts-to-a-pact-broker-and-tagging-pacts)
      - [Publishing from Go code](#publishing-from-go-code)
      - [Publishing Provider Verification Results to a Pact Broker](#publishing-provider-verification-results-to-a-pact-broker)
      - [Publishing from the CLI](#publishing-from-the-cli)
      - [Using the Pact Broker with Basic authentication](#using-the-pact-broker-with-basic-authentication)
      - [Using the Pact Broker with Bearer Token authentication](#using-the-pact-broker-with-bearer-token-authentication)
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
  - [Tutorial (60 minutes)](#tutorial-60-minutes)
  - [Examples](#examples)
    - [HTTP APIs](#http-apis)
    - [Asynchronous APIs](#asynchronous-apis)
    - [Integrated examples](#integrated-examples)
  - [Troubleshooting](#troubleshooting) - [Splitting tests across multiple files](#splitting-tests-across-multiple-files) - [Output Logging](#output-logging) - [Check if the CLI tools are up to date](#check-if-the-cli-tools-are-up-to-date) - [Disable CLI checks to speed up tests](#disable-cli-checks-to-speed-up-tests) - [Re-run a specific provider verification test](#re-run-a-specific-provider-verification-test)
    - [Verifying APIs with a self-signed certificate](#verifying-apis-with-a-self-signed-certificate)
    - [Testing AWS API Gateway APIs](#testing-aws-api-gateway-apis)
  - [Contact](#contact)
  - [Documentation](#documentation)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)

<!-- /TOC -->

## Versions

<details><summary>Specification Compatibility</summary>

| Version | Stable | [Spec] Compatibility | Install            |
| ------- | ------ | -------------------- | ------------------ |
| 2.x.x   | Beta   | 2, 3                 | See [v2.x.x]       |
| 1.0.x   | Yes    | 2, 3\*               | See [installation] |
| 0.x.x   | Yes    | Up to v2             | 0.x.x [stable]     |

_\*_ v3 support is limited to the subset of functionality required to enable language inter-operable [Message support].

</details>

## Installation

1.  Download the latest [CLI tools] of the standalone tools and ensure the binaries are on your `PATH`:
1.  Unzip the package into a known location, and ensuring the `pact` and other binaries in the `bin` directory are on the `PATH`.
1.  Run `go get github.com/pact-foundation/pact-go@v1` to install the source packages

See below for how to automate this:

### Go get

Since `1.x.x` Pact is go-gettable, and uses tags for versioning, so `dep ensure --add github.com/pact-foundation/pact-go@1.0.0` or `go get gopkg.in/pact-foundation/pact-go.v1` is now possible.

See the [Changelog] for versions to pin to and their history.

### Installation on \*nix

The following will install pact binaries into `/opt/pact/bin`:

```sh
cd /opt
curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash
export PATH=$PATH:/opt/pact/bin
go get github.com/pact-foundation/pact-go@v1
```

Test the installation:

```sh
pact help
```

_NOTE_: the above script installs the latest standalone tools at the time it was ran. It is recommended you pin the installation to a [specific version](https://github.com/pact-foundation/pact-ruby-standalone/releases) of a release so that you may control the upgrade cycle.

## V3 Beta

If you are interested in testing out the new new beta package that supports all of the V3 [spec], and moves to a rust shared core, please head to [v2.x.x] and also let us know on [slack].

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
func TestConsumer(t *testing.T) {
	type User struct {
		Name     string `json:"name" pact:"example=billy"`
		LastName string `json:"lastName" pact:"example=sampson"`
	}

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "MyConsumer",
		Provider: "MyProvider",
		Host:     "localhost",
	}
	defer pact.Teardown()

	// Pass in test case. This is the component that makes the external HTTP call
	var test = func() (err error) {
		u := fmt.Sprintf("http://localhost:%d/foobar", pact.Server.Port)
		req, err := http.NewRequest("GET", u, strings.NewReader(`{"name":"billy"}`))
		if err != nil {
			return
		}

		// NOTE: by default, request bodies are expected to be sent with a Content-Type
		// of application/json. If you don't explicitly set the content-type, you
		// will get a mismatch during Verification.
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer 1234")

		_, err = http.DefaultClient.Do(req)
		return
	}

	// Set up our expected interactions.
	pact.
		AddInteraction().
		Given("User foo exists").
		UponReceiving("A request to get foo").
		WithRequest(dsl.Request{
			Method:  "GET",
			Path:    dsl.String("/foobar"),
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json"), "Authorization": dsl.String("Bearer 1234")},
			Body: map[string]string{
				"name": "billy",
			},
		}).
		WillRespondWith(dsl.Response{
			Status:  200,
			Headers: dsl.MapMatcher{"Content-Type": dsl.String("application/json")},
			Body:    dsl.Match(&User{}),
		})

	// Run the test, verify it did what we expected and capture the contract
	if err := pact.Verify(test); err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

	return nil
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
    var lastName = "" // User doesn't exist
    func startServer() {
      mux := http.NewServeMux()

      mux.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
        w.Header().Add("Content-Type", "application/json")
        fmt.Fprintf(w, fmt.Sprintf(`{"lastName":"%s"}`, lastName))
      })
      log.Fatal(http.ListenAndServe(":8000", mux))
    }
    ```

2.  Verify provider API

    You can now tell Pact to read in your Pact files and verify that your API will
    satisfy the requirements of each of your known consumers:

    ```go
    func TestProvider(t *testing.T) {

      // Create Pact connecting to local Daemon
      pact := &dsl.Pact{
        Provider: "MyProvider",
      }

      // Start provider API in the background
      go startServer()

      // Verify the Provider using the locally saved Pact Files
      pact.VerifyProvider(t, types.VerifyRequest{
        ProviderBaseURL:        "http://localhost:8000",
        PactURLs:               []string{filepath.ToSlash(fmt.Sprintf("%s/myconsumer-myprovider.json", pactDir))},
        StateHandlers: types.StateHandlers{
          // Setup any state required by the test
          // in this case, we ensure there is a "user" in the system
          "User foo exists": func() error {
            lastName = "crickets"
            return nil
          },
        },
      })
    }
    ```

The `VerifyProvider` will handle all verifications, treating them as subtests
and giving you granular test reporting. If you don't like this behaviour, you may call `VerifyProviderRaw` directly and handle the errors manually.

Note that `PactURLs` may be a list of local pact files or remote based
urls (e.g. from a
[Pact Broker](https://docs.pact.io/pact_broker)).

#### Provider Verification

When validating a Provider, you have 3 options to provide the Pact files:

1.  Use `PactURLs` to specify the exact set of pacts to be replayed:

    ```go
    pact.VerifyProvider(t, types.VerifyRequest{
    	ProviderBaseURL:        "http://myproviderhost",
    	PactURLs:               []string{"http://broker/pacts/provider/them/consumer/me/latest/dev"},
    	BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
    	BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
    	BrokerToken:            os.Getenv("PACT_BROKER_TOKEN"),
    })
    ```

1.  Use `BrokerURL` to automatically find all of the latest consumers:

    ```go
    pact.VerifyProvider(t, types.VerifyRequest{
    	ProviderBaseURL:        "http://myproviderhost",
    	BrokerURL:              "http://brokerHost",
    	BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
    	BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
    	BrokerToken:            os.Getenv("PACT_BROKER_TOKEN"),
    })
    ```

1.  Use `BrokerURL` and `Tags` to automatically find all of the latest consumers given one or more tags:

    ```go
    pact.VerifyProvider(t, types.VerifyRequest{
    	ProviderBaseURL:        "http://myproviderhost",
    	BrokerURL:              "http://brokerHost",
    	Tags:                   []string{"master", "prod"},
    	BrokerUsername:         os.Getenv("PACT_BROKER_USERNAME"),
    	BrokerPassword:         os.Getenv("PACT_BROKER_PASSWORD"),
    	BrokerToken:            os.Getenv("PACT_BROKER_TOKEN"),
    })
    ```

Options 2 and 3 are particularly useful when you want to validate that your
Provider is able to meet the contracts of what's in Production and also the latest
in development.

See this [article](http://rea.tech/enter-the-pact-matrix-or-how-to-decouple-the-release-cycles-of-your-microservices/)
for more on this strategy.

#### Provider States

If you have defined any states (as denoted by a `Given()`) in your consumer tests, the `Verifier` can put the provider into the correct state prior to sending the actual request for validation. For example, the provider can use the state to mock away certain database queries. To support this, set up a `StateHandler` for each state using hooks on the `StateHandlers` property. Here is an example:

```go
pact.VerifyProvider(t, types.VerifyRequest{
	...
  StateHandlers: types.StateHandlers{
		"User jmarie exists": func() error {
			userRepository = jmarieExists
			return nil
		},
		"User jmarie is unauthenticated": func() error {
			userRepository = jmarieUnauthorized
			token = "invalid"

			return nil
		},
		"User jmarie does not exist": func() error {
			userRepository = jmarieDoesNotExist
			return nil
		},
		...
	},
})
```

As you can see, for each state (`"User jmarie exists"` etc.) we configure the local datastore differently. If this option is not configured, the `Verifier` will ignore the provider states defined in the pact and log a warning.

Note that if the State Handler errors, the test will exit early with a failure.

Read more about [Provider States](https://docs.pact.io/getting_started/provider_states).

#### Before and After Hooks

Sometimes, it's useful to be able to do things before or after a test has run, such as reset a database, log a metric etc. A `BeforeEach` runs before any other part of the Pact test lifecycle, and a `AfterEach` runs as the last step before returning the verification result back to the test.

You can add them to your Verification as follows:

```go
	pact.VerifyProvider(t, types.VerifyRequest{
		...
		BeforeEach: func() error {
			fmt.Println("before hook, do something")
			return nil
		},
		AfterEach: func() error {
			fmt.Println("after hook, do something")
			return nil
		},
	})
```

If the Hook errors, the test will fail.

#### Request Filtering

Sometimes you may need to add things to the requests that can't be persisted in a pact file. Examples of these are authentication tokens with a small life span. e.g. an OAuth bearer token: `Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42`.

For these cases, we have two facilities that should be carefully used during verification:

1. the ability to specify custom headers to be sent during provider verification. The flag to achieve this is `CustomProviderHeaders`.
1. the ability to modify a request/response and change the payload. The parameter to achieve this is `RequestFilter`.

Read on for more.

##### Example: API with Authorization

**Custom Headers**:

This header will always be sent for each and every request, and can't be dynamic. For example:

```go
  pact.VerifyProvider(t, types.VerifyRequest{
    ...
    CustomProviderHeaders:  []string{"Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42"},
  })
```

As you can see, this is your opportunity to modify\add to headers being sent to the Provider API, for example to create a valid time-bound token.

**Request Filters**

_WARNING_: This should only be attempted once you know what you're doing!

Request filters are custom middleware, that are executed for each request, allowing `token` to change between invocations. Request filters can change the request coming in, _and_ the response back to the verifier. It is common to pair this with `StateHandlers` as per above, that can set/expire the token
for different test cases:

```go
  pact.VerifyProvider(t, types.VerifyRequest{
    ...
    RequestFilter: func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
			next.ServeHTTP(w, r)
		})
	}
  })
```

_Important Note_: You should only use this feature for things that can not be persisted in the pact file. By modifying the request, you are potentially modifying the contract from the consumer tests!

#### Pending Pacts

_NOTE_: This feature is currently only available on [Pactflow]

Pending pacts is a feature that allows consumers to publish new contracts or changes to existing contracts without breaking Provider's builds. It does so by flagging the contract as "unverified" in the Pact Broker the first time a contract is published. A Provider can then enable a behaviour (via `EnablePending: true`) that will still perform a verification (and thus share the results back to the broker) but _not_ fail the verification step itself.

This enables safe introduction of new contracts into the system, without breaking Provider builds, whilst still providing feedback to Consumers as per before.

See the [docs](https://docs.pact.io/pending) and this [article](https://docs.pact.io/blog/2020/02/24/how-weve-fixed-the-biggest-problem-with-the-pact-workflow) for more background.

#### WIP Pacts

_NOTE_: This feature is currently only available on [Pactflow]

WIP Pacts builds upon pending pacts, enabling provider tests to pull in _any_ contracts applicable to the provider regardless of the `tag` it was given. This is useful, because often times consumers won't follow the exact same tagging convention and so their workflow would be interrupted. This feature enables any pacts determined to be "work in progress" to be verified by the Provider, without causing a build failure. You can enable this behaviour by specifying a valid `time.Time` field for `IncludeWIPPactsSince`. This sets the start window for which new WIP pacts will be pulled down for verification, regardless of the tag.

See the [docs](https://docs.pact.io/wip) and this [article](https://docs.pact.io/blog/2020/02/24/introducing-wip-pacts) for more background.

#### Lifecycle of a provider verification

For each _interaction_ in a pact file, the order of execution is as follows:

`BeforeEach` -> `StateHandler` -> `RequestFilter (pre)` -> `Execute Provider Test` -> `RequestFilter (post)` -> `AfterEach`

If any of the middleware or hooks fail, the tests will also fail.

### Publishing pacts to a Pact Broker and Tagging Pacts

Using a [Pact Broker] is recommended for any serious workloads, you can run your own one or use a [hosted broker].

By integrating with a Broker, you get much more advanced collaboration features and can take advantage of automation tools, such as the [can-i-deploy tool], which can tell you at any point in time, which component is safe to release.

See the [Pact Broker](https://docs.pact.io/getting_started/sharing_pacts)
documentation for more details on the Broker.

#### Publishing from Go code

```go
p := Publisher{}
err := p.Publish(types.PublishRequest{
	PactURLs:	[]string{"./pacts/my_consumer-my_provider.json"},
	PactBroker:	"http://pactbroker:8000",
	ConsumerVersion: "1.0.0",
	Tags:		[]string{"master", "dev"},
})
```

#### Publishing Provider Verification Results to a Pact Broker

If you're using a Pact Broker (e.g. a hosted one at pactflow.io), you can
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
publishing or retrieving Pact files with a Pact Broker:

- `BrokerUsername` - the username for Pact Broker basic authentication.
- `BrokerPassword` - the password for Pact Broker basic authentication.

#### Using the Pact Broker with Bearer Token authentication

The following flags are required to use bearer token authentication when
publishing or retrieving Pact files with a Pact Broker:

- `BrokerToken` - the token to authenticate with (excluding the `"Bearer"` prefix)

## Asynchronous API Testing

Modern distributed architectures are increasingly integrated in a decoupled, asynchronous fashion. Message queues such as ActiveMQ, RabbitMQ, SQS, Kafka and Kinesis are common, often integrated via small and frequent numbers of microservices (e.g. lambda).

Furthermore, the web has things like WebSockets which involve bidirectional messaging.

Pact now has experimental support for these use cases, by abstracting away the protocol and focussing on the messages passing between them.

For further reading and introduction into this topic, see this [article](https://dius.com.au/2017/09/22/contract-testing-serverless-and-asynchronous-applications/)
and our [example](https://github.com/pact-foundation/pact-go/tree/master/examples/messages) for a more detailed overview of these concepts.

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
	Consumer: "PactGoMessageConsumer",
	Provider: "PactGoMessageProvider",
}

// 4 Write the consumer test, and call VerifyMessageConsumer
// passing through the function
func TestMessageConsumer_Success(t *testing.T) {
	message := pact.AddMessage()
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
    }).
    AsType(&User{}) // Optional

	pact.VerifyMessageConsumer(t, message, userHandlerWrapper)
}
```

**Explanation**:

1.  The API - a contrived API handler example. Expects a User object and throws an `Error` if it can't handle it.
    - In most applications, some form of transactionality exists and communication with a MQ/broker happens.
    - It's important we separate out the protocol bits from the message handling bits, so that we can test that in isolation.
1.  Creates the MessageConsumer class
1.  Setup the expectations for the consumer - here we expect a `User` object with three fields
1.  Pact will send the message to your message handler. If the handler does not error, the message is saved, otherwise the test fails. There are a few key things to consider:
    - The actual request body that Pact will invoke on your handler will be contained within a `dsl.Message` object along with other context, so the body must be retrieved via `Content` attribute. If you set `Message.AsType(T)` this object will be mapped for you. If you don't want Pact to perform the conversion, you may do so on the object (`dsl.Message.Content`) or on the raw JSON (`dsl.Message.ContentRaw`).
    - All handlers to be tested must be of the shape `func(dsl.Message) error` - that is, they must accept a `Message` and return an `error`. This is how we get around all of the various protocols, and will often require a lightweight adapter function to convert it.
    - In this case, we wrap the actual `userHandler` with `userHandlerWrapper` provided by Pact.

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
    - Similar to the Consumer tests, we map the various interactions that are going to be verified as denoted by their `description` field. In this case, `a request for a dog`, maps to the `createDog` handler. Notice how this matches the original Consumer test.
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

| method          | description                                                                                     |
| --------------- | ----------------------------------------------------------------------------------------------- |
| `Identifier()`  | Match an ID (e.g. 42)                                                                           |
| `Integer()`     | Match all numbers that are integers (both ints and longs)                                       |
| `Decimal()`     | Match all real numbers (floating point and decimal)                                             |
| `HexValue()`    | Match all hexadecimal encoded strings                                                           |
| `Date()`        | Match string containing basic ISO8601 dates (e.g. 2016-01-01)                                   |
| `Timestamp()`   | Match a string containing an RFC3339 formatted timestapm (e.g. Mon, 31 Oct 2016 15:21:41 -0400) |
| `Time()`        | Match string containing times in ISO date format (e.g. T22:44:30.652Z)                          |
| `IPv4Address()` | Match string containing IP4 formatted address                                                   |
| `IPv6Address()` | Match string containing IP6 formatted address                                                   |
| `UUID()`        | Match strings containing UUIDs                                                                  |

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

## Tutorial (60 minutes)

Learn everything in Pact Go in 60 minutes: https://github.com/pact-foundation/pact-workshop-go

## Examples

### HTTP APIs

- [API Consumer](https://github.com/pact-foundation/pact-go/tree/master/examples/)
- [Golang ServeMux](https://github.com/pact-foundation/pact-go/tree/master/examples/mux)
- [Gin](https://github.com/pact-foundation/pact-go/tree/master/examples/gin)

### Asynchronous APIs

- [Message Queue](https://github.com/pact-foundation/pact-go/tree/master/examples/messages)

### Integrated examples

There are number of examples we use as end-to-end integration test prior to releasing a new binary, including publishing to a Pact Broker. To enable them, set the following environment variables

```sh
make pact
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

#### Output Logging

Pact Go uses a simple log utility ([logutils](https://github.com/hashicorp/logutils))
to filter log messages. The CLI already contains flags to manage this,
should you want to control log level in your tests, you can set it like so:

```go
pact := Pact{
  ...
	LogLevel: "DEBUG", // One of TRACE, DEBUG, INFO, ERROR, NONE
}
```

`TRACE` level logging will print the entire request/response cycle.

#### Check if the CLI tools are up to date

Pact ships with a CLI that you can also use to check if the tools are up to date. Simply run `pact-go install`, exit status `0` is good, `1` or higher is bad.

#### Disable CLI checks to speed up tests

Pact relies on a number of CLI tools for successful operation, and it performs some pre-emptive checks
during test runs to ensure that everything will run smoothly. This check, unfortunately, can add up
if spread across a large test suite. You can disable the check by setting the environment variable `PACT_DISABLE_TOOL_VALIDITY_CHECK=1` or specifying it when creating a `dsl.Pact` struct:

```go
dsl.Pact{
  ...
  DisableToolValidityCheck: true,
}
```

You can then [check if the CLI tools are up to date](#check-if-the-cli-tools-are-up-to-date) as part of your CI process once up-front and speed up the rest of the process!

#### Re-run a specific provider verification test

Sometimes you want to target a specific test for debugging an issue or some other reason.

This is easy for the consumer side, as each consumer test can be controlled
within a valid `*testing.T` function, however this is not possible for Provider verification.

But there is a way! Given an interaction that looks as follows (taken from the message examples):

```go
	message := pact.AddMessage()
	message.
		Given("user with id 127 exists").
		ExpectsToReceive("a user").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"id":   like(127),
			"name": "Baz",
			"access": eachLike(map[string]interface{}{
				"role": term("admin", "admin|controller|user"),
			}, 3),
		}).
		AsType(&types.User{})
```

and the function used to run provider verification is `go test -run TestMessageProvider`, you can test the verification of this specific interaction by setting two environment variables `PACT_DESCRIPTION` and `PACT_PROVIDER_STATE` and re-running the command. For example:

```
cd examples/message/provider
PACT_DESCRIPTION="a user" PACT_PROVIDER_STATE="user with id 127 exists" go test -v .
```

### Verifying APIs with a self-signed certificate

Supply your own TLS configuration to customise the behaviour of the runtime:

```go
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: "https://localhost:8080",
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/consumer-selfsignedtls.json", pactDir))},
		CustomTLSConfig: &tls.Config{
			RootCAs: getCaCertPool(), // Specify a custom CA pool
			// InsecureSkipVerify: true, // Disable SSL verification altogether
		},
	})
```

See [self-signed certificate](https://github.com/pact-foundation/pact-go/examles/customTls/self_signed_certificate_test.go) for an example.

### Testing AWS API Gateway APIs

AWS changed their certificate authority last year, and not all OSs have the latest CA chains. If you can't update to the latest certificate bunidles, see "Verifying APIs with a self-signed certificate" for how to work around this.

## Contact

Join us in slack: [![slack](https://slack.pact.io/badge.svg)](https://slack.pact.io)

or

- Twitter: [@pact_up](https://twitter.com/pact_up)
- Stack Overflow: [stackoverflow.com/questions/tagged/pact](https://stackoverflow.com/questions/tagged/pact)

## Documentation

Additional documentation can be found at the main [Pact website](https://pact.io).

## Roadmap

The current focus is on moving to the Rust shared core, the project can be followed [here](https://github.com/pact-foundation/pact-go/projects/2).

The general [roadmap](https://docs.pact.io/roadmap/) for Pact is available on the Pact website.

## Contributing

See [CONTRIBUTING](https://github.com/pact-foundation/pact-go/edit/master/CONTRIBUTING.md).

[spec]: https://github.com/pact-foundation/pact-specification
[stable]: https://github.com/pact-foundation/pact-go/tree/release/0.x.x
[alpha]: https://github.com/pact-foundation/pact-go/tree/release/1.1.x
[troubleshooting]: https://github.com/pact-foundation/pact-go/wiki/Troubleshooting
[getting started with pact]: http://dius.com.au/2016/02/03/microservices-pact/
[pact website]: https://docs.pact.io/
[slack channel]: https://pact-foundation.slack.com
[@pact_up]: https://twitter.com/pact_up
[pact specification v2]: https://github.com/pact-foundation/pact-specification/tree/version-2
[pact specification v3]: https://github.com/pact-foundation/pact-specification/tree/version-3
[cli tools]: https://github.com/pact-foundation/pact-ruby-standalone/releases
[installation]: #installation
[message support]: https://github.com/pact-foundation/pact-specification/tree/version-3#introduces-messages-for-services-that-communicate-via-event-streams-and-message-queues
[changelog]: https://github.com/pact-foundation/pact-go/blob/master/CHANGELOG.md
[pact broker]: https://github.com/pact-foundation/pact_broker
[hosted broker]: https://pactflow.io
[can-i-deploy tool]: https://docs.pact.io/can_i_deploy
[pactflow]: https://pactflow.io
[v2.x.x]: https://github.com/pact-foundation/pact-go/tree/2.x.x
[slack]: https://slack.pact.io
