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
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)

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
	4. [Using the Pact Broker with Basic authentication](#using-the-pact-broker-with-basic-authentication)
	5. [Output Logging](#output-logging)
4. [Examples](#examples)
5. [Contact](#contact)
6. [Documentation](#documentation)
7. [Roadmap](#roadmap)
8. [Contributing](#contributing)

<!-- /TOC -->

## Installation

* Download a [release](https://github.com/pact-foundation/pact-go/releases) for your OS.
* Unzip the package into a known location, and add to the `PATH`.
* Run `pact-go` to see what options are available.

## Running

Due to some design constraints, Pact Go runs a two-step process:

1. Run `pact-go daemon` in a separate process/shell. The Consumer and Provider
DSLs communicate over a local (RPC) connection, and is transparent to clients.
1. Create your Pact Consumer/Provider Tests. It defaults to run on port `6666`.

*NOTE: The daemon is completely thread safe and it is normal to leave the daemon
running for long periods (e.g. on a CI server).*

### Consumer
1. Start the daemon with `./pact-go daemon`.
1. `cd <pact-go>/examples`.
1. `go run consumer.go`.

```go
import "github.com/pact-foundation/pact-go/dsl"
import ...

func TestLogin(t *testing.T) {

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
		_, err := http.Get("http://localhost:8000/")
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


### Provider

1. Start your Provider API:

	```go
	mux := http.NewServeMux()
	mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
	})
	mux.HandleFunc("/states", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `{"My Consumer": ["Some state", "Some state2"]}`)
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
	go http.ListenAndServe(":8000"), mux)
	```

	Note that the server has 2 endpoints: `/states` and `/setup` that allows the
	verifier to setup
	[provider states](http://docs.pact.io/documentation/provider_states.html) before
	each test is run.

2. Verify provider API

	You can now tell Pact to read in your Pact files and verify that your API will
	satisfy the requirements of each of your known consumers:

	```go
	response := pact.VerifyProvider(&types.VerifyRequest{
		ProviderBaseURL:        "http://localhost:8000",
		PactURLs:               []string{"./pacts/my_consumer-my_provider.json"},
		ProviderStatesURL:      "http://localhost:8000/states",
		ProviderStatesSetupURL: "http://localhost:8000/setup",
	})

	if response.ExitCode != 0 {
		t.Fatalf("Got non-zero exit code '%d', expected 0", response.ExitCode)
	}
	```

	Note that `PactURLs` is a list of local pact files or remote based
	urls (e.g. from a
	[Pact Broker](http://docs.pact.io/documentation/sharings_pacts.html)).


	See the `Skip()'ed` [integration tests](https://github.com/pact-foundation/pact-go/blob/master/dsl/pact_test.go)
	for a more complete E2E example.

#### Provider States

Each interaction in a pact should be verified in isolation, with no context
maintained from the previous interactions. So how do you test a request that
requires data to exist on the provider? Provider states are how you achieve
this using Pact.

Provider states also allow the consumer to make the same request with different
expected responses (e.g. different response codes, or the same resource with a
different subset of data).

States are configured on the consumer side when you issue a dsl.Given() clause
with a corresponding request/response pair.

Configuring the provider is a little more involved, and (currently) requires 2
running API endpoints to retrieve and configure available states during the
verification process. The two options you must provide to the dsl.VerifyRequest
are:

```go
ProviderStatesURL: 			 GET URL to fetch all available states (see types.ProviderStates)
ProviderStatesSetupURL: 	POST URL to set the provider state (see types.ProviderState)
```

Example routes using the standard Go http package might look like this:

```go
// Return known provider states to the verifier (ProviderStatesURL):
mux.HandleFunc("/states", func(w http.ResponseWriter, req *http.Request) {
	states :=
	`{
		"My Front end consumer": [
			"User A exists",
			"User A does not exist"
		],
		"My api friend": [
			"User A exists",
			"User A does not exist"
		]
	}`
		fmt.Fprintf(w, states)
		w.Header().Add("Content-Type", "application/json")
})

// Handle a request from the verifier to configure a provider state (ProviderStatesSetupURL)
mux.HandleFunc("/setup", func(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Retrieve the Provider State
	var state types.ProviderState

	body, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	json.Unmarshal(body, &state)

	// Setup database for different states
	if state.State == "User A exists" {
		svc.userDatabase = aExists
	} else if state.State == "User A is unauthorized" {
		svc.userDatabase = aUnauthorized
	} else {
		svc.userDatabase = aDoesNotExist
	}
})
```

See the examples or read more at http://docs.pact.io/documentation/provider_states.html.

### Publishing Pacts to a Broker and Tagging Pacts

See the [Pact Broker](http://docs.pact.io/documentation/sharings_pacts.html)
documentation for more details on the Broker and this [article](http://rea.tech/enter-the-pact-matrix-or-how-to-decouple-the-release-cycles-of-your-microservices/)
on how to make it work for you.

#### Publishing from Go code

```go
pact.PublishPacts(&types.PublishRequest{
	PactBroker:             "http://pactbroker:8000",
	PactURLs:               []string{"./pacts/my_consumer-my_provider.json"},
	ConsumerVersion:        "1.0.0",
	Tags:                   []string{"latest", "dev"},
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

## Examples

* [Simple Consumer](https://github.com/pact-foundation/pact-go/tree/master/examples/consumer.go)
* [Go Kit](https://github.com/pact-foundation/pact-go/tree/master/examples/go-kit)

## Contact

* Twitter: [@pact_up](https://twitter.com/pact_up)
* Google users group: https://groups.google.com/forum/#!forum/pact-support
* Gitter: https://gitter.im/realestate-com-au/pact

## Documentation

Additional documentation can be found at the main [Pact website](http://pact.io) and in the [Pact Wiki](https://github.com/realestate-com-au/pact/wiki).

## Roadmap

The [roadmap](docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).
