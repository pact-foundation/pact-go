# Examples

This folder contains a number of examples in different frameworks to demonstrate
how Pact could be used in each.

Each Provider API currently exposes a single `Login` endpoint at `POST /users/login`, 
which the [Consumer](consumer/goconsumer) uses to authenticate a User.

We test 3 scenarios, highlighting the use of [Provider States](/pact-foundation/pact-go#provider#provider-states):

1. When the user "Billy" exists, and we perform a login, we expect an HTTP `200`
1. When the user "Billy" does not exists, and we perform a login, we expect an HTTP `404`
1. When the user "Billy" is unauthorized, and we perform a login, we expect an HTTP `403`

# Getting started

Before any of these tests can be run, ensure Pact Go is installed and run the
daemon in the background:

```
go get ./...
<path to>/pact-go daemon
```

## Providers

1. [Go-Kit](go-kit)
2. [Gin](gin)

## Consumer

The "Consumer" is a very simple web application exposing a login form and an
authenticated page. In this example it is helpful to assume that the UI (Consumer)
and the API (Provider) are in separated code bases, maintained by separate teams.

Note that in the Pact testing, we test the `loginHandler` function (an `http.HandlerFunc`)
to test the remote interface and we don't just test the remote interface with
raw http calls. This is important as it means we are testing the remote interface
to our collaborator, not something completely synthetic.

```
cd consumer/goconsumer
go test -v .
```

This will generate a Pact file in `./pacts/billy-bobby.json`.

### Running the Consumer

Before you can run the consumer make sure the provider is
[running](#running-the-provider).

```
go run cmd/web/main.go
```

Hit http://localhost:8081/ in your browser. You can use the username/password 
combination of "billy" / "issilly" to authenticate.
