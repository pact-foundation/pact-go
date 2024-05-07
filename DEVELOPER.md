# Developer documentation

## Tooling

* Docker
* Java (>= 19) - required for the Avro plugin example

## Key Branches

### `1.x.x` 

The previous major version. Only bug fixes and security updates will be considered.

### `master`

The `2.x.x` release line. Current major version.


## Running locally

### Pre-reqs

- GoLang

### Unit tests

```sh
make ci_unit
```

### E2E Examples

#### E2E Pre-reqs

- Pact CLI tools (for examples). Available via
  - Docker
  - Ruby
  - Standalone package
  - 
- Pact Broker
  - Locally via Docker
  - Hosted (OSS or PactFlow)
- Java
  - For Avro example
- Protobuf Compiler
  - For gRPC / Protobuf examples

#### Running

With Docker

```sh
APP_SHA=foo make ci_examples
```

Without Docker

```sh
APP_SHA=foo make ci_hosted_examples
```

You will need to use a Pact Broker to run the examples.

You can either use an OSS Pact Broker and set the following env vars

- `PACT_BROKER_BASE_URL`
- `PACT_BROKER_USERNAME`
- `PACT_BROKER_PASSWORD`

Or a PactFlow Broker

- `PACT_BROKER_BASE_URL`
- `PACT_BROKER_TOKEN`

To use the different cli tools, use

- `PACT_TOOL=standalone make ci_examples`
  - Install it `make install-pact-ruby-standalone`
  - https://github.com/pact-foundation/pact-ruby-standalone
- `PACT_TOOL=ruby make ci_examples`
  - Install it `make install-pact-ruby-cli`
  - Requires ruby
  - https://github.com/pact-foundation/pact_broker-client
- `PACT_TOOL=docker make ci_examples`
  - Requires docker
  - https://github.com/pact-foundation/pact-ruby-cli
You will want to install them first

### Docker

Build an image from the repo

```sh
make docker_build
```

Run the repo's unit test

```
make docker_run_test
```

Run the repo's examples

```
make docker_run_examples
```