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
make fake_pact    # Run the Pact tests - consumer + provider
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