// Package avro contains a runnable Pact example using Avro-encoded
// message contents.
package avro

// User is the example message payload used by the avro consumer and
// provider tests.
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}
