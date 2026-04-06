package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// mockDeleteProductRepository extends mockProductRepository with get-by-ID controls.
type mockDeleteProductRepository struct {
	mockProductRepository
	getByIDResult *aggregate.Product
	getByIDErr    error
}

func (m *mockDeleteProductRepository) GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error) {
	return m.getByIDResult, m.getByIDErr
}

func TestDeleteProductHandler_Handle(t *testing.T) {
	t.Parallel()

	existingProduct := aggregate.UnmarshalProductFromDB(
		"prod-1", "cat-1", "Old Title", "Old Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD,
		valueobject.ConditionGood, valueobject.ProductStatusPublished,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	tests := []struct {
		name    string
		cmd     DeleteProductCommand
		repo    *mockDeleteProductRepository
		wantErr error
	}{
		{
			name: "soft-deletes product successfully",
			cmd:  DeleteProductCommand{ProductID: "prod-1"},
			repo: &mockDeleteProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingProduct,
			},
		},
		{
			name:    "returns ErrProductNotFound when product does not exist",
			cmd:     DeleteProductCommand{ProductID: "nonexistent"},
			repo:    &mockDeleteProductRepository{getByIDErr: caterrors.ErrProductNotFound},
			wantErr: caterrors.ErrProductNotFound,
		},
		{
			name:    "returns ErrInternal on unexpected GetByID error",
			cmd:     DeleteProductCommand{ProductID: "prod-1"},
			repo:    &mockDeleteProductRepository{getByIDErr: errors.New("unexpected db error")},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrInternal on unexpected Delete error",
			cmd:  DeleteProductCommand{ProductID: "prod-1"},
			repo: &mockDeleteProductRepository{
				mockProductRepository: mockProductRepository{deleteErr: errors.New("unexpected db error")},
				getByIDResult:         existingProduct,
			},
			wantErr: caterrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			pub := &mockPublisher{}
			h := NewDeleteProductHandler(tc.repo, db, pub)

			_, err := h.Handle(context.Background(), tc.cmd)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errpkg.IsCode(err, tc.wantErr.(*errpkg.AppError).Code()) {
					t.Errorf("error code = %v, want %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDeleteProductResponse(t *testing.T) {
	t.Parallel()

	got := DeleteProductResponse{}
	want := DeleteProductResponse{}
	if got != want {
		t.Errorf("DeleteProductResponse = %+v, want %+v", got, want)
	}
}
