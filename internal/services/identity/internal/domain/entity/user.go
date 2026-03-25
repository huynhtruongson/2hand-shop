package entity

import (
	"regexp"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
)

type User struct {
	id             string
	authProvider   string
	authProviderID string
	email          string
	name           string
	gender         string
	verifiedAt     *time.Time
	role           valueobject.UserRole
	createdAt      time.Time
	updatedAt      time.Time
	deletedAt      *time.Time
}

var (
	// RFC 5322 simplified — covers virtually all real-world emails
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// E.164 format: +[country code][number], 8–15 digits total
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
)

func NewUser(id, email, name, gender string, role valueobject.UserRole) (*User, error) {
	u := User{
		id:         id,
		email:      email,
		name:       name,
		gender:     gender,
		verifiedAt: nil,
		role:       role,
		createdAt:  time.Now().UTC(),
		updatedAt:  time.Now().UTC(),
		deletedAt:  nil,
	}
	if err := u.validate(); err != nil {
		return nil, err
	}
	return &u, nil
}

func (u *User) ID() string                 { return u.id }
func (u *User) AuthProvider() string       { return u.authProvider }
func (u *User) AuthProviderID() string     { return u.authProviderID }
func (u *User) Email() string              { return u.email }
func (u *User) Name() string               { return u.name }
func (u *User) Gender() string             { return u.gender }
func (u *User) Role() valueobject.UserRole { return u.role }
func (u *User) VerifiedAt() *time.Time     { return u.verifiedAt }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) UpdatedAt() time.Time       { return u.updatedAt }
func (u *User) DeletedAt() *time.Time      { return u.deletedAt }

func (u *User) IsVerified() bool {
	return u.verifiedAt != nil
}

func (u *User) UpdateProfile(name, gender string) error {
	u.name = name
	u.gender = gender
	u.updatedAt = time.Now().UTC()
	return u.validate()
}

func (u *User) WithAuthProvider(authProvider string, authProviderID string) {
	u.authProvider = authProvider
	u.authProviderID = authProviderID
}

func UnmarshalUserFromDB(id string,
	authProvider string,
	authProviderID string,
	email string,
	name string,
	gender string,
	verifiedAt *time.Time,
	role valueobject.UserRole,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time) *User {
	return &User{
		id:             id,
		authProvider:   authProvider,
		authProviderID: authProviderID,
		email:          email,
		name:           name,
		gender:         gender,
		verifiedAt:     verifiedAt,
		role:           role,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		deletedAt:      deletedAt,
	}
}

func (u *User) validate() error {
	switch {
	case strings.TrimSpace(u.id) == "":
		return errors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(u.email) == "":
		return errors.ErrValidation.WithDetail("email", "email is empty")
	case !emailRegex.MatchString(u.email):
		return errors.ErrValidation.WithDetail("email", "email format is invalid")
	case strings.TrimSpace(u.name) == "":
		return errors.ErrValidation.WithDetail("name", "name is empty")
	case strings.TrimSpace(u.gender) == "":
		return errors.ErrValidation.WithDetail("gender", "gender is empty")
	}

	return nil
}
