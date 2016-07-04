// Package types contains types to use across the Consumer/Provider tests.
package types

// UserRepository is an in-memory user database.
type UserRepository struct {
	Users map[string]*User
}

// ByUsername finds a user by their username.
func (u *UserRepository) ByUsername(username string) (*User, error) {
	if user, ok := u.Users[username]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}
