# Migration Guide

## Major changes from v1.x.x to v2.x.x

1. Ruby shared core has been replaced by the [Rust shared core](https://github.com/pact-foundation/pact-reference/tree/master/rust/)
1. In general, the interface has been reduced in favour of encouraging use of the Pact [CLI tools]
1. `Publisher` interface has been removed
1. Ability to create and verify both [v2] and [v3] [specification] pacts
1. A bunch of new features, including the new v3 specification [matchers and generators](https://github.com/pact-foundation/pact-specification/tree/version-3/), injected provider states and more.
1. Each test is given a dedicated server and port, simplifying lifecycle events and improving isolation between tests.
1. You _must_ clear out any pact directory prior to the tests running. Any pact files will be appended to during test execution or may result in conflicts.

### Package version

The new package is migratable, with the the `/v2` major version bump

i.e. `github.com/pact-foundation/pact-go/v2`

### Consumer

#### Primary Interface

- `dsl.Pact` was the primary interface. This is now replaced with `NewV2Pact` and the `NewV3Pact` methods, which will return you a builder for the corresponding specification.
- `Verify` is now `ExecuteTest` to avoid ambiguity with provider side verification. It also accepts a `*testing.T` argument, to improve error reporting and resolution.

These are available in consumer package: `"github.com/pact-foundation/pact-go/v2/consumer"`

The interface now also exposes an improved builder interface that aids discoverability, better types and readability. The previous "all-in-one" request/response style has been preserved (`WithCompleteRequest` and `WithCompleteResponse`), to aid in migration. See the `TestConsumerV2AllInOne` test in the examples to see this in action.

#### Consumer Test `func`

The test function provided to `Verify` now accepts a `func(MockServerConfig) error` (previously a `func() error`).

The test function will receive the details of the mock server configuration, such as the host and port, which may be useful in configuring the test HTTP client, for example. This prevents the need to maintain global state between tests as was previously the case.

#### Regexes

You can now use proper POSIX compliant regexes :)

#### Speed / Parallelism

the `NewV2Pact` and `NewV3Pact` interfaces are not thread safe (i.e. you shouldn't run parallel tests with them), but you don't need to. Here is one of the examples run on a loop 1000 times, writing to a different file on each run:

```
➜  examples ✗ time go test -count=1 -tags=consumer -run TestConsumerV2 .
ok  	github.com/pact-foundation/pact-go/examples	4.269s
go test -count=1 -tags=consumer -run TestConsumerV2 .  4.34s user 1.93s system 113% cpu 5.542 total
➜  examples ✗ ls -la pacts/*.json | wc -l
    1001
➜  examples ✗
```

There is no real need, when in < 5 seconds you can run 1000 consumer pact tests!

#### Body `content-type`

The `content-type` is now a mandatory field, to improve matching for bodies of various types.

#### Binary Payload and Multipart Requests

Two new builder methods exist for binary/file payloads:

- `WithBinaryBody` accepts a `[]byte` for matching on binary payloads (e.g. images)
- `WithMultipartFile` accepts a path to file from the file system, and the multipart boundary

#### Query Strings

Query strings are now accepted (and [encoded](https://github.com/pact-foundation/pact-specification/tree/version-3/#query-strings-are-stored-as-map-instead-of-strings)) as nested data structures, instead of flat strings and may have an array of values.

### Provider

#### Provider Interface

- `dsl.Pact` was the primary interface. This is now replaced with `HTTPVerifier` struct and `VerifyProvider` method.

These are available in provider package: `"github.com/pact-foundation/pact-go/v2/provider"`

#### Provider State Handlers

Consumers may now specify multiple provider states. When the provider verifies them, it will invoke the `StateHandler` for each of the states in the interaction.

The `StateHandler` type has also changed in 3 important ways:

1. Provider states may contain [parameters], which may be useful for the state setup (e.g. data for creation of a state)
1. There is now a `setup` bool, indicating if the state is being setup or torn down. This is helpful for cleaning up state specific items (separate to `AfterEach` and `BeforeEach` which do not have access to the current state information)
1. If the consumer code uses provider state injected values (citation needed), the provider may return a JSON payload that will be substituted into the incoming request. See this [example](https://pactflow.io/blog/injecting-values-from-provider-states/) for more

### Message pacts

#### Consumer

- `dsl.Pact` was the primary interface. This is now replaced with `NewMessagePactV3` builder. The `VerifyMessageConsumer` method is now replaced by the method `Verify` on the builder.

#### Provider

- `dsl.Pact` was the primary interface. This is now replaced by the `MessageVerifier` struct and the `Verify` method. The main difference is the state handlers, as discussed above.

These are available in message package: `"github.com/pact-foundation/pact-go/v2/message"`

[CLI Tools](https://docs.pact.io/implementation_guides/cli/)
[v2](https://github.com/pact-foundation/pact-specification/tree/version-3/)
[v3](https://github.com/pact-foundation/pact-specification/tree/version-2/)
[specification](https://github.com/pact-foundation/pact-specification/)
[parameters](https://github.com/pact-foundation/pact-specification/tree/version-3/#allow-multiple-provider-states-with-parameters)
