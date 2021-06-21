# Consumer Tests

## Contract Testing Process (HTTP)

Pact is a consumer-driven contract testing tool, which is a fancy way of saying that the API `Consumer` writes a test to set out its assumptions and needs of its API `Provider`(s). By unit testing our API client with Pact, it will produce a `contract` that we can share to our `Provider` to confirm these assumptions and prevent breaking changes.

The process looks like this:

![diagram](./diagrams/summary.png)

1. The consumer writes a unit test of its behaviour using a Mock provided by Pact
1. Pact writes the interactions into a contract file (as a JSON document)
1. The consumer publishes the contract to a broker (or shares the file in some other way)
1. Pact retrieves the contracts and replays the requests against a locally running provider
1. The provider should stub out its dependencies during a Pact test, to ensure tests are fast and more deterministic.

In this document, we will cover steps 1-3.

## Consumer package

The consumer interface is in the package: `github.com/pact-foundation/pact-go/v2/consumer`.

The two primary interfaces are `NewV2Pact` and `NewV3Pact`. If your provider is also V3 compatible, you can use the V3 variant, otherwise you should stick with V2.

### Writing a Consumer test

The purpose of a Pact test is to unit test the API Client of the consumer.

In this example, we are going to be testing our Product API client, responsible for communicating with the `ProductAPI` over HTTP. It currently has a single method `GetProduct(id)` that will return a `*Product`.

Pact tests have a few key properties. We'll demonstrate a common example using the 3A `Arrange/Act/Assert` pattern.

Here is a sequence diagram that shows how a consumer test works:

![diagram](./diagrams/workshop_step3_pact.svg)

### Example

```golang
func TestProductAPIClient(t *testing.T) {
	// Specify the two applications in the integration we are testing
	// NOTE: this can usually be extracted out of the individual test for re-use)
	mockProvider, err := NewV2Pact(MockHTTPProviderConfig{
		Consumer: "ProductAPIConsumer",
		Provider: "ProductAPI",
	})
	assert.NoError(t, err)

	// Arrange: Setup our expected interactions
	mockProvider.
		AddInteraction().
		Given("A Product with ID 10 exists").
		UponReceiving("A request for Product 10").
		WithRequest("GET", S("/product/10")).
		WillRespondWith(200).
		WithBodyMatch(&Product{}) // This uses struct tags for matchers

	// Act: test our API client behaves correctly
	err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
		// Initialise the API client and point it at the Pact mock server
		// Pact spins up a dedicated mock server for each test
		client := newClient(config.Host, config.Port)

		// Execute the API client
		product, err := client.GetProduct("10")

		// Assert: check the result
		assert.NoError(t, err)
		assert.Equal(t, 10, product.ID)

		return err
	})
	assert.NoError(t, err)
}
```

### Matching

In addition to matching on exact values, there are a number of useful matching functions
in the `matching` package that can increase the expressiveness of your tests and reduce brittle
test cases.

Rather than use hard-coded values which must then be present on the Provider side,
you can use regular expressions and type matches on objects and arrays to validate the
structure of your APIs.

Matchers can be used on the `Body`, `Headers`, `Path` and `Query` fields of the request,
and the `Body` and `Headers` on the response.

_NOTE: Some matchers are only compatible with the V3 interface, and must not be used with a V2 Pact. Your test will panic if this is attempted_

| Matcher                            | Min. Compatibility | Description                                                                                                                                                                                                                                                 |
| ---------------------------------- | ------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Like(content)`                    | V2                 | Tells Pact that the value itself is not important, as long as the element _type_ (valid JSON number, string, object etc.) itself matches.                                                                                                                   |
| `Term(example, matcher)`           | V2                 | Tells Pact that the value should match using a given regular expression, using `example` in mock responses. `example` must be a string.                                                                                                                     |
| `EachLike(content, min)`           | V2                 | Tells Pact that the value should be an array type, consisting of elements like those passed in. `min` must be >= 1. `content` may be any valid JSON value: e.g. strings, numbers and objects.                                                               |
| `Equality(content)`                | V3                 | Matchers cascade, equality resets the matching process back to exact values                                                                                                                                                                                 |
| `Includes(content)`                | V3                 | Checks if the given string is contained by the actual value                                                                                                                                                                                                 |
| `FromProviderState(expr, example)` | V3                 | Marks an item as to be dynamically injected from the provider state during provider verification                                                                                                                                                            |
| `EachKeyLike(key, template)`       | V3                 | Object where the key itself is ignored, but the value template must match. Useful for dynamic keys.                                                                                                                                                         |
| `ArrayContaining(variants)`        | V3                 | Allows heterogenous items to be matched within a list. Unlike EachLike which must be an array with elements of the same shape, ArrayContaining allows objects of different types and shapes. Useful for hypermedia responses such as Siron, HAL and JSONAPI |
| `ArrayMinMaxLike(min, max`         | V3                 | Like EachLike except has a bounds on the max and the min                                                                                                                                                                                                    |
| `ArrayMaxLike`                     | V3                 | Like EachLike except has a bounds on the max                                                                                                                                                                                                                |
| `DateGenerated`                    | V3                 | Matches a cross platform formatted date, and generates a current date during verification                                                                                                                                                                   |
| `TimeGenerated`                    | V3                 | Matches a cross platform formatted date, and generates a current time during verification                                                                                                                                                                   |
| `DateTimeGenerated`                | V3                 | Matches a cross platform formatted datetime, and generates a current datetime during verification                                                                                                                                                           |

#### Match common formats

| method          | Min. Compatibility | description                                                                                     |
| --------------- | ------------------ | ----------------------------------------------------------------------------------------------- |
| `Identifier()`  | V2                 | Match an ID (e.g. 42)                                                                           |
| `Integer()`     | V3                 | Match all numbers that are integers (both ints and longs)                                       |
| `Decimal()`     | V3                 | Match all real numbers (floating point and decimal)                                             |
| `HexValue()`    | V2                 | Match all hexadecimal encoded strings                                                           |
| `Date()`        | V2                 | Match string containing basic ISO8601 dates (e.g. 2016-01-01)                                   |
| `Timestamp()`   | V2                 | Match a string containing an RFC3339 formatted timestamp (e.g. Mon, 31 Oct 2016 15:21:41 -0400) |
| `Time()`        | V2                 | Match string containing times in ISO date format (e.g. T22:44:30.652Z)                          |
| `IPv4Address()` | V2                 | Match string containing IP4 formatted address                                                   |
| `IPv6Address()` | V2                 | Match string containing IP6 formatted address                                                   |
| `UUID()`        | V2                 | Match strings containing UUIDs                                                                  |

#### Nesting Matchers

Matchers may be nested in other objects. Here is a more complex example that shows how 3 common matchers can be used together:

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

#### Auto-generate matchers from struct tags

Furthermore, if you isolate your Data Transfer Objects (DTOs) to an adapters package so that they exactly reflect the interface between you and your provider, then you can leverage `WithBodyMatch(object)` option to auto-generate the expected response body in your contract tests. Under the hood, it recursively traverses the DTO struct and uses `Term, Like, and EachLike` to create the contract.

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
    ...
		WithBodyMatch(DTO{}), // That's it!!!
```

The `pact` struct tags shown above are optional. By default, it asserts that the JSON shape matches the struct and that the field types match.

#### Matching binary payload and multipart requests

Two builder methods exist for binary/file payloads:

- `WithBinaryBody` accepts a `[]byte` for matching on binary payloads (e.g. images)
- `WithMultipartFile` accepts a path to file from the file system, and the multipart boundary

### Managing Test Data (using Provider States)

Each interaction in a pact should be verified in isolation, with no context maintained from the previous interactions. Tests that depend on the outcome of previous tests are brittle and hard to manage.
Provider states is the feature that allows you to test a request that requires data to exist on the provider.

Read more about [provider states](https://docs.pact.io/getting_started/provider_states/)

There are several ways to define a provider state:

1. Using the `Given` builder method passing in a plain string.
1. Using the `GivenWithParameters` builder method, passing in a string description and a hash of parameters to be used by the provider during verification.
1. Using the `FromProviderState` builder methid, specifying an expression to be replaced by the provider during erification, and example value to use in the consumer test. Example: `FromProviderState("${name}", "billy"),`. [Read more](https://pactflow.io/blog/injecting-values-from-provider-states/) on this feature.

For V3 tests, these methods may be called multiple times, resulting in more than 1 state for a given interaction.

## Publishing pacts to a Broker

We recommend publishing the contracts to a Pact Broker (or https://pactflow.io) using the [CLI Tools]()https://docs.pact.io/implementation_guides/cli/#pact-cli.

[Read more](https://docs.pact.io/pact_broker/publishing_and_retrieving_pacts/) about publishing pacts.
