/*
Pact Go enables consumer driven contract testing, providing a mock service and
DSL for the consumer project, and interaction playback and verification
for the service provider project.

Consumer tests

Consumer side Pact testing is an isolated test that ensures a given component
is able to collaborate with another (remote) component. Pact will automatically
start a Mock server in the background that will act as the collaborators' test
double.

This implies that any interactions expected on the Mock server will be validated,
meaning a test will fail if all interactions were not completed, or if unexpected
interactions were found:

A typical consumer-side test would look something like this:

	func TestLogin(t *testing.T) {

		// Create Pact, connecting to local Daemon
		// Ensure the port matches the daemon port!
		pact := &Pact{
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
			WithRequest(&Request{
				Method: "GET",
				Path:   "/login",
			}).
			WillRespondWith(&Response{
				Status: 200,
			})

		// Run the test and verify the interactions.
		err := pact.Verify(test)
		if err != nil {
			t.Fatalf("Error on Verify: %v", err)
		}

		// Write pact to file
		pact.WritePact()
	}

If this test completed successfully, a Pact file should have been written to
./pacts/my_consumer-my_provider.json containing all of the interactions
expected to occur between the Consumer and Provider.

Provider tests

Provider side Pact testing, involves verifying that the contract - the Pact file
- can be satisfied by the Provider.

A typical Provider side test would like something like:

	func TestProvider_PactContract(t *testing.T) {
		go startMyAPI("http://localhost:8000")

		response := pact.VerifyProvider(&types.VerifyRequest{
			ProviderBaseURL:        "http://localhost:8000",
			PactURLs:               []string{"./pacts/my_consumer-my_provider.json"},
			ProviderStatesURL:      "http://localhost:8000/states",
			ProviderStatesSetupURL: "http://localhost:8000/setup",
		})

		if response.ExitCode != 0 {
			t.Fatalf("Got non-zero exit code '%d', expected 0", response.ExitCode)
		}
	}

Note that `PactURLs` can be a list of local pact files or remote based
urls (possibly from a Pact Broker
- http://docs.pact.io/documentation/sharings_pacts.html).

Pact reads the specified pact files (from remote or local sources) and replays
the interactions against a running Provider. If all of the interactions are met
we can say that both sides of the contract are satisfied and the test passes.

Matching - Consumer Tests

In addition to verbatim value matching, you have 3 useful matching functions
in the `dsl` package that can increase expressiveness and reduce brittle test
cases.

	Term(example, matcher)	tells Pact that the value should match using a given regular expression, using `example` in mock responses. `example` must be a string.
	Like(content)		tells Pact that the value itself is not important, as long as the element _type_ (valid JSON number, string, object etc.) itself matches.
	EachLike(content, min)	tells Pact that the value should be an array type, consisting of elements like those passed in. `min` must be >= 1. `content` may be a valid JSON value: e.g. strings, numbers and objects.

Here is a complex example that shows how all 3 terms can be used together:

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


This example will result in a response body from the mock server that looks like:
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

See the examples in the dsl package and the matcher tests
(https://github.com/pact-foundation/pact-go/blob/master/dsl/matcher_test.go)
for more matching examples.

NOTE: You will need to use valid Ruby regular expressions
(http://ruby-doc.org/core-2.1.5/Regexp.html) and double escape backslashes.

Read more about flexible matching (https://github.com/realestate-com-au/pact/wiki/Regular-expressions-and-type-matching-with-Pact.

Publishing Pacts to a Broker and Tagging Pacts

See the Pact Broker (http://docs.pact.io/documentation/sharings_pacts.html)
documentation for more details on the Broker and this article
(http://rea.tech/enter-the-pact-matrix-or-how-to-decouple-the-release-cycles-of-your-microservices/)
on how to make it work for you.

Using the Pact Broker with Basic authentication

The following flags are required to use basic authentication when
publishing or retrieving Pact files to/from a Pact Broker:

	BrokerUsername	the username for Pact Broker basic authentication.
	BrokerPassword	the password for Pact Broker basic authentication.

Output Logging

Pact Go uses a simple log utility (logutils - https://github.com/hashicorp/logutils)
to filter log messages. The CLI already contains flags to manage this,
should you want to control log level in your tests, you can set it like so:

	pact := &Pact{
	  ...
		LogLevel: "DEBUG", // One of DEBUG, INFO, ERROR, NONE
	}
*/
package main
