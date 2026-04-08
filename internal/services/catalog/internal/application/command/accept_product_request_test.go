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
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// mockProductRequestRepository satisfies repository.ProductRequestRepository for testing.
type mockProductRequestRepository struct {
	getByIDResult       *aggregate.ProductRequest
	getByIDErr          error
	updateErr           error
	savedProductRequest *aggregate.ProductRequest
}

func (m *mockProductRequestRepository) GetByID(ctx context.Context, q postgressqlx.Querier, id string) (*aggregate.ProductRequest, error) {
	return m.getByIDResult, m.getByIDErr
}

func (m *mockProductRequestRepository) Update(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	return m.updateErr
}

func (m *mockProductRequestRepository) Save(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	m.savedProductRequest = pr
	return nil
}

func (m *mockProductRequestRepository) ListBySellerID(ctx context.Context, q postgressqlx.Querier, sellerID string) ([]*aggregate.ProductRequest, error) {
	return nil, nil
}

func (m *mockProductRequestRepository) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListProductRequestsFilter, page postgressqlx.Page) ([]*aggregate.ProductRequest, int, error) {
	return nil, 0, nil
}

func (m *mockProductRequestRepository) Delete(ctx context.Context, q postgressqlx.Querier, id string) error { return nil }

// mockAcceptProductRepository satisfies repository.ProductRepository for testing.
type mockAcceptProductRepository struct {
	mockProductRepository
	saveErr error
}

func (m *mockAcceptProductRepository) Save(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	m.saved = product
	return m.saveErr
}

func TestAcceptProductRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	// approvedRequest is never mutated — safe to share.
	approvedRequest := aggregate.UnmarshalProductRequestFromDB(
		"req-2", "seller-001", "cat-123",
		"Approved Title", "Approved Description",
		customtypes.MustNewPrice("3000"), valueobject.CurrencyUSD, nil,
		valueobject.ConditionLikeNew, valueobject.ProductRequestStatusApproved,
		customtypes.Attachments{}, "contact@example.com",
		nil, nil, now, now, nil,
	)

	// newPendingRequest returns a fresh pending ProductRequest so that each test case
	// gets an independent pointer that is not affected by mutations in other subtests.
	newPendingRequest := func() *aggregate.ProductRequest {
		return aggregate.UnmarshalProductRequestFromDB(
			"req-1", "seller-001", "cat-123",
			"Pending Title", "Pending Description",
			customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD, nil,
			valueobject.ConditionGood, valueobject.ProductRequestStatusPending,
			customtypes.Attachments{}, "contact@example.com",
			nil, nil, now, now, nil,
		)
	}

	tests := []struct {
		name       string
		cmd        AcceptProductRequestCommand
		prRepo     *mockProductRequestRepository
		prodRepo   *mockAcceptProductRepository
		cateRepo   *mockCategoryRepository
		wantResp   AcceptProductRequestResponse
		wantErr    error
		assertRepo func(t *testing.T, prRepo *mockProductRequestRepository, prodRepo *mockProductRepository, tx *mockTX)
	}{
		{
			name: "creates product from pending request and returns product ID",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-1"},
			prRepo: &mockProductRequestRepository{
				getByIDResult: newPendingRequest(),
			},
			prodRepo: &mockAcceptProductRepository{
				mockProductRepository: mockProductRepository{},
			},
			cateRepo: newMockCategoryRepository(),
			wantResp: AcceptProductRequestResponse{},
			assertRepo: func(t *testing.T, prRepo *mockProductRequestRepository, prodRepo *mockProductRepository, tx *mockTX) {
				if prodRepo.saved == nil {
					t.Fatal("expected product to be saved, got nil")
				}
				if prodRepo.saved.CategoryID() != "cat-123" {
					t.Errorf("CategoryID = %q, want %q", prodRepo.saved.CategoryID(), "cat-123")
				}
				if prodRepo.saved.Title() != "Pending Title" {
					t.Errorf("Title = %q, want %q", prodRepo.saved.Title(), "Pending Title")
				}
				if prodRepo.saved.Price().Cents() != 500000 {
					t.Errorf("Price.Cents() = %d, want %d", prodRepo.saved.Price().Cents(), 500000)
				}
				if !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "returns ErrProductRequestNotFound when request does not exist",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "nonexistent"},
			prRepo: &mockProductRequestRepository{
				getByIDErr: caterrors.ErrProductRequestNotFound,
			},
			wantErr: caterrors.ErrProductRequestNotFound,
			assertRepo: func(t *testing.T, prRepo *mockProductRequestRepository, prodRepo *mockProductRepository, tx *mockTX) {
				if prodRepo != nil && prodRepo.saved != nil {
					t.Error("expected product not to be saved when request not found")
				}
			},
		},
		{
			name: "returns ErrInternal on unexpected GetByID error",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-1"},
			prRepo: &mockProductRequestRepository{
				getByIDErr: errors.New("unexpected db error"),
			},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrProductRequestNotEditable when request is already approved",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-2"},
			prRepo: &mockProductRequestRepository{
				getByIDResult: approvedRequest,
			},
			wantErr: caterrors.ErrProductRequestNotEditable,
		},
		{
			name: "returns ErrInternal on unexpected Update error",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-1"},
			prRepo: &mockProductRequestRepository{
				getByIDResult: newPendingRequest(),
				updateErr:     errors.New("unexpected db error"),
			},
			prodRepo: &mockAcceptProductRepository{
				mockProductRepository: mockProductRepository{},
			},
			cateRepo: newMockCategoryRepository(),
			wantErr:  caterrors.ErrInternal,
		},
		{
			name: "returns ErrInternal on unexpected product Save error",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-1"},
			prRepo: &mockProductRequestRepository{
				getByIDResult: newPendingRequest(),
			},
			prodRepo: &mockAcceptProductRepository{
				mockProductRepository: mockProductRepository{},
				saveErr:             errors.New("unexpected db error"),
			},
			cateRepo: newMockCategoryRepository(),
			wantErr:  caterrors.ErrInternal,
		},
		{
			name: "returns ErrInternal on unexpected Category GetByID error",
			cmd:  AcceptProductRequestCommand{ProductRequestID: "req-1"},
			prRepo: &mockProductRequestRepository{
				getByIDResult: newPendingRequest(),
			},
			prodRepo: &mockAcceptProductRepository{
				mockProductRepository: mockProductRepository{},
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
			if tc.prodRepo == nil {
				tc.prodRepo = &mockAcceptProductRepository{mockProductRepository: mockProductRepository{}}
			}
			cateRepo := tc.cateRepo
			if cateRepo == nil {
				cateRepo = newMockCategoryRepository()
			}
			h := NewAcceptProductRequestHandler(tc.prRepo, tc.prodRepo, cateRepo, db, pub)

			resp, err := h.Handle(context.Background(), tc.cmd)

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
				if resp.ProductID == "" {
					t.Error("ProductID is empty, want non-empty")
				}
			}

			if tc.assertRepo != nil {
				tc.assertRepo(t, tc.prRepo, &tc.prodRepo.mockProductRepository, db.tx)
			}
		})
	}
}

func TestAcceptProductRequestResponse(t *testing.T) {
	t.Parallel()

	got := AcceptProductRequestResponse{ProductID: "prod-123"}
	want := AcceptProductRequestResponse{ProductID: "prod-123"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("AcceptProductRequestResponse mismatch (-want +got):\n%s", diff)
	}
}

var _ repository.ProductRequestRepository = (*mockProductRequestRepository)(nil)
