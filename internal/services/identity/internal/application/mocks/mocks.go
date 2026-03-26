// Package mocks provides test doubles shared across the application layer.
package mocks

import (
	"context"
	"database/sql"
	"time"

	pkgerrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

// ── UserRepo ─────────────────────────────────────────────────────────────────

// MockUserRepo is a test double for repository.UserRepo.
type MockUserRepo struct {
	GetUserByIDResp     *entity.User
	GetUserByIDErr      error
	UpdateProfileErr    error
	UpdateProfileCalled bool
	UpdateProfileArg    *entity.User
}

// satisfy repository.UserRepo
var _ repository.UserRepo = (*MockUserRepo)(nil)

func (m *MockUserRepo) GetUserByID(ctx context.Context, _ postgressqlx.Querier, _ string) (*entity.User, error) {
	return m.GetUserByIDResp, m.GetUserByIDErr
}

func (m *MockUserRepo) UpdateUserProfile(ctx context.Context, _ postgressqlx.Querier, user *entity.User) error {
	m.UpdateProfileCalled = true
	m.UpdateProfileArg = user
	return m.UpdateProfileErr
}

// ── postgressqlx.DB ──────────────────────────────────────────────────────────

// MockDB is a test double for postgressqlx.DB.
type MockDB struct {
	TxStarterErr error
}

// satisfy postgressqlx.DB (only the methods needed by the test suite)
var _ postgressqlx.DB = (*MockDB)(nil)

func (m *MockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (postgressqlx.TX, error) {
	if m.TxStarterErr != nil {
		return nil, m.TxStarterErr
	}
	return &MockTX{}, nil
}

func (m *MockDB) Close() error { return nil }

func (m *MockDB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (m *MockDB) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}

func (m *MockDB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}

func (m *MockDB) NamedQuery(query string, arg any) (*sqlx.Rows, error) { return nil, nil }

func (m *MockDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	return nil
}

func (m *MockDB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}

// ── postgressqlx.TX ──────────────────────────────────────────────────────────

// MockTX is a test double for postgressqlx.TX.  Query methods are no-ops
// because the command/query handlers under test do not read through the
// transaction — they only write via the repository.
type MockTX struct{}

// satisfy postgressqlx.TX
var _ postgressqlx.TX = (*MockTX)(nil)

func (m *MockTX) GetContext(ctx context.Context, dest any, query string, args ...any) error { return nil }
func (m *MockTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *MockTX) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}
func (m *MockTX) SelectContext(ctx context.Context, dest any, query string, args ...any) error { return nil }
func (m *MockTX) NamedQuery(query string, arg any) (*sqlx.Rows, error)                         { return nil, nil }
func (m *MockTX) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row      { return nil }
func (m *MockTX) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}
func (m *MockTX) Rollback() error { return nil }
func (m *MockTX) Commit() error   { return nil }

// ── test helpers ──────────────────────────────────────────────────────────────

// TestUser returns a populated entity.User for use in test cases.
func TestUser(id string) *entity.User {
	now := time.Now().UTC()
	return entity.UnmarshalUserFromDB(
		id, "cognito", "cognito-sub-123",
		"test@example.com", "Test User", "male",
		nil, valueobject.UserRoleClient,
		now, now, nil,
	)
}

// NotFoundErr returns an AppError with KindNotFound for the given code/message.
func NotFoundErr(code, msg string) *pkgerrors.AppError {
	return pkgerrors.NewAppError(pkgerrors.KindNotFound, code, msg)
}

// InternalErr returns an AppError with KindInternal for the given code/message.
func InternalErr(code, msg string) *pkgerrors.AppError {
	return pkgerrors.NewAppError(pkgerrors.KindInternal, code, msg)
}
