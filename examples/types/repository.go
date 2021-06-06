// package v3 contains types to use across the Consumer/Provider tests.
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

// ByID finds a user by their ID
func (u *UserRepository) ByID(ID int) (*User, error) {
	for _, user := range u.Users {
		if user.ID == ID {
			return user, nil
		}
	}
	return nil, ErrNotFound
}
