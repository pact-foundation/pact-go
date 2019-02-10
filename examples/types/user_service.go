package types

import "errors"

// User is a representation of a User. Dah.
type User struct {
	Name     string `json:"name" pact:"example=Jean-Marie de La Beaujardière😀😍"`
	Username string `json:"username" pact:"example=jmarie"`
	Password string `json:"password" pact:"example=password123"`
	Type     string `json:"type" pact:"example=admin,regex=^(admin|user|guest)$"`
}

var (
	// ErrNotFound represents a resource not found (404)
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized represents a Forbidden (403)
	ErrUnauthorized = errors.New("unauthorized")

	// ErrEmpty is returned when input string is empty
	ErrEmpty = errors.New("empty string")
)

// LoginRequest is the login request API struct.
type LoginRequest struct {
	Username string `json:"username" pact:"example=jmarie"`
	Password string `json:"password" pact:"example=issilly"`
}

// LoginResponse is the login response API struct.
type LoginResponse struct {
	User *User `json:"user"`
}
