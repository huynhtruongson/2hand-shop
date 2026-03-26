package query

import (
	"context"
	"errors"
	"testing"

	pkgerrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/mocks"
)

// ── GetProfile (Profile) tests ───────────────────────────────────────────────

func TestGetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repo        *mocks.MockUserRepo
		query       Profile
		wantErr     bool
		wantCode    string
		wantKind    pkgerrors.Kind
		checkFn     func(t *testing.T, got *mocks.MockUserRepo, userID string)
	}{
		{
			name: "success",
			repo: func() *mocks.MockUserRepo {
				repo := &mocks.MockUserRepo{}
				repo.GetUserByIDResp = mocks.TestUser("user-123")
				return repo
			}(),
			query: Profile{ID: "user-123"},
			checkFn: func(t *testing.T, repo *mocks.MockUserRepo, userID string) {
				if repo.GetUserByIDResp == nil {
					t.Fatal("expected GetUserByIDResp to be set")
				}
				if repo.GetUserByIDResp.ID() != userID {
					t.Errorf("user ID = %q, want %q", repo.GetUserByIDResp.ID(), userID)
				}
			},
		},
		{
			name:        "error — user not found",
			repo:        &mocks.MockUserRepo{GetUserByIDErr: mocks.NotFoundErr("USER_NOT_FOUND", "user not found")},
			query:       Profile{ID: "user-999"},
			wantErr:     true,
			wantCode:    "USER_NOT_FOUND",
			wantKind:    pkgerrors.KindNotFound,
		},
		{
			name:        "error — database error",
			repo:        &mocks.MockUserRepo{GetUserByIDErr: mocks.InternalErr("DB_ERROR", "database error")},
			query:       Profile{ID: "user-123"},
			wantErr:     true,
			wantCode:    "DB_ERROR",
			wantKind:    pkgerrors.KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := &mocks.MockDB{} // Profile handler only calls userRepo.GetUserByID; DB is unused.
			handler := NewProfileHandler(db, tt.repo)
			got, err := handler.Handle(context.Background(), tt.query)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !pkgerrors.IsKind(err, tt.wantKind) {
					t.Errorf("expected kind %v, got %v", tt.wantKind, err)
				}
				var ae *pkgerrors.AppError
				if !errors.As(err, &ae) {
					t.Fatalf("expected *AppError, got %T", err)
				}
				if tt.wantCode != "" && ae.Code() != tt.wantCode {
					t.Errorf("expected code %q, got %q", tt.wantCode, ae.Code())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected user, got nil")
			}
			if tt.checkFn != nil {
				tt.checkFn(t, tt.repo, tt.query.ID)
			}
		})
	}
}
