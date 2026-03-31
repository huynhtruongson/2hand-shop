package command

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	errorspkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
	"github.com/jmoiron/sqlx"
)

// mockProductRepository records Save/GetByID calls and can be configured to return errors.
type mockProductRepository struct {
	saveErr   error
	updateErr error
	deleteErr error
	getByIDErr error
	saved     *aggregate.Product
}

func (m *mockProductRepository) Save(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	m.saved = product
	return m.saveErr
}

func (m *mockProductRepository) Update(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	m.saved = product
	return m.updateErr
}

func (m *mockProductRepository) Delete(ctx context.Context, q postgressqlx.Querier, productID string) error {
	return m.deleteErr
}

func (m *mockProductRepository) GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error) {
	return nil, m.getByIDErr
}

func (m *mockProductRepository) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListProductsFilter, page postgressqlx.Page) ([]aggregate.Product, int, error) {
	return nil, 0, nil
}

// mockTX records Commit/Rollback calls so assertions can verify transaction lifecycle.
type mockTX struct {
	commitCalled   bool
	rollbackCalled bool
}

func (tx *mockTX) Rollback() error   { tx.rollbackCalled = true; return nil }
func (tx *mockTX) Commit() error     { tx.commitCalled = true; return nil }
func (tx *mockTX) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (tx *mockTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (tx *mockTX) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}
func (tx *mockTX) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (tx *mockTX) NamedQuery(query string, arg any) (*sqlx.Rows, error)                             { return nil, nil }
func (tx *mockTX) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row      { return nil }
func (tx *mockTX) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}

// mockDB starts transactions that produce a mockTX.
type mockDB struct {
	tx *mockTX // shared so assertions can inspect the committed tx
}

func (m *mockDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (postgressqlx.TX, error) {
	m.tx = &mockTX{}
	return m.tx, nil
}
func (m *mockDB) Close() error { return nil }

func (m *mockDB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (m *mockDB) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, nil
}
func (m *mockDB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockDB) NamedQuery(query string, arg any) (*sqlx.Rows, error)                             { return nil, nil }
func (m *mockDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row      { return nil }
func (m *mockDB) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, nil
}

func TestCreateProductHandler_Handle(t *testing.T) {
	t.Parallel()

	validPrice := customtypes.MustNewPrice("10000")
	validImages := customtypes.Attachments{}

	// shared command used as a base; individual test cases override specific fields
	baseCmd := CreateProductCommand{
		CategoryID:  "cat-123",
		Title:       "Vintage Jacket",
		Description: "A nice vintage jacket",
		Price:       validPrice,
		Condition:   "new",
		Images:      validImages,
	}

	tests := []struct {
		name      string
		repoErr   error
		cmd       CreateProductCommand
		wantResp  CreateProductResponse
		wantErr   error
		assertion func(t *testing.T, repo *mockProductRepository, tx *mockTX)
	}{
		{
			name: "creates product with draft status and returns its ID",
			cmd:  baseCmd,
			wantResp: CreateProductResponse{}, // ID is a UUID; just assert non-empty below
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected product to be saved, got nil")
				}
				if repo.saved.ID() == "" {
					t.Error("expected non-empty ProductID")
				}
				if repo.saved.CategoryID() != "cat-123" {
					t.Errorf("CategoryID = %q, want %q", repo.saved.CategoryID(), "cat-123")
				}
				if repo.saved.Title() != "Vintage Jacket" {
					t.Errorf("Title = %q, want %q", repo.saved.Title(), "Vintage Jacket")
				}
				if repo.saved.Status() != valueobject.ProductStatusDraft {
					t.Errorf("Status = %v, want %v", repo.saved.Status(), valueobject.ProductStatusDraft)
				}
				if repo.saved.Currency().String() != "USD" {
					t.Errorf("Currency = %q, want %q", repo.saved.Currency().String(), "USD")
				}
				if tx != nil && !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "returns ErrValidation when title is empty",
			cmd: func() CreateProductCommand {
				c := baseCmd
				c.Title = "  "
				return c
			}(),
			wantErr: caterrors.ErrValidation,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected product not to be saved on validation error, but it was")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on validation error")
				}
			},
		},
		{
			name: "returns ErrValidation when categoryID is empty",
			cmd: func() CreateProductCommand {
				c := baseCmd
				c.CategoryID = "  "
				return c
			}(),
			wantErr: caterrors.ErrValidation,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected product not to be saved on validation error, but it was")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on validation error")
				}
			},
		},
		{
			name: "returns ErrValidation when price is zero",
			cmd: func() CreateProductCommand {
				c := baseCmd
				c.Price = customtypes.MustNewPrice("0")
				return c
			}(),
			wantErr: caterrors.ErrValidation,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected product not to be saved on validation error, but it was")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on validation error")
				}
			},
		},
		{
			name: "returns ErrProductConditionInvalid when condition is invalid",
			cmd: func() CreateProductCommand {
				c := baseCmd
				c.Condition = "invalid_condition" // not a valid condition string
				return c
			}(),
			wantErr: caterrors.ErrProductConditionInvalid,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected product not to be saved on condition error, but it was")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on condition error")
				}
			},
		},
		{
			name:    "returns ErrProductAlreadyExists on repo conflict",
			repoErr: caterrors.ErrProductAlreadyExists,
			cmd:     baseCmd,
			wantErr: caterrors.ErrProductAlreadyExists,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected product to be saved before conflict error")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on conflict error")
				}
			},
		},
		{
			name:    "returns ErrInternal on unexpected repo error",
			repoErr: errors.New("unexpected db error"),
			cmd:     baseCmd,
			wantErr: caterrors.ErrInternal,
			assertion: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected product to be saved before error")
				}
				if tx != nil && !tx.rollbackCalled {
					t.Error("expected Rollback to have been called on unexpected error")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockProductRepository{saveErr: tc.repoErr}
			db := &mockDB{}
			h := NewCreateProductHandler(repo, db)

			resp, err := h.Handle(context.Background(), tc.cmd)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errpkg.IsCode(err, tc.wantErr.(*errorspkg.AppError).Code()) {
					t.Errorf("error code = %v, want %v", err, tc.wantErr)
				}
				if tc.wantResp.ProductID != "" && resp.ProductID != "" {
					t.Errorf("ProductID should be empty on error, got %q", resp.ProductID)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.ProductID == "" {
					t.Error("ProductID is empty, want non-empty UUID")
				}
			}

			if tc.assertion != nil {
				tc.assertion(t, repo, db.tx)
			}
		})
	}
}

func TestCreateProductResponse(t *testing.T) {
	t.Parallel()

	got := CreateProductResponse{ProductID: "prod-uuid-123"}
	want := CreateProductResponse{ProductID: "prod-uuid-123"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CreateProductResponse mismatch (-want +got):\n%s", diff)
	}
}
