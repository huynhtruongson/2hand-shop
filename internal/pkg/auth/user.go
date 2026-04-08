package auth

type User struct {
	id   string
	role string
}

func (u User) UserID() string {
	return u.id
}

func (u User) UserRole() string {
	return u.role
}

func (u User) IsAdmin() bool {
	return u.role == "admin"
}

// NewUser constructs a User with the given id and role.
// Exported to allow construction from packages that import auth (e.g. application-layer tests).
func NewUser(id, role string) User {
	return User{id: id, role: role}
}

// Ptr returns a pointer to u. Useful for constructing *User values inline.
func (u User) Ptr() *User {
	return &u
}
