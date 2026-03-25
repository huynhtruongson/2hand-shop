package valueobject

import "errors"

type UserRole struct {
	value string
}

var (
	UserRoleAdmin  = UserRole{"admin"}
	UserRoleClient = UserRole{"client"}
)

func (r UserRole) String() string { return r.value }
func (r UserRole) IsAdmin() bool  { return r.value == "admin" }

func NewUserRoleFromString(role string) (UserRole, error) {
	switch role {
	case "client":
		return UserRoleClient, nil
	case "admin":
		return UserRoleAdmin, nil
	}
	return UserRole{}, errors.New("invalid user role")
}
