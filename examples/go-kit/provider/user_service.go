package provider

import (
	"errors"

	ex "github.com/pact-foundation/pact-go/examples/types"
)

var (
	// ErrNotFound represents a resource not found (404)
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized represents a Unauthorized (401)
	ErrUnauthorized = errors.New("unauthorized")

	// ErrEmpty is returned when input string is empty
	ErrEmpty = errors.New("empty string")
)

// Service provides operations on Users.
type Service interface {
	Login(string, string) (*ex.User, error)
}

type userService struct {
	userDatabase *UserRepository
}

// NewInmemService gets you a shiny new UserService!
func NewInmemService() Service {
	return &userService{
		userDatabase: &UserRepository{
			users: map[string]*ex.User{
				"jmarie": &ex.User{
					Name:     "Jean-Marie de La Beaujardière😀😍",
					Username: "jmarie",
					Password: "issilly",
					Type:     "admin",
				},
			},
		},
	}
}

// Login to the system.
func (u *userService) Login(username string, password string) (user *ex.User, err error) {
	if user, err = u.userDatabase.ByUsername(username); err != nil {
		return nil, ErrNotFound
	}

	if user.Username != username || user.Password != password {
		return nil, ErrUnauthorized
	}
	return user, nil
}
