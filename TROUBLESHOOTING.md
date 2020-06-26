# Troubleshooting

Here we can find common issues and how to solve them.

## Bodies match and only differ by whitespace/formatting

An example error might look like this:

```	  
1) Verifying a pact between service-one and service-two Given User traffic exists for the current month A request for all traffic for the current month with GET /v1/traffic/usage/json returns a response which has a matching body
		     Failure/Error: expect(response_body).to match_term expected_response_body, diff_options
		       Actual: [{"org":"a6df0d8e-916b-4998-ac6e-149b299e2c9f","usage":12345}]
		       
		       @@ -1,7 +1,2 @@
		       -[
		       -  {
		       -    "org": "a6df0d8e-916b-4998-ac6e-149b299e2c9f",
		       -    "usage": 12345
		       -  },
		       -]
		       +[{"org":"a6df0d8e-916b-4998-ac6e-149b299e2c9f","usage":12345}]
		       
		       Key: - means "expected, but was not found". 
		            + means "actual, should not be found". 
		            Values where the expected matches the actual are not shown.
		     # /usr/local/bin/pact-provider-verifier/lib/app/pact-provider-verifier.rb:3:in `<main>'
		
		Finished in 0.0183 seconds
		2 examples, 1 failure
```

The problem is that there is no `Content-Type` being sent through to inform the underlying validation library how to compare requests/response. In this case, it defaults to `text/plain` which is sensitive to differences.

You likely want `application/json` or similar, to compare JSON payloads e.g.

```
		WillRespondWith(dsl.Response{
			Status: 200,
			Headers: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
			Body: `[{
				    "org": "a6df0d8e-916b-4998-ac6e-149b299e2c9f",
                                    "usage": 12345
			       }]`,
		})
```

See https://github.com/pact-foundation/pact-go/issues/15 for background.
