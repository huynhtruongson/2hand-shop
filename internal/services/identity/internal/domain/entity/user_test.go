package entity

import (
	"testing"
	"time"

	pkgerrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	validID := "user-123"
	validEmail := "test@example.com"
	validName := "John Doe"
	validGender := "male"
	validRole := valueobject.UserRoleClient

	tests := []struct {
		name      string
		id        string
		email     string
		userName  string
		gender    string
		role      valueobject.UserRole
		wantErr   bool
		wantField string
		wantDetail string
	}{
		{
			name:     "valid user — client role",
			id:       validID,
			email:    validEmail,
			userName: validName,
			gender:   validGender,
			role:     validRole,
			wantErr:  false,
		},
		{
			name:     "valid user — admin role",
			id:       validID,
			email:    validEmail,
			userName: validName,
			gender:   validGender,
			role:     valueobject.UserRoleAdmin,
			wantErr:  false,
		},
		{
			name:      "empty id",
			id:        "",
			email:     validEmail,
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "id",
			wantDetail: "id is empty",
		},
		{
			name:      "whitespace-only id",
			id:        "   ",
			email:     validEmail,
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "id",
			wantDetail: "id is empty",
		},
		{
			name:      "empty email",
			id:        validID,
			email:     "",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email is empty",
		},
		{
			name:      "whitespace-only email",
			id:        validID,
			email:     "   ",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email is empty",
		},
		{
			name:      "invalid email — no @",
			id:        validID,
			email:     "notanemail",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email format is invalid",
		},
		{
			name:      "invalid email — no domain",
			id:        validID,
			email:     "user@",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email format is invalid",
		},
		{
			name:      "invalid email — no local part",
			id:        validID,
			email:     "@example.com",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email format is invalid",
		},
		{
			name:      "invalid email — spaces",
			id:        validID,
			email:     "user @example.com",
			userName:  validName,
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "email",
			wantDetail: "email format is invalid",
		},
		{
			name:      "empty name",
			id:        validID,
			email:     validEmail,
			userName:  "",
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "name",
			wantDetail: "name is empty",
		},
		{
			name:      "whitespace-only name",
			id:        validID,
			email:     validEmail,
			userName:  "   ",
			gender:    validGender,
			role:      validRole,
			wantErr:   true,
			wantField: "name",
			wantDetail: "name is empty",
		},
		{
			name:      "empty gender",
			id:        validID,
			email:     validEmail,
			userName:  validName,
			gender:    "",
			role:      validRole,
			wantErr:   true,
			wantField: "gender",
			wantDetail: "gender is empty",
		},
		{
			name:      "whitespace-only gender",
			id:        validID,
			email:     validEmail,
			userName:  validName,
			gender:    "  ",
			role:      validRole,
			wantErr:   true,
			wantField: "gender",
			wantDetail: "gender is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, err := NewUser(tt.id, tt.email, tt.userName, tt.gender, tt.role)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				ae, ok := err.(*pkgerrors.AppError)
				if !ok {
					t.Fatalf("expected *AppError, got %T", err)
				}
				if ae.Code() != "VALIDATION_ERROR" {
					t.Errorf("expected code VALIDATION_ERROR, got %s", ae.Code())
				}
				if ae.Kind() != pkgerrors.KindValidation {
					t.Errorf("expected kind %s, got %s", pkgerrors.KindValidation, ae.Kind())
				}
				if tt.wantField != "" {
					if ae.UserFacing().Details == nil {
						t.Fatalf("expected details map, got nil")
					}
					if detail, ok := ae.UserFacing().Details[tt.wantField]; !ok {
						t.Errorf("expected detail key %q in details %v", tt.wantField, ae.UserFacing().Details)
					} else if detail != tt.wantDetail {
						t.Errorf("expected detail %q, got %q", tt.wantDetail, detail)
					}
				}
				if u != nil {
					t.Errorf("expected nil user on error, got %v", u)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if u == nil {
					t.Fatal("expected user, got nil")
				}
				if u.ID() != tt.id {
					t.Errorf("ID() = %q, want %q", u.ID(), tt.id)
				}
				if u.Email() != tt.email {
					t.Errorf("Email() = %q, want %q", u.Email(), tt.email)
				}
				if u.Name() != tt.userName {
					t.Errorf("Name() = %q, want %q", u.Name(), tt.userName)
				}
				if u.Gender() != tt.gender {
					t.Errorf("Gender() = %q, want %q", u.Gender(), tt.gender)
				}
				if u.Role() != tt.role {
					t.Errorf("Role() = %v, want %v", u.Role(), tt.role)
				}
				if u.verifiedAt != nil {
					t.Errorf("expected verifiedAt nil, got %v", u.verifiedAt)
				}
				if u.deletedAt != nil {
					t.Errorf("expected deletedAt nil, got %v", u.deletedAt)
				}
				if u.createdAt.IsZero() {
					t.Error("expected createdAt to be non-zero")
				}
				if u.updatedAt.IsZero() {
					t.Error("expected updatedAt to be non-zero")
				}
			}
		})
	}
}

func TestNewUser_TimestampsAreSet(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC().Add(-time.Second)
	u, err := NewUser("id", "a@b.com", "name", "m", valueobject.UserRoleClient)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	if u.CreatedAt().Before(before) || u.CreatedAt().After(after) {
		t.Errorf("CreatedAt %v not in expected window [%v, %v]", u.CreatedAt(), before, after)
	}
	if u.UpdatedAt().Before(before) || u.UpdatedAt().After(after) {
		t.Errorf("UpdatedAt %v not in expected window [%v, %v]", u.UpdatedAt(), before, after)
	}
}

func TestUser_UpdateProfile(t *testing.T) {
	t.Parallel()

	u, err := NewUser("id", "a@b.com", "old-name", "male", valueobject.UserRoleClient)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	updatedAtBefore := u.UpdatedAt()

	tests := []struct {
		name      string
		newName   string
		newGender string
		wantErr   bool
		wantName  string
		wantGender string
	}{
		{
			name:       "valid update",
			newName:    "new-name",
			newGender:  "female",
			wantErr:    false,
			wantName:   "new-name",
			wantGender: "female",
		},
		{
			name:      "empty name",
			newName:   "",
			newGender: "female",
			wantErr:   true,
		},
		{
			name:      "whitespace name",
			newName:   "  ",
			newGender: "female",
			wantErr:   true,
		},
		{
			name:      "empty gender",
			newName:   "new-name",
			newGender: "",
			wantErr:   true,
		},
		{
			name:      "whitespace gender",
			newName:   "new-name",
			newGender: "  ",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// reset to baseline
			u.name = "old-name"
			u.gender = "male"
			u.updatedAt = updatedAtBefore

			err := u.UpdateProfile(tt.newName, tt.newGender)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if u.Name() != tt.wantName {
					t.Errorf("Name() = %q, want %q", u.Name(), tt.wantName)
				}
				if u.Gender() != tt.wantGender {
					t.Errorf("Gender() = %q, want %q", u.Gender(), tt.wantGender)
				}
				if !u.UpdatedAt().After(updatedAtBefore) {
					t.Error("UpdatedAt was not advanced after UpdateProfile")
				}
			}
		})
	}
}

func TestUser_WithAuthProvider(t *testing.T) {
	t.Parallel()

	u, err := NewUser("id", "a@b.com", "name", "m", valueobject.UserRoleClient)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	u.WithAuthProvider("google", "google-sub-123")

	if u.AuthProvider() != "google" {
		t.Errorf("AuthProvider() = %q, want %q", u.AuthProvider(), "google")
	}
	if u.AuthProviderID() != "google-sub-123" {
		t.Errorf("AuthProviderID() = %q, want %q", u.AuthProviderID(), "google-sub-123")
	}
}

func TestUser_IsVerified(t *testing.T) {
	t.Parallel()

	t.Run("not verified", func(t *testing.T) {
		t.Parallel()
		u, err := NewUser("id", "a@b.com", "name", "m", valueobject.UserRoleClient)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		if u.IsVerified() {
			t.Error("expected IsVerified() = false, got true")
		}
	})

	t.Run("verified", func(t *testing.T) {
		t.Parallel()
		now := time.Now().UTC()
		u, err := NewUser("id", "a@b.com", "name", "m", valueobject.UserRoleClient)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		u.verifiedAt = &now
		if !u.IsVerified() {
			t.Error("expected IsVerified() = true, got false")
		}
	})
}

func TestUser_UnmarshalUserFromDB(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	verifiedAt := time.Now().UTC().Add(-time.Hour)
	deletedAt := time.Now().UTC().Add(-time.Minute)

	u := UnmarshalUserFromDB(
		"user-123",
		"google",
		"google-456",
		"test@example.com",
		"John Doe",
		"male",
		&verifiedAt,
		valueobject.UserRoleAdmin,
		now,
		now,
		&deletedAt,
	)

	if u.ID() != "user-123" {
		t.Errorf("ID() = %q, want %q", u.ID(), "user-123")
	}
	if u.AuthProvider() != "google" {
		t.Errorf("AuthProvider() = %q, want %q", u.AuthProvider(), "google")
	}
	if u.AuthProviderID() != "google-456" {
		t.Errorf("AuthProviderID() = %q, want %q", u.AuthProviderID(), "google-456")
	}
	if u.Email() != "test@example.com" {
		t.Errorf("Email() = %q, want %q", u.Email(), "test@example.com")
	}
	if u.Name() != "John Doe" {
		t.Errorf("Name() = %q, want %q", u.Name(), "John Doe")
	}
	if u.Gender() != "male" {
		t.Errorf("Gender() = %q, want %q", u.Gender(), "male")
	}
	if u.VerifiedAt() == nil || !u.VerifiedAt().Equal(verifiedAt) {
		t.Errorf("VerifiedAt() = %v, want %v", u.VerifiedAt(), verifiedAt)
	}
	if u.Role() != valueobject.UserRoleAdmin {
		t.Errorf("Role() = %v, want %v", u.Role(), valueobject.UserRoleAdmin)
	}
	if u.CreatedAt() != now {
		t.Errorf("CreatedAt() = %v, want %v", u.CreatedAt(), now)
	}
	if u.UpdatedAt() != now {
		t.Errorf("UpdatedAt() = %v, want %v", u.UpdatedAt(), now)
	}
	if u.DeletedAt() == nil || !u.DeletedAt().Equal(deletedAt) {
		t.Errorf("DeletedAt() = %v, want %v", u.DeletedAt(), deletedAt)
	}

	// nil deletedAt
	u2 := UnmarshalUserFromDB(
		"user-456", "", "", "a@b.com", "Jane", "female",
		nil, valueobject.UserRoleClient,
		now, now, nil,
	)
	if u2.DeletedAt() != nil {
		t.Errorf("DeletedAt() = %v, want nil", u2.DeletedAt())
	}
	if u2.VerifiedAt() != nil {
		t.Errorf("VerifiedAt() = %v, want nil", u2.VerifiedAt())
	}
}

func TestUser_Getters(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	verifiedAt := time.Now().UTC()
	deletedAt := time.Now().UTC()

	u := UnmarshalUserFromDB(
		"id", "provider", "provider-id",
		"test@example.com", "Test User", "other",
		&verifiedAt, valueobject.UserRoleAdmin,
		now, now, &deletedAt,
	)

	if u.ID() != "id" {
		t.Errorf("ID() = %q, want %q", u.ID(), "id")
	}
	if u.AuthProvider() != "provider" {
		t.Errorf("AuthProvider() = %q, want %q", u.AuthProvider(), "provider")
	}
	if u.AuthProviderID() != "provider-id" {
		t.Errorf("AuthProviderID() = %q, want %q", u.AuthProviderID(), "provider-id")
	}
	if u.Email() != "test@example.com" {
		t.Errorf("Email() = %q, want %q", u.Email(), "test@example.com")
	}
	if u.Name() != "Test User" {
		t.Errorf("Name() = %q, want %q", u.Name(), "Test User")
	}
	if u.Gender() != "other" {
		t.Errorf("Gender() = %q, want %q", u.Gender(), "other")
	}
	if u.Role() != valueobject.UserRoleAdmin {
		t.Errorf("Role() = %v, want %v", u.Role(), valueobject.UserRoleAdmin)
	}
	if u.VerifiedAt() == nil || !u.VerifiedAt().Equal(verifiedAt) {
		t.Errorf("VerifiedAt() = %v, want %v", u.VerifiedAt(), verifiedAt)
	}
	if u.CreatedAt() != now {
		t.Errorf("CreatedAt() = %v, want %v", u.CreatedAt(), now)
	}
	if u.UpdatedAt() != now {
		t.Errorf("UpdatedAt() = %v, want %v", u.UpdatedAt(), now)
	}
	if u.DeletedAt() == nil || !u.DeletedAt().Equal(deletedAt) {
		t.Errorf("DeletedAt() = %v, want %v", u.DeletedAt(), deletedAt)
	}
}
