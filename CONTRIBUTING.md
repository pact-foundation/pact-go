# Contributing to Pact go

## Raising defects

Before raising an issue, make sure you have checked the open and closed issues to see if an answer is provided there.
There may also be an answer to your question on [stackoverflow](https://stackoverflow.com/questions/tagged/pact).

Please provide the following information with your issue to enable us to respond as quickly as possible.

1. The relevant versions of the packages you are using.
1. The steps to recreate your issue.
1. An executable code example where possible. You can fork this repository and modify the e2e [examples](https://github.com/pact-foundation/pact-go/blob/master/examples) to quickly recreate your issue.

You can run the E2E tests by:

```sh
make tools   # Assemble the latest Pact Go from Ruby and compile Go
make pact    # Run the Pact tests - consumer + provider
```

## New features / changes

1. Fork it
1. Create your feature branch (git checkout -b my-new-feature)
1. Commit your changes (git commit -am 'Add some feature')
1. Push to the branch (git push origin my-new-feature)
1. Create new Pull Request

### Commit messages

Pact Go uses the [Conventional Changelog](https://github.com/bcoe/conventional-changelog-standard/blob/master/convention.md)
message conventions. Please ensure you follow the guidelines.

If you'd like to get some CLI assistance, getting setup is easy:

```shell
npm install commitizen -g
npm i -g cz-conventional-changelog
```

`git cz` to commit and commitizen will guide you.

### Developing

For full integration testing locally, Ruby 2.2.0 must be installed. Under the
hood, Pact Go bundles the
[Pact Mock Service](https://github.com/bethesque/pact-mock_service) and
[Pact Provider Verifier](https://github.com/pact-foundation/pact-provider-verifier)
projects to implement up to v2.0 of the Pact Specification. This is only
temporary, until [Pact Reference](https://github.com/pact-foundation/pact-reference/)
work is completed.

* Git clone https://github.com/pact-foundation/pact-go.git
* Run `make dev` to build the package and setup the Ruby 'binaries' locally

#### Vendoring

We use [dep](https://github.com/golang/dep) to vendor packages. Please ensure
any new packages that need to have a specific version locked are added to `Gopkg.toml`.

## Integration Tests

Before releasing a new version, in addition to the standard (isolated) tests
we smoke test the key features against the latest code and Broker.

Run `make pact` to run the integration tests.
