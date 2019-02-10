# Example - HTTP Mux

Example using the standard libraries Mux Router.

The following example is a simple Login UI ([Consumer](#consumer)) that calls a
User Service ([Provider](#provider)) using JSON over HTTP.

The API currently exposes a single `Login` endpoint at `POST /users/login/:id`, which
the Consumer uses to authenticate a User.

We test 3 scenarios, highlighting the use of [Provider States](/pact-foundation/pact-go#provider#provider-states):

1.  When the user "jmarie" exists, and we perform a login, we expect an HTTP `200`
1.  When the user "jmarie" does not exists, and we perform a login, we expect an HTTP `404`
1.  When the user "jmarie" is unauthorized, and we perform a login, we expect an HTTP `403`

# Getting started

Before any of these tests can be run, ensure Pact Go is installed and run the
daemon in the background:

```
go get ./...
```

## Provider

The "Provider" is a real HTTP API containing the `/users/login/:id` API call:

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
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
  "username":"jmarie",
  "password":"issilly"
}' "http://localhost:8080/users/login/1"

# 403
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
  "username":"jmarie",
  "password":"issilly"
}' "http://localhost:8080/users/login/1"

# 404
curl -v -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d '{
  "username":"someoneelse",
  "password":"issilly"
}' "http://localhost:8080/users/login/1"
```
