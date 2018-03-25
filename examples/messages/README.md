# Example - Message Pact

Implements a POC of the message implementation at https://gist.github.com/bethesque/c858e5c15649ae525ef0cc5264b8477c.

## Requirements

* Ruby 2.3.4+
* Golang

## Running

```
./run.sh
```

## Output

Should look something like:

```
=== RUN   TestPact_Provider
2018/02/27 09:36:45 [DEBUG] API handler starting: port 9393 ([::]:9393)
2018/02/27 09:36:45 [DEBUG] waiting for port 9393 to become available
2018/02/27 09:36:45 [DEBUG] daemon - verifying provider
2018/02/27 09:36:45 [DEBUG] starting verification service with args: [exec pact-provider-verifier message-pact.json --provider-base-url http://localhost:9393 --format json]
Calling 'text' function that would produce a message
=== RUN   TestPact_Provider/has_matching_content
--- PASS: TestPact_Provider (1.94s)
    --- PASS: TestPact_Provider/has_matching_content (0.00s)
    	pact.go:363: Verifying a pact between Foo and Bar A test message has matching content
PASS
ok  	github.com/pact-foundation/pact-go/examples/messages/provider	1.966s
--> Shutting down running processes
```