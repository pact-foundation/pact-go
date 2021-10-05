# Provider Verification

## Contract Testing Process (HTTP)

Pact is a consumer-driven contract testing tool, which is a fancy way of saying that the API `Consumer` writes a test to set out its assumptions and needs of its API `Provider`(s). By unit testing our API client with Pact, it will produce a `contract` that we can share to our `Provider` to confirm these assumptions and prevent breaking changes.

The process looks like this:

![diagram](./diagrams/summary.png)

1. The consumer writes a unit test of its behaviour using a Mock provided by Pact
1. Pact writes the interactions into a contract file (as a JSON document)
1. The consumer publishes the contract to a broker (or shares the file in some other way)
1. Pact retrieves the contracts and replays the requests against a locally running provider
1. The provider should stub out its dependencies during a Pact test, to ensure tests are fast and more deterministic.

In this document, we will cover steps 4-5.

## Provider package

The provider interface is in the package: `github.com/pact-foundation/pact-go/v2/provider`.

### Verifying a Provider

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

### Managing Test Data (Provider States)

Each interaction in a pact should be verified in isolation, with no context maintained from the previous interactions. Tests that depend on the outcome of previous tests are brittle and hard to manage.
Provider states is the feature that allows you to test a request that requires data to exist on the provider.

Read more about [provider states](https://docs.pact.io/getting_started/provider_states/).

If you have defined any states (as denoted by a `Given()`) in your consumer tests, the `Verifier` can put the provider into the correct state prior to sending the actual request for validation. For example, the provider can use the state to mock away certain API endpoints or seed data into a database. To support this, you registar a `StateHandler` func for each state using hooks on the `StateHandlers` property. Here is an example:

```go
pact.VerifyProvider(t, types.VerifyRequest{
	...
	StateHandlers: v3.StateHandlers{
		"User 1234 exists": func(setup bool, s provider.ProviderStateV3) (provider.ProviderStateV3Response, error) {
			// set the database to have users
			userRepository = fullUsersRepository

			// if you have dynamic IDs and you are using provider state value generators
			// you can return a key/value response that will be used by the verifier to substitute
			// the pact file values, with the replacements here
			return provider.ProviderStateV3Response{"uuid": "1234"}, nil
		},
		"No users exist": func(setup bool, s provider.ProviderStateV3) (provider.ProviderStateV3Response, error) {
			// set the database to an empty database
			userRepository = emptyRepository

			return nil, nil
		},
	},
})
```

As you can see, for each state (`"User 1234 exists"` etc.) we configure the local datastore differently. If this option is not configured, the `Verifier` will ignore the provider states defined in the pact and log a warning.

Each handler takes a `setup` property indicating if the state is being setup (before the test) or torn dowmn (post request). This is useful if you want to cleanup after the test.

You may also optionally return a key/value map for provider state value generators to substitute values in the incoming test request.

Note that if the State Handler errors, the test will exit early with a failure.

Read more about [Provider States](https://docs.pact.io/getting_started/provider_states).

### Before and After Hooks

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

### Request Filtering

Sometimes you may need to add things to the requests that can't be persisted in a pact file. Examples of these are authentication tokens with a small life span. e.g. an OAuth bearer token: `Authorization: Bearer 0b79bab50daca910b000d4f1a2b675d604257e42`.

For these cases, we you may want the ability to modify a request/response and change the payload. The parameter to achieve this is `RequestFilter`.

#### Example: API with Authorization

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

### Connecting to a Pact Broker

In most cases, you will want to use a [Pact Broker](https://docs.pact.io/pact_broker) to manage your contracts.

The first part of the configuration is to tell Pact how to connect to your broker:

```golang
		Provider:        "V3Provider",
		ProviderVersion: os.Getenv("APP_SHA"),
		BrokerURL:       os.Getenv("PACT_BROKER_BASE_URL"),
```

The provider name uniquely identifies the application and automatically discovers consumer contracts for this provider. The [version information](https://docs.pact.io/getting_started/versioning_in_the_pact_broker) is extremely important, and is used to send compatibility information back to the broker. This should be unique per build, and is recommended to be the Git SHA.

#### Selecting pacts to verify

Once connected to the broker, you need to configure which pacts you care about verifying. Consumer version selectors are how we do this. For example, in the following setup we collect all contracts where the tag is `master` or `prod`:

```golang
		ConsumerVersionSelectors: []Selector{
			&ConsumerVersionSelector{
				Tag: "master",
			},
			&ConsumerVersionSelector{
				Tag: "prod",
			},
		},
```

Read more on [selectors](https://docs.pact.io/pact_broker/advanced_topics/consumer_version_selectors/)

#### Publishing test results

Lastly, you will want to send the verification results so that consumers can query if they are safe
to release. In your broker, it may look like this:

![screenshot of verification result](https://cloud.githubusercontent.com/assets/53900/25884085/2066d98e-3593-11e7-82af-3b41a20af8e5.png)

You need to specify the following in your verification options:

```go
PublishVerificationResults: true, // recommended only in CI
```

#### Pending Pacts

Pending pacts is a feature that allows consumers to publish new contracts or changes to existing contracts without breaking Provider's builds. It does so by flagging the contract as "unverified" in the Pact Broker the first time a contract is published. A Provider can then enable a behaviour (via `EnablePending: true`) that will still perform a verification (and thus share the results back to the broker) but _not_ fail the verification step itself.

This enables safe introduction of new contracts into the system, without breaking Provider builds, whilst still providing feedback to Consumers as per before.

See the [docs](https://docs.pact.io/pending) and this [article](http://blog.pact.io/2020/02/24/how-we-have-fixed-the-biggest-problem-with-the-pact-workflow/) for more background.

#### WIP Pacts

WIP Pacts builds upon pending pacts, enabling provider tests to pull in _any_ contracts applicable to the provider regardless of the `tag` it was given. This is useful, because often times consumers won't follow the exact same tagging convention and so their workflow would be interrupted. This feature enables any pacts determined to be "work in progress" to be verified by the Provider, without causing a build failure. You can enable this behaviour by specifying a valid `time.Time` field for `IncludeWIPPactsSince`. This sets the start window for which new WIP pacts will be pulled down for verification, regardless of the tag.

See the [docs](https://docs.pact.io/wip) and this [article](http://blog.pact.io/2020/02/24/introducing-wip-pacts/) for more background.

### Lifecycle of a provider verification

For each _interaction_ in a pact file, the order of execution is as follows:

`BeforeEach` -> `StateHandler (pre)` -> `RequestFilter (pre)` -> `Execute Provider Test` -> `RequestFilter (post)` -> `StateHandler (post)` -> `AfterEach`

If any of the middleware or hooks fail, the tests will also fail.
