<span align="center">

![logo](https://user-images.githubusercontent.com/53900/121775784-0191d200-cbcd-11eb-83dd-adc001b94519.png)

# Pact Go

[![Test](https://github.com/pact-foundation/pact-go/actions/workflows/test.yml/badge.svg?branch=2.x.x)](https://github.com/pact-foundation/pact-go/actions/workflows/test.yml)
[![Coverage Status](https://coveralls.io/repos/github/pact-foundation/pact-go/badge.svg?branch=HEAD)](https://coveralls.io/github/pact-foundation/pact-go?branch=HEAD)
[![Go Report Card](https://goreportcard.com/badge/github.com/pact-foundation/pact-go)](https://goreportcard.com/report/github.com/pact-foundation/pact-go)
[![GoDoc](https://godoc.org/github.com/pact-foundation/pact-go?status.svg)](https://godoc.org/github.com/pact-foundation/pact-go)

#### Fast, easy and reliable testing for your APIs and microservices.

</span>

<br />
<p align="center">
  <a href="https://docs.pact.io"><img src="https://user-images.githubusercontent.com/53900/121777789-32770480-cbd7-11eb-903b-e6623b0798ff.gif" alt="Pact Go Demo"/></a>
</p>
<br />

<table>
<tr>
<td>

**Pact** is the de-facto API contract testing tool. Replace expensive and brittle end-to-end integration tests with fast, reliable and easy to debug unit tests.

- ⚡ Lightning fast
- 🎈 Effortless full-stack integration testing - from the front-end to the back-end
- 🔌 Supports HTTP/REST and event-driven systems
- 🛠️ Configurable mock server
- 😌 Powerful matching rules prevents brittle tests
- 🤝 Integrates with Pact Broker / Pactflow for powerful CI/CD workflows
- 🔡 Supports 12+ languages

**Why use Pact?**

Contract testing with Pact lets you:

- ⚡ Test locally
- 🚀 Deploy faster
- ⬇️ Reduce the lead time for change
- 💰 Reduce the cost of API integration testing
- 💥 Prevent breaking changes
- 🔎 Understand your system usage
- 📃 Document your APIs for free
- 🗄 Remove the need for complex data fixtures
- 🤷‍♂️ Reduce the reliance on complex test environments

Watch our [series](https://www.youtube.com/playlist?list=PLwy9Bnco-IpfZ72VQ7hce8GicVZs7nm0i) on the problems with end-to-end integrated tests, and how contract testing can help.

</td>
</tr>
</table>

![----------](https://raw.githubusercontent.com/pactumjs/pactum/master/assets/rainbow.png)

## Documentation

This readme offers an basic introduction to the library. The full documentation for Pact Go and the rest of the framework is available at https://docs.pact.io/.

- [Installation](#installation)
- [Consumer Testing](./docs/consumer.md)
- [Provider Testing](./docs/provider.md)
- [Event Driven Systems](./docs/messages.md)
- [Migration guide](./MIGRATION.md)
- [Troubleshooting](./docs/troubleshooting.md)

### Tutorial (60 minutes)

Learn everything in Pact Go in 60 minutes: https://github.com/pact-foundation/pact-workshop-go

## Need Help

- [Join](<(http://slack.pact.io)>) our community [slack workspace](http://pact-foundation.slack.com/).
- Stack Overflow: https://stackoverflow.com/questions/tagged/pact
- Say 👋 on Twitter: [@pact_up]

## Installation

```shell
# install pact-go as a dev dependency
go get github.com/pact-foundation/pact-go/v2@2.x.x

# NOTE: If using Go 1.19 or later, you need to run go install instead 
# go install github.com/pact-foundation/pact-go/v2@2.x.x

# download and install the required libraries. The pact-go will be installed into $GOPATH/bin, which is $HOME/go/bin by default. 
pact-go -l DEBUG install 

# 🚀 now write some tests!
```

If the `pact-go` command above is not found, make sure that `$GOPATH/bin` is in your path. I.e.,
```shell
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

You can also keep the library versions up to date by running the `version.CheckVersion()` function.

<details><summary>Manual Installation Instructions</summary>

### Manual

Downlod the latest `Pact FFI Library` [library] for your OS, and install onto a standard library search path (we suggest: `/usr/local/lib` on OSX/Linux):

Ensure you have the correct extension for your OS:

- For Mac OSX: `.dylib` (For M1 users, you need the `aarch64-apple-darwin` version)
- For Linux: `.so`
- For Windows: `.dll`

```sh
wget https://github.com/pact-foundation/pact-reference/releases/download/libpact_ffi-v0.1.2/libpact_ffi-osx-x86_64.dylib.gz
gunzip libpact_ffi-osx-x86_64.dylib.gz
mv libpact_ffi-osx-x86_64.dylib /usr/local/lib/libpact_ffi.dylib
```

Test the installation:

```sh
pact-go help
```

</details>

![----------](https://raw.githubusercontent.com/pactumjs/pactum/master/assets/rainbow.png)

## Usage

### Consumer package

The consumer interface is in the package: `github.com/pact-foundation/pact-go/v2/consumer`.

#### Writing a Consumer test

Pact is a consumer-driven contract testing tool, which is a fancy way of saying that the API `Consumer` writes a test to set out its assumptions and needs of its API `Provider`(s). By unit testing our API client with Pact, it will produce a `contract` that we can share to our `Provider` to confirm these assumptions and prevent breaking changes.

In this example, we are going to be testing our User API client, responsible for communicating with the `UserAPI` over HTTP. It currently has a single method `GetUser(id)` that will return a `*User`.

Pact tests have a few key properties. We'll demonstrate a common example using the 3A `Arrange/Act/Assert` pattern.

```golang
func TestUserAPIClient(t *testing.T) {
	// Specify the two applications in the integration we are testing
	// NOTE: this can usually be extracted out of the individual test for re-use)
	mockProvider, err := NewV2Pact(MockHTTPProviderConfig{
		Consumer: "UserAPIConsumer",
		Provider: "UserAPI",
	})
	assert.NoError(t, err)

	// Arrange: Setup our expected interactions
	mockProvider.
		AddInteraction().
		Given("A user with ID 10 exists").
		UponReceiving("A request for User 10").
		WithRequest("GET", S("/user/10")).
		WillRespondWith(200).
		WithBodyMatch(&User{})

	// Act: test our API client behaves correctly
	err = mockProvider.ExecuteTest(func(config MockServerConfig) error {
		// Initialise the API client and point it at the Pact mock server
		// Pact spins up a dedicated mock server for each test
		client := newClient(config.Host, config.Port)

		// Execute the API client
		user, err := client.GetUser("10")

		// Assert: check the result
		assert.NoError(t, err)
		assert.Equal(t, 10, user.ID)

		return err
	})
	assert.NoError(t, err)
}
```

You can see (and run) the full version of this in `./examples/basic_test.go`.

For a full example, see the Pactflow terraform provider [pact tests](https://github.com/pactflow/terraform-provider-pact/blob/master/client/client_pact_test.go).

![----------](https://raw.githubusercontent.com/pactumjs/pactum/master/assets/rainbow.png)

### Provider package

The provider interface is in the package: `github.com/pact-foundation/pact-go/v2/provider`

#### Verifying a Provider

A provider test takes one or more pact files (contracts) as input, and Pact verifies that your provider adheres to the contract. In the simplest case, you can verify a provider as per below.

```golang
func TestV3HTTPProvider(t *testing.T) {
	// 1. Start your Provider API in the background
	go startServer()

	verifier := HTTPVerifier{}

	// Verify the Provider with local Pact Files
	// The console will display if the verification was successful or not, the
	// assertions being made and any discrepancies with the contract
	err := verifier.VerifyProvider(t, VerifyRequest{
		ProviderBaseURL: "http://localhost:1234",
		PactFiles: []string{
			filepath.ToSlash("/path/to/SomeConsumer-SomeProvider.json"),
		},
	})

	// Ensure the verification succeeded
	assert.NoError(t, err)
}
```

![----------](https://raw.githubusercontent.com/pactumjs/pactum/master/assets/rainbow.png)

## Compatibility

<details><summary>Specification Compatibility</summary>

| Version | Stable | [Spec] Compatibility | Install            |
| ------- | ------ | -------------------- | ------------------ |
| 2.0.x   | Beta   | 2, 3                 | See [installation] |
| 1.0.x   | Yes    | 2, 3\*               | 1.x.x [1xx]        |
| 0.x.x   | Yes    | Up to v2             | 0.x.x [stable]     |

_\*_ v3 support is limited to the subset of functionality required to enable language inter-operable [Message support].

</details>

## Roadmap

The [roadmap](https://docs.pact.io/roadmap/) for Pact and Pact Go is outlined on our main website.
Detail on the native Go implementation can be found [here](https://github.com/pact-foundation/pact-go/wiki/Native-implementation-roadmap).

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md).

<a href="https://github.com/pact-foundation/pact-go/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=pact-foundation/pact-go" />
</a>
<br />

[spec]: https://github.com/pact-foundation/pact-specification
[1xx]: https://github.com/pact-foundation/pact-go/
[stable]: https://github.com/pact-foundation/pact-go/tree/release/0.x.x
[alpha]: https://github.com/pact-foundation/pact-go/tree/release/1.1.x
[troubleshooting]: https://github.com/pact-foundation/pact-go/wiki/Troubleshooting
[pact wiki]: https://github.com/pact-foundation/pact-ruby/wiki
[getting started with pact]: http://dius.com.au/2016/02/03/microservices-pact/
[pact website]: http://docs.pact.io/
[slack channel]: https://gophers.slack.com/messages/pact/
[@pact_up]: https://twitter.com/pact_up
[pact specification v2]: https://github.com/pact-foundation/pact-specification/tree/version-2
[pact specification v3]: https://github.com/pact-foundation/pact-specification/tree/version-3
[library]: https://github.com/pact-foundation/pact-reference/releases
[cli tools]: https://github.com/pact-foundation/pact-reference/releases
[installation]: #installation
[message support]: https://github.com/pact-foundation/pact-specification/tree/version-3#introduces-messages-for-services-that-communicate-via-event-streams-and-message-queues
[changelog]: https://github.com/pact-foundation/pact-go/blob/master/CHANGELOG.md
[pact broker]: https://github.com/pact-foundation/pact_broker
[hosted broker]: pact.dius.com.au
[can-i-deploy tool]: https://github.com/pact-foundation/pact_broker/wiki/Provider-verification-results
[pactflow]: https://pactflow.io
