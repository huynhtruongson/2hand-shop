package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LukaGiorgadze/gonull/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// mockUpdateProductRequestRepository satisfies repository.ProductRequestRepository for testing.
type mockUpdateProductRequestRepository struct {
	mockProductRequestRepository
	updateErr    error
	updateCalled bool
}

func (m *mockUpdateProductRequestRepository) Update(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	m.updateCalled = true
	return m.updateErr
}

func TestUpdateProductRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	newPendingRequest := func(id, sellerID string) *aggregate.ProductRequest {
		return aggregate.UnmarshalProductRequestFromDB(
			id, sellerID, "cat-123",
			"Original Title", "Original Description",
			customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD, nil,
			valueobject.ConditionGood, valueobject.ProductRequestStatusPending,
			customtypes.Attachments{}, "contact@example.com",
			nil, nil, now, now, nil,
		)
	}
	approvedRequest := aggregate.UnmarshalProductRequestFromDB(
		"req-approved", "seller-001", "cat-123",
		"Approved Title", "Approved Description",
		customtypes.MustNewPrice("5000"), valueobject.CurrencyUSD, nil,
		valueobject.ConditionGood, valueobject.ProductRequestStatusApproved,
		customtypes.Attachments{}, "contact@example.com",
		nil, nil, now, now, nil,
	)

	tests := []struct {
		name       string
		cmd        UpdateProductRequestCommand
		prRepo     *mockUpdateProductRequestRepository
		wantResp   UpdateProductRequestResponse
		wantErr    error
		assertRepo func(t *testing.T, repo *mockUpdateProductRequestRepository, tx *mockTX)
	}{
		{
			name: "updates all provided fields",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("New Title"),
				Description:      gonull.NewNullable("New Description"),
				ExpectedPrice:    gonull.NewNullable(customtypes.MustNewPrice("7000")),
				Condition:        gonull.NewNullable("new"),
				Images:           gonull.NewNullable(customtypes.Attachments{{Key: "s3://bucket/img1.png"}}),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantResp: UpdateProductRequestResponse{},
			assertRepo: func(t *testing.T, repo *mockUpdateProductRequestRepository, tx *mockTX) {
				if repo.updateCalled == false {
					t.Error("expected Update to have been called")
				}
				if !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "updates only title when title field is present",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("Partial Update Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantResp: UpdateProductRequestResponse{},
			assertRepo: func(t *testing.T, repo *mockUpdateProductRequestRepository, tx *mockTX) {
				if repo.updateCalled == false {
					t.Error("expected Update to have been called")
				}
			},
		},
		{
			name: "returns ErrProductRequestNotFound when request does not exist",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "nonexistent",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("New Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: caterrors.ErrProductRequestNotFound,
				},
			},
			wantErr: caterrors.ErrProductRequestNotFound,
		},
		{
			name: "returns ErrInternal on unexpected GetByID error",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("New Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: errors.New("unexpected db error"),
				},
			},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrValidation when condition string is invalid",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
				Condition:        gonull.NewNullable("invalid_condition"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantErr: caterrors.ErrValidation,
		},
		{
			name: "returns ErrUnauthorized when seller does not own the request",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "wrong-seller",
				Title:            gonull.NewNullable("New Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
			},
			wantErr: caterrors.ErrUnauthorized,
		},
		{
			name: "returns ErrProductRequestNotEditable when request is already approved",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-approved",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("New Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: approvedRequest,
				},
			},
			wantErr: caterrors.ErrProductRequestNotEditable,
		},
		{
			name: "returns ErrInternal on unexpected Update error",
			cmd: UpdateProductRequestCommand{
				ProductRequestID: "req-1",
				SellerID:         "seller-001",
				Title:            gonull.NewNullable("New Title"),
			},
			prRepo: &mockUpdateProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest("req-1", "seller-001"),
				},
				updateErr: errors.New("unexpected db error"),
			},
			wantErr: caterrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			pub := &mockPublisher{}
			h := NewUpdateProductRequestHandler(tc.prRepo, db, pub)

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

func TestUpdateProductRequestResponse(t *testing.T) {
	t.Parallel()

	got := UpdateProductRequestResponse{}
	want := UpdateProductRequestResponse{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("UpdateProductRequestResponse mismatch (-want +got):\n%s", diff)
	}
}