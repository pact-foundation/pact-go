# Troubleshooting

## Output Logging

Pact Go uses a simple log utility ([logutils](https://github.com/hashicorp/logutils))
to filter log messages. The CLI already contains flags to manage this,
should you want to control log level in your tests, you can set it like so:

```
SetLogLevel("trace")
```

You can also export `LOG_LEVEL=trace` before running a test to increase verbosity.

## Library status check

Pact ships with a CLI that you can also use to check if the tools are up to date. Simply run `pact-go check` - an exit status of `0` is good, `1` or higher is bad. `pact-go install` will also do this, and also install any dependencies if missing.

You can also opt to have Pact automatically upgrade library version using the function `CheckVersion()`.

Pact go from 2.0.0-beta-11 onwards, also stores a configuration file in `~/.pact/pact-go.yml` that contains the version information for the current libraries it manages. You should not edit this file, however it has a structure as follows:

```yaml
libraries:
  libpact_ffi:
    libname: libpact_ffi
    version: 0.3.2
    hash: d6503769896eecbc027815d20aff19c3
```

#### Re-run a specific provider verification test

Sometimes you want to target a specific test for debugging an issue or some other reason.

This is easy for the consumer side, as each consumer test can be controlled
within a valid `*testing.T` function, however this is not possible for Provider verification.

But there is a way! Given an interaction that looks as follows (taken from the message examples):

```go
	message := pact.AddMessage()
	message.
		Given("user with id 127 exists").
		ExpectsToReceive("a user").
		WithMetadata(commonHeaders).
		WithContent(map[string]interface{}{
			"id":   like(127),
			"name": "Baz",
			"access": eachLike(map[string]interface{}{
				"role": term("admin", "admin|controller|user"),
			}, 3),
		}).
		AsType(&types.User{})
```

and the function used to run provider verification is `go test -run TestMessageProvider`, you can test the verification of this specific interaction by setting two environment variables `PACT_DESCRIPTION` and `PACT_PROVIDER_STATE` and re-running the command. For example:

```
cd examples/message/provider
PACT_DESCRIPTION="a user" PACT_PROVIDER_STATE="user with id 127 exists" go test -v .
```

### Verifying APIs with a self-signed certificate

Supply your own TLS configuration to customise the behaviour of the runtime:

```go
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL: "https://localhost:8080",
		PactURLs:        []string{filepath.ToSlash(fmt.Sprintf("%s/consumer-selfsignedtls.json", pactDir))},
		CustomTLSConfig: &tls.Config{
			RootCAs: getCaCertPool(), // Specify a custom CA pool
			// InsecureSkipVerify: true, // Disable SSL verification altogether
		},
	})
```
