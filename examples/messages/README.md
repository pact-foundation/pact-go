# Message Pact Example

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
	return dsl.Pact{
		Consumer:                 "PactGoMessageConsumer",
		Provider:                 "PactGoMessageProvider",
	}
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

## More information

See the discussion for the design behind the message implementation at https://gist.github.com/bethesque/c858e5c15649ae525ef0cc5264b8477c.
