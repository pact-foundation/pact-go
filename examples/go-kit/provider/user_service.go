package provider

import "errors"

// User is a representation of a User. Dah.
type User struct {
	Name     string `json:"name"`
	username string
	password string
	Type     string `json:"type"`
}

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
	Login(string, string) (*User, error)
}

type userService struct {
	userDatabase *UserRepository
}

// NewInmemService gets you a shiny new UserService!
func NewInmemService() Service {
	return &userService{
		userDatabase: &UserRepository{
			users: map[string]*User{
				"Jean-Marie de La Beaujardière😀😍": &User{
					Name:     "Jean-Marie de La Beaujardière😀😍",
					username: "Jean-Marie de La Beaujardière😀😍",
					password: "issilly",
					Type:     "admin",
				},
			},
		},
	}
}

// Login to the system.
func (u *userService) Login(username string, password string) (user *User, err error) {
	if user, err = u.userDatabase.ByUsername(username); err != nil {
		return nil, ErrNotFound
	}

	if user.username != username || user.password != password {
		return nil, ErrUnauthorized
	}
	return user, nil
}
