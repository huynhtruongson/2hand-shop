package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// mockGetProductRepository is a test double for repository.ProductRepository.
type mockGetProductRepository struct {
	getByIDResult *aggregate.Product
	getByIDErr    error
}

func (m *mockGetProductRepository) GetByID(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Product, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockGetProductRepository) Save(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Product) error {
	return errors.New("not implemented")
}
func (m *mockGetProductRepository) Update(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Product) error {
	return errors.New("not implemented")
}
func (m *mockGetProductRepository) Delete(_ context.Context, _ postgressqlx.Querier, _ string) error {
	return errors.New("not implemented")
}
func (m *mockGetProductRepository) List(_ context.Context, _ postgressqlx.Querier, _ repository.ListProductsFilter, _ postgressqlx.Page) ([]aggregate.Product, int, error) {
	return nil, 0, errors.New("not implemented")
}

func TestGetProductHandler_Handle(t *testing.T) {
	t.Parallel()

	existingProduct := aggregate.UnmarshalProductFromDB(
		"prod-1", "cat-1", "Test Product", "A test description",
		customtypes.MustNewPrice("9999"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusPublished,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	tests := []struct {
		name    string
		query   GetProductQuery
		repo    *mockGetProductRepository
		wantID  string
		wantErr bool
	}{
		{
			name:   "returns product when found",
			query:  GetProductQuery{ProductID: "prod-1"},
			repo:   &mockGetProductRepository{getByIDResult: existingProduct},
			wantID: "prod-1",
		},
		{
			name:    "returns ErrProductNotFound when product does not exist",
			query:   GetProductQuery{ProductID: "nonexistent"},
			repo:    &mockGetProductRepository{getByIDErr: caterrors.ErrProductNotFound},
			wantErr: true,
		},
		{
			name:    "propagates unexpected repo error",
			query:   GetProductQuery{ProductID: "prod-1"},
			repo:    &mockGetProductRepository{getByIDErr: errors.New("unexpected db error")},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := NewGetProductHandler(tc.repo, nil)
			got, err := h.Handle(context.Background(), tc.query)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("Handle() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("Handle() returned nil, want product with ID %s", tc.wantID)
			}
			if got.ID() != tc.wantID {
				t.Errorf("Handle() product.ID() = %s, want %s", got.ID(), tc.wantID)
			}
		})
	}
}
