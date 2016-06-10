# Contributing to Pact go

1. Fork it
1. Create your feature branch (git checkout -b my-new-feature)
1. Commit your changes (git commit -am 'Add some feature')
1. Push to the branch (git push origin my-new-feature)
1. Create new Pull Request

## Commit messages

Pact Go uses the [Conventional Changelog](https://github.com/bcoe/conventional-changelog-standard/blob/master/convention.md)
message conventions. Please ensure you follow the guidelines.

If you'd like to get some CLI assistance, getting setup is easy:

```shell
npm install commitizen -g
npm i -g cz-conventional-changelog
```

`git cz` to commit and commitizen will guide you.

## Developing

For full integration testing locally, Ruby 2.1.5 must be installed. Under the
hood, Pact Go bundles the
[Pact Mock Service](https://github.com/bethesque/pact-mock_service) and
[Pact Provider Verifier](https://github.com/pact-foundation/pact-provider-verifier)
projects to implement up to v2.0 of the Pact Specification. This is only
temporary, until [Pact Reference](https://github.com/pact-foundation/pact-reference/)
work is completed.

* Git clone https://github.com/pact-foundation/pact-go.git
* Run `make dev` to build the package and setup the Ruby 'binaries' locally

### Vendoring

We use [Govend](https://github.com/govend/govend) to vendor packages. Please ensure
any new packages are added to `vendor.yml` prior to patching.

## Integration Tests

Before releasing a new version, in addition to the standard (isolated) tests
we smoke test the key features against a running Daemon and Broker.

1. Start daemon:

  ```
  go build .
  ./pact-go daemon
  ```

2. Start a broker

  See [Pact Broker](https://github.com/bethesque/pact_broker#usage) for details.
  Make sure you have basic auth setup so we can test authentication.

3. Run the integrated tests

```
cd dsl
PACT_INTEGRATED_TESTS=1 PACT_BROKER_USERNAME="pactuser" PACT_BROKER_PASSWORD="pactpassword" PACT_BROKER_HOST="http://pactbroker" go test -run TestPact_Integration
```
