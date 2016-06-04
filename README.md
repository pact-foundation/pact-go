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


[![wercker status](https://app.wercker.com/status/273436f3ec1ec8e6ea348b81e93aeea1/s/master "wercker status")](https://app.wercker.com/project/bykey/273436f3ec1ec8e6ea348b81e93aeea1)
[![Coverage Status](https://coveralls.io/repos/github/mefellows/pact-go/badge.svg?branch=master)](https://coveralls.io/github/mefellows/pact-go?branch=master)
[![Code Climate](https://codeclimate.com/github/mefellows/pact-go/badges/gpa.svg)](https://codeclimate.com/github/mefellows/pact-go)
[![Issue Count](https://codeclimate.com/github/mefellows/pact-go/badges/issue_count.svg)](https://codeclimate.com/github/mefellows/pact-go)
[![GoDoc](https://godoc.org/github.com/mefellows/pact-go?status.svg)](https://godoc.org/github.com/mefellows/pact-go)

## Installation

* Download a [release](https://github.com/mefellows/pact-go/releases) for your OS.
* Unzip the package into a known location, and add to the `PATH`.
* Run `pact-go` to see what options are available

## Running

Due to some design constraints, Pact Go runs a two-step process

1. Run `pact-go daemon` in a separate process/shell. The Consumer and Provider
DSLs communicate over a local (RPC) connection, and is transparent to clients.
1. Create your Pact Consumer/Provider Tests. It defaults to run on port `6666`.

NOTE: The daemon is completely thread safe and it is safe to leave the daemon
running for long periods (e.g. on a CI server).

### Examples

```golang
import "github.com/mefellows/pact-go/dsl"
import ...
// 1. Start the daemon with `./pact-go daemon`
// 2. cd <pact-go>/examples
// 3. go run consumer.go
func main() {

	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Port:     6666, // Ensure this port matches the daemon port!
		Consumer: "My Consumer",
		Provider: "My Provider",
	}
	defer pact.Teardown()

	// Pass in test case
	var test = func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/", pact.Server.Port))
		return err
	}

	// Set up our interactions. Note we have multiple in this test case!
	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(&dsl.Request{
			Method: "GET",
			Path:   "/",
		}).
		WillRespondWith(&dsl.Response{
			Status: 200,
		})

	// Verify
	err := pact.Verify(test)
	if err != nil {
		log.Fatalf("Error on Verify: %v", err)
	}

  // You should now have a pact file in the file `<pact-go>/pacts/my_consumer-my_provider.json`
}
```

## Contact

* Twitter: [@pact_up](https://twitter.com/pact_up)
* Google users group: https://groups.google.com/forum/#!forum/pact-support

## Documentation

Additional documentation can be found at the main [Pact website](http://pact.io) and in the [Pact Wiki](https://github.com/realestate-com-au/pact/wiki).

## Developing

For full integration testing locally, Ruby 2.1.5 must be installed. Under the hood, Pact Go bundles the [Pact Mock Service]() and [Pact Provider Verifier]() projects to implement up to v2.0 of the Pact Specification. This is only temporary, until [Pact Reference](https://github.com/pact-foundation/pact-reference/) work is completed.

* Git clone https://github.com/mefellows/pact-go.git
* Run `make dev` to build the package and setup the Ruby 'binaries' locally

### Docker

The current Wercker build uses this custom [Docker image](https://github.com/mefellows/pact-go-docker-build).
