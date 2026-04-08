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
)

// mockDeletePRRepo satisfies repository.ProductRequestRepository for testing.
type mockDeletePRRepo struct {
	mockProductRequestRepository
	deleteErr error
}

func (m *mockDeletePRRepo) Delete(ctx context.Context, q postgressqlx.Querier, id string) error {
	return m.deleteErr
}

func TestDeleteProductRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	newPendingRequest := func(id, sellerID string) *aggregate.ProductRequest {
		return aggregate.UnmarshalProductRequestFromDB(
			id, sellerID, "cat-123",
			"Pending Title", "Pending Description",
			customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD, nil,
			valueobject.ConditionGood, valueobject.ProductRequestStatusPending,
			customtypes.Attachments{}, "contact@example.com",
			nil, nil, now, now, nil,
		)
	}

	tests := []struct {
		name       string
		cmd        DeleteProductRequestCommand
		prRepo     *mockDeletePRRepo
		wantResp   DeleteProductRequestResponse
		wantErr    error
		assertRepo func(t *testing.T, repo *mockDeletePRRepo, tx *mockTX)
	}{
		{
			name: "soft-deletes pending product request as the owning seller",
			cmd: DeleteProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
			},
			prRepo: &mockDeletePRRepo{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantResp: DeleteProductRequestResponse{},
			assertRepo: func(t *testing.T, repo *mockDeletePRRepo, tx *mockTX) {
				if tx.commitCalled == false {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "returns ErrProductRequestNotFound when request does not exist",
			cmd: DeleteProductRequestCommand{
				ProductRequestID: "nonexistent",
				SellerID:         "seller-001",
			},
			prRepo: &mockDeletePRRepo{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: caterrors.ErrProductRequestNotFound,
				},
			},
			wantErr: caterrors.ErrProductRequestNotFound,
		},
		{
			name: "returns ErrInternal on unexpected GetByID error",
			cmd: DeleteProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
			},
			prRepo: &mockDeletePRRepo{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: errors.New("unexpected db error"),
				},
			},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrUnauthorized when seller does not own the request",
			cmd: DeleteProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "wrong-seller",
			},
			prRepo: &mockDeletePRRepo{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantErr: caterrors.ErrUnauthorized,
		},
		{
			name: "returns ErrInternal on unexpected Delete error",
			cmd: DeleteProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
			},
			prRepo: &mockDeletePRRepo{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
				deleteErr: errors.New("unexpected db error"),
			},
			wantErr: caterrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			pub := &mockPublisher{}
			h := NewDeleteProductRequestHandler(tc.prRepo, db, pub)

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
				if diff := cmp.Diff(tc.wantResp, resp); diff != "" {
					t.Errorf("response mismatch (-want +got):\n%s", diff)
				}
			}

			if tc.assertRepo != nil {
				tc.assertRepo(t, tc.prRepo, db.tx)
			}
		})
	}
}

func TestDeleteProductRequestResponse(t *testing.T) {
	t.Parallel()

	got := DeleteProductRequestResponse{}
	want := DeleteProductRequestResponse{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DeleteProductRequestResponse mismatch (-want +got):\n%s", diff)
	}
}