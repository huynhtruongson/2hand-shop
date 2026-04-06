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

// mockPublishProductRepository extends mockProductRepository with GetByID and Update controls.
type mockPublishProductRepository struct {
	mockProductRepository
	getByIDResult *aggregate.Product
	getByIDErr    error
	updateErr     error
}

func (m *mockPublishProductRepository) GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.getByIDResult == nil {
		return nil, nil
	}
	// Return a fresh copy so that each Handle invocation sees an independent product
	// (product.Publish() mutates in-place; repeated GetByID calls must not share the mutated copy).
	return aggregate.UnmarshalProductFromDB(
		m.getByIDResult.ID(), m.getByIDResult.CategoryID(),
		m.getByIDResult.Title(), m.getByIDResult.Description(),
		m.getByIDResult.Price(), m.getByIDResult.Currency(),
		m.getByIDResult.Condition(), m.getByIDResult.Status(),
		m.getByIDResult.Images(), m.getByIDResult.Brand(),
		m.getByIDResult.CreatedAt(), m.getByIDResult.UpdatedAt(),
		m.getByIDResult.DeletedAt(),
	), nil
}

func (m *mockPublishProductRepository) Update(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	return m.updateErr
}

func TestPublishProductHandler_Handle(t *testing.T) {
	t.Parallel()

	var deletedAt *time.Time
	draftProduct := aggregate.UnmarshalProductFromDB(
		"prod-draft", "cat-1", "Draft Title", "Draft Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD,
		valueobject.ConditionGood, valueobject.ProductStatusDraft,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), deletedAt,
	)
	publishedProduct := aggregate.UnmarshalProductFromDB(
		"prod-published", "cat-1", "Published Title", "Published Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD,
		valueobject.ConditionGood, valueobject.ProductStatusPublished,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), deletedAt,
	)
	soldProduct := aggregate.UnmarshalProductFromDB(
		"prod-sold", "cat-1", "Sold Title", "Sold Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD,
		valueobject.ConditionGood, valueobject.ProductStatusSold,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), deletedAt,
	)

	tests := []struct {
		name    string
		cmd     PublishProductCommand
		repo    *mockPublishProductRepository
		cateRepo *mockCategoryRepository
		wantErr error
	}{
		{
			name: "publishes draft product successfully",
			cmd:  PublishProductCommand{ProductID: "prod-draft"},
			repo: &mockPublishProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:        draftProduct,
			},
			cateRepo: newMockCategoryRepository(),
		},
		{
			name:    "returns ErrProductNotFound when product does not exist",
			cmd:     PublishProductCommand{ProductID: "nonexistent"},
			repo:    &mockPublishProductRepository{getByIDErr: caterrors.ErrProductNotFound},
			wantErr: caterrors.ErrProductNotFound,
		},
		{
			name:    "returns ErrInternal on unexpected GetByID error",
			cmd:     PublishProductCommand{ProductID: "prod-1"},
			repo:    &mockPublishProductRepository{getByIDErr: errors.New("unexpected db error")},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrProductInvalidStatusTransition when product is already published",
			cmd:  PublishProductCommand{ProductID: "prod-published"},
			repo: &mockPublishProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:        publishedProduct,
			},
			wantErr: caterrors.ErrProductInvalidStatusTransition,
		},
		{
			name: "returns ErrProductInvalidStatusTransition when product is sold",
			cmd:  PublishProductCommand{ProductID: "prod-sold"},
			repo: &mockPublishProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:        soldProduct,
			},
			wantErr: caterrors.ErrProductInvalidStatusTransition,
		},
		{
			name: "returns ErrInternal on unexpected Update error",
			cmd:  PublishProductCommand{ProductID: "prod-draft"},
			repo: &mockPublishProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:        draftProduct,
				updateErr:            errors.New("unexpected db error"),
			},
			cateRepo: newMockCategoryRepository(),
			wantErr:  caterrors.ErrInternal,
		},
		{
			name: "returns ErrInternal on unexpected Category GetByID error",
			cmd:  PublishProductCommand{ProductID: "prod-draft"},
			repo: &mockPublishProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:        draftProduct,
			},
			cateRepo: &mockCategoryRepository{getByIDErr: errors.New("unexpected db error")},
			wantErr:  caterrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			pub := &mockPublisher{}
			cateRepo := tc.cateRepo
			if cateRepo == nil {
				cateRepo = newMockCategoryRepository()
			}
			h := NewPublishProductHandler(tc.repo, cateRepo, db, pub)

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

func TestPublishProductResponse(t *testing.T) {
	t.Parallel()

	got := PublishProductResponse{}
	want := PublishProductResponse{}
	if got != want {
		t.Errorf("PublishProductResponse = %+v, want %+v", got, want)
	}
}
