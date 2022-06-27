# Event driven-systems

## Introduction to asynchronous API Testing

Modern distributed architectures are increasingly integrated in a decoupled, asynchronous fashion. Message queues such as ActiveMQ, RabbitMQ, SQS, Kafka and Kinesis are common, often integrated via small and frequent numbers of microservices (e.g. lambda).

Furthermore, the web has things like WebSockets which involve bidirectional messaging.

Pact has support for these use cases, by abstracting away the protocol and focussing on the messages passing between them.

[Read the docs](https://docs.pact.io/getting_started/how_pact_works#non-http-testing-message-pact) for more on how Pact deals with this.

## Contract Testing Process (Async)

Pact is a consumer-driven contract testing tool, which is a fancy way of saying that the API `Consumer` writes a test to set out its assumptions and needs of its API `Provider`(s). By unit testing our API client with Pact, it will produce a `contract` that we can share to our `Provider` to confirm these assumptions and prevent breaking changes.

The process looks like this on the consumer side:

![diagram](./diagrams/message-consumer.png)

The process looks like this on the provider (producer) side:

![diagram](./diagrams/message-provider.png)

1. The consumer writes a unit test of its behaviour using a Mock provided by Pact
1. Pact writes the interactions into a contract file (as a JSON document)
1. The consumer publishes the contract to a broker (or shares the file in some other way)
1. Pact retrieves the contracts and replays the requests against a locally running provider
1. The provider should stub out its dependencies during a Pact test, to ensure tests are fast and more deterministic.

In this document, we will cover steps 1-3.

### Consumer

A Consumer is the system that will be reading a message from a queue or some intermediary - like a Kinesis stream, websocket or S3 bucket - and be able to handle it.

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

// 2 We write a small adapter that will take the incoming Message
// and call the function with the correct type
var userHandlerWrapper = func(m AsynchronousMessage) error {
	return userHandler(*m.Content.(*User))
}

// 3 Create the Pact Message Consumer
provider, err := NewMessagePactV3(MessageConfig{
  Consumer: "V3MessageConsumer",
  Provider: "V3MessageProvider", // NOTE: this must be different to the HTTP one, can't mix both interaction styles until v4
})

// 4 Write the consumer test, and call VerifyMessageConsumer
// passing through the function
func TestMessagePact(t *testing.T) {
  // ...
	err = provider.AddMessage().
		Given(ProviderStateV3{
			Name: "User with id 127 exists",
			Parameters: map[string]interface{}{
				"id": 127,
			},
		}).
		ExpectsToReceive("a user event").
		WithMetadata(map[string]string{
			"Content-Type": "application/json",
		}).
		WithJSONContent(Map{
			"datetime": Regex("2020-01-01", "[0-9\\-]+"),
			"name":     S("Billy"),
			"lastName": S("Sampson"),
			"id":       Integer(12),
		}).
		AsType(&User{}).
		ConsumedBy(userHandlerWrapper).
		Verify(t)
	assert.NoError(t, err)
}
```

**Explanation**:

1.  The API - a contrived API handler example. Expects a User object and throws an `Error` if it can't handle it.
    - In most applications, some form of transactionality exists and communication with a MQ/broker happens.
    - It's important we separate out the protocol bits from the message handling bits, so that we can test that in isolation.
1.  Creates the MessageConsumer class
1.  Setup the expectations for the consumer - here we expect a `User` object with three fields
1.  Pact will send the message to your message handler. If the handler does not error, the message is saved, otherwise the test fails. There are a few key things to consider:
    - The actual request body that Pact will invoke on your handler will be contained within a `message.AsynchronousMessage` object along with other context, so the body must be retrieved via `Content` attribute. If you set `Message.AsType(T)` this object will be mapped for you. If you don't want Pact to perform the conversion, you may do so on the `Content` field..
    - All handlers to be tested must be of the shape `func(AsynchronousMessage) error` - that is, they must accept a `AsynchronousMessage` and return an `error`. This is how we get around all of the various protocols, and will often require a lightweight adapter function to convert it.
    - In this case, we wrap the actual `userHandler` with `userHandlerWrapper` provided by Pact.

### Provider (Producer)

A Provider (Producer in messaging parlance) is the system that will be putting a message onto the queue.

As per the Consumer case, Pact takes the position of the intermediary (MQ/broker) and checks to see whether or not the Provider sends a message that matches the Consumer's expectations.

```golang
func TestV3MessageProvider(t *testing.T) {
	var user *User

	verifier := MessageVerifier{}

	// 1. Map test descriptions to message producer functions (handlers)
	functionMappings := MessageHandlers{
		"a user event": func([]ProviderStateV3) (interface{}, error) {
			if user != nil {
				return user, nil
			} else {
				return ProviderStateV3Response{
					"message": "not found",
				}, nil
			}
		},
	}

  // 2. Setup any required states for the handlers
	stateMappings := StateHandlers{
		"User with id 127 exists": func(setup bool, s ProviderStateV3) (ProviderStateV3Response, error) {
			if setup {
				user = &User{
					ID:       127,
					Name:     "Billy",
					Date:     "2020-01-01",
					LastName: "Sampson",
				}
			}

			return ProviderStateV3Response{"id": user.ID}, nil
		},
	}

	// V3. erify the Provider with local Pact Files
	verifier.Verify(t, VerifyMessageRequest{
		VerifyRequest: VerifyRequest{
			PactFiles:     []string{filepath.ToSlash(fmt.Sprintf("%s/V3MessageConsumer-V3MessageProvider.json", pactDir))},
			StateHandlers: stateMappings,
		},
		MessageHandlers: functionMappings,
	})
}
```

**Explanation**:

1. We configure the function mappings. In this case, we have a function that generates `a user event` which is responsible for generating the `User` event that will be sent to the consumer via some message queue
1. We setup any provider states for the interaction (see [provider](./provider.md) for more on this).
1. We configure Pact to stand-in for the queue and run the verification process. Pact will read all of the interactions specified by its consumer, invokisc each function that is responsible for generating that message and inspecting their responses


## Contract Testing (Synchronous)

In additional to "fire and forget", Pact supports bi-directional messaging protocols such as gRPC and websockets.

[Diagram TBC]

| Mode  | Custom Transport | Method |
| ----- | ---------------- | ------ |
| Sync  | Yes              | b      |
| Sync  | No               | c      |
| Async | Yes              | d      |
| Async | No               | e      |