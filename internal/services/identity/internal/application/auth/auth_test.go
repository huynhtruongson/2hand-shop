package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	pkgerrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

// ── mockAuthProvider ─────────────────────────────────────────────────────────

type mockAuthProvider struct {
	name          string
	signUpResp    *AuthProviderSignUpResp
	signUpErr     error
	signInResp    *AuthProviderSignInResp
	signInErr     error
	confirmErr    error
	confirmCalled bool
	confirmEmail  string
	confirmCode   string
}

func (m *mockAuthProvider) ProviderName() string { return m.name }

func (m *mockAuthProvider) SignUp(ctx context.Context, email, password string, attrs map[string]string) (*AuthProviderSignUpResp, error) {
	return m.signUpResp, m.signUpErr
}

func (m *mockAuthProvider) SignIn(ctx context.Context, email, password string) (*AuthProviderSignInResp, error) {
	return m.signInResp, m.signInErr
}

func (m *mockAuthProvider) ConfirmAccount(ctx context.Context, email, code string) error {
	m.confirmCalled = true
	m.confirmEmail = email
	m.confirmCode = code
	return m.confirmErr
}

// ── mockUserRepo ─────────────────────────────────────────────────────────────

type mockUserRepo struct {
	getUserByEmailResp       *entity.User
	getUserByEmailErr        error
	createUserCalled         bool
	createUserArg            *entity.User
	createUserErr            error
	updateUserVerifiedCalled bool
	updateUserVerifiedEmail  string
	updateUserVerifiedErr    error
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, db postgressqlx.Querier, email string) (*entity.User, error) {
	return m.getUserByEmailResp, m.getUserByEmailErr
}

func (m *mockUserRepo) CreateUser(ctx context.Context, db postgressqlx.Querier, user *entity.User) error {
	m.createUserCalled = true
	m.createUserArg = user
	return m.createUserErr
}

func (m *mockUserRepo) UpdateUserVerified(ctx context.Context, db postgressqlx.Querier, email string) error {
	m.updateUserVerifiedCalled = true
	m.updateUserVerifiedEmail = email
	return m.updateUserVerifiedErr
}

// ── mockDB ───────────────────────────────────────────────────────────────────

type mockDB struct {
	txStarterErr error
}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (postgressqlx.TX, error) {
	if m.txStarterErr != nil {
		return nil, m.txStarterErr
	}
	return &mockTX{}, nil
}
func (m *mockDB) Close() error { return nil }
func (m *mockDB) GetContext(ctx context.Context, dest interface{}, query string, args ...any) error {
	return nil
}
func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *mockDB) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}
func (m *mockDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...any) error {
	return nil
}
func (m *mockDB) NamedQuery(query string, arg any) (*sqlx.Rows, error) { return nil, nil }
func (m *mockDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return nil
}
func (m *mockDB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}

// mockTX satisfies postgressqlx.TX; methods are no-ops since auth service
// never queries through the transaction directly.
type mockTX struct{}

func (m *mockTX) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *mockTX) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}
func (m *mockTX) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockTX) NamedQuery(query string, arg any) (*sqlx.Rows, error) { return nil, nil }
func (m *mockTX) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return nil
}
func (m *mockTX) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}
func (m *mockTX) Rollback() error { return nil }
func (m *mockTX) Commit() error   { return nil }

// ── mockLogger ───────────────────────────────────────────────────────────────

type mockLogger struct{}

func (m *mockLogger) Info(msg string, keysAndValues ...any)   {}
func (m *mockLogger) Debug(msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...any)   {}
func (m *mockLogger) Error(msg string, keysAndValues ...any)  {}
func (m *mockLogger) Fatal(msg string, keysAndValues ...any)  {}
func (m *mockLogger) With(keysAndValues ...any) logger.Logger { return m }

// ── helpers ─────────────────────────────────────────────────────────────────

// verifiedUser builds a verified user via UnmarshalUserFromDB (avoids
// accessing unexported struct fields directly).
func verifiedUser(email string) *entity.User {
	now := time.Now().UTC()
	return entity.UnmarshalUserFromDB(
		"user-123", "cognito", "cognito-sub-123",
		email, "Test User", "male",
		&now, valueobject.UserRoleClient,
		now, now, nil,
	)
}

// unverifiedUser builds an unverified user via UnmarshalUserFromDB.
func unverifiedUser(email string) *entity.User {
	now := time.Now().UTC()
	return entity.UnmarshalUserFromDB(
		"user-123", "cognito", "cognito-sub-123",
		email, "Test User", "male",
		nil, valueobject.UserRoleClient,
		now, now, nil,
	)
}

// ── SignUp tests ─────────────────────────────────────────────────────────────

func TestSignUp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		signUpResp        *AuthProviderSignUpResp
		signUpErr         error
		existingUser      *entity.User
		getUserByEmailErr error
		txStarterErr      error
		wantErr           bool
		wantCode          string
		wantKind          pkgerrors.Kind
		checkFn           func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo, result *SignUpResult)
	}{
		{
			name:       "success — user unverified after sign-up",
			signUpResp: &AuthProviderSignUpResp{UserSub: "cognito-sub-123", IsVerified: false},
			wantErr:    false,
			checkFn: func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo, result *SignUpResult) {
				if result.UserID == "" {
					t.Error("expected non-empty UserID")
				}
				if result.IsVerified {
					t.Error("expected IsVerified = false from auth provider")
				}
				if !repo.createUserCalled {
					t.Error("expected CreateUser to be called")
				}
				if repo.createUserArg == nil {
					t.Fatal("expected createUserArg to be set")
				}
				if repo.createUserArg.Email() != "test@example.com" {
					t.Errorf("Email() = %q, want %q", repo.createUserArg.Email(), "test@example.com")
				}
				if repo.createUserArg.AuthProvider() != "cognito" {
					t.Errorf("AuthProvider() = %q, want %q", repo.createUserArg.AuthProvider(), "cognito")
				}
				if repo.createUserArg.AuthProviderID() != "cognito-sub-123" {
					t.Errorf("AuthProviderID() = %q, want %q", repo.createUserArg.AuthProviderID(), "cognito-sub-123")
				}
			},
		},
		{
			name:       "success — user verified after sign-up (email confirmed immediately)",
			signUpResp: &AuthProviderSignUpResp{UserSub: "cognito-sub-123", IsVerified: true},
			wantErr:    false,
			checkFn: func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo, result *SignUpResult) {
				if !repo.createUserCalled {
					t.Error("expected CreateUser to be called")
				}
			},
		},
		{
			name:         "error — user already exists",
			signUpResp:   &AuthProviderSignUpResp{UserSub: "cognito-sub-123"},
			existingUser: verifiedUser("test@example.com"),
			wantErr:      true,
			wantCode:     "USER_ALREADY_EXISTS",
			wantKind:     pkgerrors.KindConflict,
		},
		{
			name:              "error — GetUserByEmail returns non-notfound DB error",
			signUpResp:        &AuthProviderSignUpResp{UserSub: "cognito-sub-123"},
			getUserByEmailErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "DB_ERROR", "db error"),
			wantErr:           true,
			wantCode:          "DB_ERROR",
			wantKind:          pkgerrors.KindInternal,
		},
		{
			name:       "error — auth provider sign-up fails",
			signUpResp: nil,
			signUpErr:  pkgerrors.NewAppError(pkgerrors.KindUnprocessable, "AUTH_SIGNUP_FAILED", "sign-up failed"),
			wantErr:    true,
			wantCode:   "AUTH_SIGNUP_FAILED",
			wantKind:   pkgerrors.KindUnprocessable,
		},
		{
			name:         "error — DB transaction fails",
			signUpResp:  &AuthProviderSignUpResp{UserSub: "cognito-sub-123"},
			txStarterErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "TX_ERROR", "tx error"),
			wantErr:      true,
			wantKind:     pkgerrors.KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider, repo := &mockAuthProvider{name: "cognito"}, &mockUserRepo{}
			provider.signUpResp = tt.signUpResp
			provider.signUpErr = tt.signUpErr

			var txErr error
			if tt.txStarterErr != nil {
				txErr = tt.txStarterErr
			}
			if tt.existingUser != nil {
				repo.getUserByEmailResp = tt.existingUser
			}
			if tt.getUserByEmailErr != nil {
				repo.getUserByEmailErr = tt.getUserByEmailErr
			}

			db := &mockDB{}
			if txErr != nil {
				db.txStarterErr = txErr
			}
			svc := NewAuthService(&mockLogger{}, db, provider, repo)

			result, err := svc.SignUp(context.Background(), SignUpParams{
				Name:     "Test User",
				Email:    "test@example.com",
				Gender:   "male",
				Password: "password123",
			})

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
				tt.checkFn(t, provider, repo, result)
			}
		})
	}
}

// ── SignIn tests ─────────────────────────────────────────────────────────────

func TestSignIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		user              *entity.User
		getUserByEmailErr error
		signInResp        *AuthProviderSignInResp
		signInErr         error
		wantErr           bool
		wantCode          string
		wantKind          pkgerrors.Kind
		checkFn           func(t *testing.T, result *SignInResult)
	}{
		{
			name:   "success",
			user:   verifiedUser("test@example.com"),
			signInResp: &AuthProviderSignInResp{
				AccessToken:  "access-token",
				IDToken:      "id-token",
				RefreshToken: "refresh-token",
				ExpiresIn:    3600,
			},
			wantErr: false,
			checkFn: func(t *testing.T, result *SignInResult) {
				if result.AccessToken != "access-token" {
					t.Errorf("AccessToken = %q, want %q", result.AccessToken, "access-token")
				}
				if result.IDToken != "id-token" {
					t.Errorf("IDToken = %q, want %q", result.IDToken, "id-token")
				}
				if result.RefreshToken != "refresh-token" {
					t.Errorf("RefreshToken = %q, want %q", result.RefreshToken, "refresh-token")
				}
				if result.ExpiresIn != 3600 {
					t.Errorf("ExpiresIn = %d, want %d", result.ExpiresIn, 3600)
				}
			},
		},
		{
			name:              "error — user not found",
			getUserByEmailErr: pkgerrors.NewAppError(pkgerrors.KindNotFound, "USER_NOT_FOUND", "user not found"),
			wantErr:           true,
			wantCode:          "AUTH_INVALID_CREDENTIALS",
			wantKind:          pkgerrors.KindUnauthorized,
		},
		{
			name:       "error — user not verified",
			user:       unverifiedUser("test@example.com"),
			wantErr:    true,
			wantCode:   "USER_NOT_VERIFIED",
			wantKind:   pkgerrors.KindUnprocessable,
		},
		{
			name:       "error — auth provider sign-in fails",
			user:       verifiedUser("test@example.com"),
			signInErr:  pkgerrors.NewAppError(pkgerrors.KindUnauthorized, "AUTH_SIGNIN_FAILED", "sign-in failed"),
			wantErr:    true,
			wantCode:   "AUTH_SIGNIN_FAILED",
			wantKind:   pkgerrors.KindUnauthorized,
		},
		{
			name:              "error — GetUserByEmail returns DB error",
			getUserByEmailErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "DB_ERROR", "db error"),
			signInResp:        &AuthProviderSignInResp{AccessToken: "tok", IDToken: "id", RefreshToken: "ref", ExpiresIn: 3600},
			wantErr:           true,
			wantCode:          "DB_ERROR",
			wantKind:          pkgerrors.KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider, repo := &mockAuthProvider{name: "cognito"}, &mockUserRepo{}
			provider.signInResp = tt.signInResp
			provider.signInErr = tt.signInErr
			if tt.user != nil {
				repo.getUserByEmailResp = tt.user
			}
			if tt.getUserByEmailErr != nil {
				repo.getUserByEmailErr = tt.getUserByEmailErr
			}
			svc := NewAuthService(&mockLogger{}, &mockDB{}, provider, repo)

			result, err := svc.SignIn(context.Background(), "test@example.com", "password123")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !pkgerrors.IsKind(err, tt.wantKind) {
					t.Errorf("expected kind %v, got %v", tt.wantKind, err)
				}
				ae, ok := err.(*pkgerrors.AppError)
				if !ok {
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
				tt.checkFn(t, result)
			}
		})
	}
}

// ── ConfirmAccount tests ─────────────────────────────────────────────────────

func TestConfirmAccount(t *testing.T) {
	t.Parallel()

	const email = "test@example.com"
	const code = "123456"

	tests := []struct {
		name                  string
		user                  *entity.User
		getUserByEmailErr     error
		confirmErr            error
		updateUserVerifiedErr error
		txStarterErr          error
		wantErr               bool
		wantCode              string
		wantKind              pkgerrors.Kind
		checkFn               func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo)
	}{
		{
			name: "success — unverified user",
			user: unverifiedUser(email),
			checkFn: func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo) {
				if !provider.confirmCalled {
					t.Error("expected ConfirmAccount to be called on auth provider")
				}
				if provider.confirmEmail != email {
					t.Errorf("confirmEmail = %q, want %q", provider.confirmEmail, email)
				}
				if provider.confirmCode != code {
					t.Errorf("confirmCode = %q, want %q", provider.confirmCode, code)
				}
				if !repo.updateUserVerifiedCalled {
					t.Error("expected UpdateUserVerified to be called")
				}
				if repo.updateUserVerifiedEmail != email {
					t.Errorf("updateUserVerifiedEmail = %q, want %q", repo.updateUserVerifiedEmail, email)
				}
			},
		},
		{
			name: "success — already verified user",
			user: verifiedUser(email),
			checkFn: func(t *testing.T, provider *mockAuthProvider, repo *mockUserRepo) {
				if !provider.confirmCalled {
					t.Error("expected ConfirmAccount to be called on auth provider")
				}
				if repo.updateUserVerifiedCalled {
					t.Error("expected UpdateUserVerified NOT to be called for already-verified user")
				}
			},
		},
		{
			name:              "error — GetUserByEmail returns DB error",
			getUserByEmailErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "DB_ERROR", "db error"),
			wantErr:           true,
			wantCode:          "DB_ERROR",
			wantKind:          pkgerrors.KindInternal,
		},
		{
			name:       "error — auth provider returns invalid code error",
			user:       unverifiedUser(email),
			confirmErr: pkgerrors.NewAppError(pkgerrors.KindUnprocessable, "AUTH_INVALID_CONFIRMATION_CODE", "invalid confirmation code"),
			wantErr:    true,
			wantCode:   "AUTH_INVALID_CONFIRMATION_CODE",
			wantKind:   pkgerrors.KindUnprocessable,
		},
		{
			name:                  "error — UpdateUserVerified returns DB error",
			user:                  unverifiedUser(email),
			updateUserVerifiedErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "DB_ERROR", "db error"),
			wantErr:               true,
			wantCode:              "DB_ERROR",
			wantKind:              pkgerrors.KindInternal,
		},
		{
			name:         "error — DB transaction fails",
			user:         unverifiedUser(email),
			txStarterErr: pkgerrors.NewAppError(pkgerrors.KindInternal, "TX_ERROR", "tx error"),
			wantErr:      true,
			wantKind:     pkgerrors.KindInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider, repo := &mockAuthProvider{name: "cognito"}, &mockUserRepo{}
			provider.confirmErr = tt.confirmErr

			if tt.user != nil {
				repo.getUserByEmailResp = tt.user
			}
			if tt.getUserByEmailErr != nil {
				repo.getUserByEmailErr = tt.getUserByEmailErr
			}
			if tt.updateUserVerifiedErr != nil {
				repo.updateUserVerifiedErr = tt.updateUserVerifiedErr
			}

			db := &mockDB{}
			if tt.txStarterErr != nil {
				db.txStarterErr = tt.txStarterErr
			}
			svc := NewAuthService(&mockLogger{}, db, provider, repo)

			err := svc.ConfirmAccount(context.Background(), email, code)

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
				tt.checkFn(t, provider, repo)
			}
		})
	}
}
