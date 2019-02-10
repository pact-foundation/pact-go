package provider

import ex "github.com/pact-foundation/pact-go/examples/types"

// UserRepository is an in-memory user database.
type UserRepository struct {
	users map[string]*ex.User
}

// ByUsername finds a user by their username.
func (u *UserRepository) ByUsername(username string) (*ex.User, error) {
	if user, ok := u.users[username]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}
