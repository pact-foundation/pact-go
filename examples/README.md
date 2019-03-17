# Examples

This folder contains a number of examples in different frameworks to demonstrate
how Pact could be used in each.

Each Provider API currently exposes a single `Login` endpoint at `POST /login/1`,
which the [Consumer](consumer/goconsumer) uses to authenticate a User.

We test 5 scenarios, highlighting the use of [Provider States](/pact-foundation/pact-go#provider#provider-states), [Hooks](/pact-foundation/pact-go#before-and-after-hooks) and [RequestFilters](/pact-foundation/pact-go#request-filters):

1.  When the user "jmarie" exists, and we perform a login, we expect an HTTP `200`
1.  When the user "jmarie" does not exists, and we perform a login, we expect an HTTP `404`
1.  When the user "jmarie" is unauthorized, and we perform a login, we expect an HTTP `403`
1.  When the user is authenticated, and we request to get the user 'jmarie', we expect an HTTP `200`
1.  When the user is unauthenticated, and we request to get the user 'jmarie', we expect an HTTP `401`

# Getting started

Before any of these tests can be run, ensure Pact Go is installed and run the
daemon in the background:

```
go get ./...
```

## Providers

1.  [Mux](mux)
2.  [Gin](gin)

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

This will generate a Pact file in `./pacts/jmarie-loginprovider.json`.

### Running the Consumer

Before you can run the consumer make sure one of the providers is
running first. You can then run:

```
go run cmd/web/main.go
```

Hit http://localhost:8081/ in your browser. You can use the username/password
combination of "jmarie" / "issilly" to authenticate.
