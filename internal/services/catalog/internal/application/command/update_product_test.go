package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"

	"github.com/LukaGiorgadze/gonull/v2"
)

// mockUpdateProductRepository extends mockProductRepository with per-call get/update error controls.
type mockUpdateProductRepository struct {
	mockProductRepository
	getByIDResult *aggregate.Product
	getByIDErr    error
	updateErr     error
}

func (m *mockUpdateProductRepository) GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockUpdateProductRepository) Update(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return m.mockProductRepository.Update(ctx, q, product)
}

func TestUpdateProductHandler_Handle(t *testing.T) {
	t.Parallel()

	existingProduct := aggregate.UnmarshalProductFromDB(
		"prod-1", "seller-001",
		"Old Title", "Old Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD,
		valueobject.ConditionGood, valueobject.ProductStatusPublished,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	existingDraftProduct := aggregate.UnmarshalProductFromDB(
		"prod-2", "seller-001",
		"Draft Title", "Draft Description",
		customtypes.MustNewPrice("3000"), valueobject.CurrencyEUR,
		valueobject.ConditionLikeNew, valueobject.ProductStatusDraft,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	existingSoldProduct := aggregate.UnmarshalProductFromDB(
		"prod-3", "seller-001",
		"Sold Title", "Sold Description",
		customtypes.MustNewPrice("1000"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusSold,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	existingArchivedProduct := aggregate.UnmarshalProductFromDB(
		"prod-4", "seller-001",
		"Archived Title", "Archived Description",
		customtypes.MustNewPrice("2000"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusArchived,
		customtypes.Attachments{}, nil, time.Now().UTC(), time.Now().UTC(), nil,
	)

	newTitle := "New Title"
	newPrice := customtypes.MustNewPrice("9999")
	newCondition := "new"
	newImages := customtypes.Attachments{{Key: "s3://bucket/img1.png"}}

	tests := []struct {
		name       string
		cmd        UpdateProductCommand
		repo       *mockUpdateProductRepository
		wantResp   UpdateProductResponse
		wantErr    error
		assertRepo func(t *testing.T, repo *mockProductRepository, tx *mockTX)
	}{
		{
			name: "updates title only",
			cmd: UpdateProductCommand{
				ProductID: "prod-1",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingProduct,
			},
			wantResp: UpdateProductResponse{},
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected product to be updated, got nil")
				}
				if repo.saved.Title() != "New Title" {
					t.Errorf("Title = %q, want %q", repo.saved.Title(), "New Title")
				}
				if !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "updates all fields",
			cmd: UpdateProductCommand{
				ProductID:   "prod-1",
				Title:       gonull.NewNullable(newTitle),
				Description: gonull.NewNullable(newTitle),
				Price:       gonull.NewNullable(newPrice),
				Condition:   gonull.NewNullable(newCondition),
				Images:      gonull.NewNullable(newImages),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingProduct,
			},
			wantResp: UpdateProductResponse{},
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected product to be updated, got nil")
				}
				if repo.saved.Title() != "New Title" {
					t.Errorf("Title = %q, want %q", repo.saved.Title(), "New Title")
				}
				if repo.saved.Price().Cents() != 999900 {
					t.Errorf("Price.Cents() = %d, want %d", repo.saved.Price().Cents(), 999900)
				}
				if repo.saved.Currency().String() != "USD" {
					t.Errorf("Currency = %q, want %q", repo.saved.Currency().String(), "USD")
				}
				if repo.saved.Condition() != valueobject.ConditionNew {
					t.Errorf("Condition = %v, want %v", repo.saved.Condition(), valueobject.ConditionNew)
				}
				if len(repo.saved.Images()) != 1 {
					t.Errorf("Images len = %d, want 1", len(repo.saved.Images()))
				}
				if !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "updates draft product",
			cmd: UpdateProductCommand{
				ProductID: "prod-2",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingDraftProduct,
			},
			wantResp: UpdateProductResponse{},
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved == nil {
					t.Fatal("expected draft product to be updated, got nil")
				}
				if repo.saved.Title() != "New Title" {
					t.Errorf("Title = %q, want %q", repo.saved.Title(), "New Title")
				}
			},
		},
		{
			name: "returns ErrProductNotFound when product does not exist",
			cmd: UpdateProductCommand{
				ProductID: "nonexistent",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDErr:            caterrors.ErrProductNotFound,
			},
			wantErr: caterrors.ErrProductNotFound,
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected product not to be saved when product not found")
				}
				if tx != nil && tx.commitCalled {
					t.Error("expected Commit NOT to be called when product not found")
				}
			},
		},
		{
			name: "returns ErrInternal on unexpected GetByID error",
			cmd: UpdateProductCommand{
				ProductID: "prod-1",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDErr:            errors.New("unexpected db error"),
			},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrInternal on unexpected Update error",
			cmd: UpdateProductCommand{
				ProductID: "prod-1",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingProduct,
				updateErr:             errors.New("unexpected db error"),
			},
			wantErr: caterrors.ErrInternal,
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if tx != nil && tx.commitCalled {
					t.Error("expected Commit NOT to be called on Update error")
				}
			},
		},
		{
			name: "returns ErrProductConditionInvalid when condition string is invalid",
			cmd: UpdateProductCommand{
				ProductID: "prod-1",
				Condition: gonull.NewNullable("invalid-condition"),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingProduct,
			},
			wantErr: caterrors.ErrProductConditionInvalid,
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				// product.Update is never called because condition parsing fails first
				if repo.saved != nil {
					t.Error("expected product not to be updated when condition string is invalid")
				}
				if tx != nil && tx.commitCalled {
					t.Error("expected Commit NOT to be called on validation error")
				}
			},
		},
		{
			name: "returns ErrProductInvalidStatusTransition when updating sold product",
			cmd: UpdateProductCommand{
				ProductID: "prod-3",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingSoldProduct,
			},
			wantErr: caterrors.ErrProductInvalidStatusTransition,
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected sold product not to be updated, but it was")
				}
				if tx != nil && tx.commitCalled {
					t.Error("expected Commit NOT to be called when product is sold")
				}
			},
		},
		{
			name: "returns ErrProductInvalidStatusTransition when updating archived product",
			cmd: UpdateProductCommand{
				ProductID: "prod-4",
				Title:     gonull.NewNullable(newTitle),
			},
			repo: &mockUpdateProductRepository{
				mockProductRepository: mockProductRepository{},
				getByIDResult:         existingArchivedProduct,
			},
			wantErr: caterrors.ErrProductInvalidStatusTransition,
			assertRepo: func(t *testing.T, repo *mockProductRepository, tx *mockTX) {
				if repo.saved != nil {
					t.Error("expected archived product not to be updated, but it was")
				}
				if tx != nil && tx.commitCalled {
					t.Error("expected Commit NOT to be called when product is archived")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			cateRepo := newMockCategoryRepository()
			pub := &mockPublisher{}
			h := NewUpdateProductHandler(tc.repo, cateRepo, db, pub)

			resp, err := h.Handle(context.Background(), tc.cmd)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errpkg.IsCode(err, tc.wantErr.(interface{ Code() string }).Code()) {
					t.Errorf("error code = %v, want %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(tc.wantResp, resp); diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			}

			if tc.assertRepo != nil {
				tc.assertRepo(t, &tc.repo.mockProductRepository, db.tx)
			}
		})
	}
}

func TestUpdateProductResponse(t *testing.T) {
	t.Parallel()

	got := UpdateProductResponse{}
	want := UpdateProductResponse{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UpdateProductResponse mismatch (-want +got):\n%s", diff)
	}
}
