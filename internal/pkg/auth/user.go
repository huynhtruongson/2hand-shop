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
