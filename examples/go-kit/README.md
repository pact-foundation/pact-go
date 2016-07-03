# Example - Go kit

[Go Kit](https://github.com/go-kit/kit) is an excellent framework for building
microservices.

The following example is a simple Login UI ([Consumer](#consumer)) that calls a
User Service ([Provider](#provider)) using JSON over HTTP.

The API currently exposes a single `Login` endpoint at `POST /users/login`, which
the Consumer uses to authenticate a User.

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


## Provider

The "Provider" is a real Go Kit Endpoint (following the Profile Service [example](https://github.com/go-kit/kit/tree/master/examples/profilesvc)),
exposing a single `/users/login` API call:

```
cd provider
go test -v .
```

This will spin up the Provider API with extra routes added for the handling of
provider states, run the verification process and report back success/failure.

### Running the Provider

The provider can be run as a standalone service:

```
go run cmd/usersvc/main.go

# 200
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -H "Postman-Token: 86f8923a-2139-9223-06f5-6d154006ac42" -d '{
  "username":"billy",
  "password":"issilly"
}' "http://localhost:8080/users/login"

# 403
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -H "Postman-Token: 86f8923a-2139-9223-06f5-6d154006ac42" -d '{
  "username":"billy",
  "password":"issilly"
}' "http://localhost:8080/users/login"

# 404
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -H "Postman-Token: 86f8923a-2139-9223-06f5-6d154006ac42" -d '{
  "username":"someoneelse",
  "password":"issilly"
}' "http://localhost:8080/users/login"
```

## Consumer

The "Consumer" is a very simple web application exposing a login form and an
authenticated page. In this example it is helpful to assume that the UI (Consumer)
and the API (Provider) are in separated code bases, maintained by separate teams.

Note that in the Pact testing, we test the `loginHandler` function (an `http.HandlerFunc`)
to test the remote interface and we don't just test the remote interface with
raw http calls. This is important as it means we are testing the remote interface
to our collaborator, not something completely synthetic.

```
cd consumer
go test -v .
```

This will generate a Pact file in `./pacts/billy-bobby.json`.

### Running the Consumer

Before you can run the consumer make sure the provider is
[running](#running-the-provider).

```
go run cmd/web/main.go
```

Hit http://localhost:8081/ in your browser. You can use the username/password of
"billy" / "issilly" to authenticate.
