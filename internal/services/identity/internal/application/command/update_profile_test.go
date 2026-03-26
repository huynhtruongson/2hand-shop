package command

import (
	"context"
	"errors"
	"testing"

	pkgerrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application/mocks"
)

// ── UpdateProfile tests ─────────────────────────────────────────────────────

func TestUpdateProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		user             *mocks.MockUserRepo
		db               *mocks.MockDB
		cmd              UpdateProfile
		wantErr          bool
		wantCode         string
		wantKind         pkgerrors.Kind
		checkFn          func(t *testing.T, repo *mocks.MockUserRepo, cmd UpdateProfile)
	}{
		{
			name: "success",
			user: func() *mocks.MockUserRepo {
				repo := &mocks.MockUserRepo{}
				repo.GetUserByIDResp = mocks.TestUser("user-123")
				return repo
			}(),
			db:  &mocks.MockDB{},
			cmd: UpdateProfile{UserID: "user-123", Name: "New Name", Gender: "female"},
			checkFn: func(t *testing.T, repo *mocks.MockUserRepo, cmd UpdateProfile) {
				if !repo.UpdateProfileCalled {
					t.Error("expected UpdateUserProfile to be called")
				}
				if repo.UpdateProfileArg == nil {
					t.Fatal("expected UpdateProfileArg to be set")
				}
				if repo.UpdateProfileArg.Name() != cmd.Name {
					t.Errorf("Name() = %q, want %q", repo.UpdateProfileArg.Name(), cmd.Name)
				}
				if repo.UpdateProfileArg.Gender() != cmd.Gender {
					t.Errorf("Gender() = %q, want %q", repo.UpdateProfileArg.Gender(), cmd.Gender)
				}
			},
		},
		{
			name:             "error — user not found",
			user:             &mocks.MockUserRepo{GetUserByIDErr: mocks.NotFoundErr("USER_NOT_FOUND", "user not found")},
			db:               &mocks.MockDB{},
			cmd:              UpdateProfile{UserID: "user-123", Name: "New Name", Gender: "female"},
			wantErr:          true,
			wantCode:         "USER_NOT_FOUND",
			wantKind:         pkgerrors.KindNotFound,
		},
		{
			name:             "error — GetUserByID returns DB error",
			user:             &mocks.MockUserRepo{GetUserByIDErr: mocks.InternalErr("DB_ERROR", "db error")},
			db:               &mocks.MockDB{},
			cmd:              UpdateProfile{UserID: "user-123", Name: "New Name", Gender: "female"},
			wantErr:          true,
			wantCode:         "DB_ERROR",
			wantKind:         pkgerrors.KindInternal,
		},
		{
			name: "error — UpdateUserProfile returns DB error",
			user: func() *mocks.MockUserRepo {
				repo := &mocks.MockUserRepo{}
				repo.GetUserByIDResp = mocks.TestUser("user-123")
				repo.UpdateProfileErr = mocks.InternalErr("UPDATE_PROFILE_FAILED", "update profile failed")
				return repo
			}(),
			db:       &mocks.MockDB{},
			cmd:      UpdateProfile{UserID: "user-123", Name: "New Name", Gender: "female"},
			wantErr:  true,
			wantCode: "UPDATE_PROFILE_FAILED",
			wantKind: pkgerrors.KindInternal,
		},
		{
			name:    "error — transaction start fails",
			user:    func() *mocks.MockUserRepo { repo := &mocks.MockUserRepo{}; repo.GetUserByIDResp = mocks.TestUser("user-123"); return repo }(),
			db:      &mocks.MockDB{TxStarterErr: mocks.InternalErr("TX_BEGIN_FAILED", "begin tx failed")},
			cmd:     UpdateProfile{UserID: "user-123", Name: "New Name", Gender: "female"},
			wantErr: true,
			wantKind: pkgerrors.KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewUpdateProfileHandler(tt.db, tt.user)
			_, err := handler.Handle(context.Background(), tt.cmd)

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
			if tt.checkFn != nil {
				tt.checkFn(t, tt.user, tt.cmd)
			}
		})
	}
}
