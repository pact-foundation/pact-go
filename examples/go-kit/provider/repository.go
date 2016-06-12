package provider

// UserRepository is an in-memory user database.
type UserRepository struct {
	users map[string]*User
}

// ByUsername finds a user by their username.
func (u *UserRepository) ByUsername(username string) (*User, error) {
	if user, ok := u.users[username]; ok {
		return user, nil
	}
	return nil, ErrNotFound
}
