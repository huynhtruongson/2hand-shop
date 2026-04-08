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

// mockRejectProductRequestRepository satisfies repository.ProductRequestRepository for testing.
type mockRejectProductRequestRepository struct {
	mockProductRequestRepository
	updateErr error
}

func (m *mockRejectProductRequestRepository) Update(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	return m.updateErr
}

func TestRejectProductRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
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
	approvedRequest := aggregate.UnmarshalProductRequestFromDB(
		"req-2", "seller-001", "cat-123",
		"Approved Title", "Approved Description",
		customtypes.MustNewPrice("3000"), valueobject.CurrencyUSD, nil,
		valueobject.ConditionLikeNew, valueobject.ProductRequestStatusApproved,
		customtypes.Attachments{}, "contact@example.com",
		nil, nil, now, now, nil,
	)
	rejectedRequest := aggregate.UnmarshalProductRequestFromDB(
		"req-3", "seller-001", "cat-123",
		"Rejected Title", "Rejected Description",
		customtypes.MustNewPrice("2000"), valueobject.CurrencyUSD, nil,
		valueobject.ConditionFair, valueobject.ProductRequestStatusRejected,
		customtypes.Attachments{}, "contact@example.com",
		nil, nil, now, now, nil,
	)

	tests := []struct {
		name       string
		cmd        RejectProductRequestCommand
		prRepo     *mockRejectProductRequestRepository
		wantResp   RejectProductRequestResponse
		wantErr    error
		assertRepo func(t *testing.T, prRepo *mockRejectProductRequestRepository, tx *mockTX)
	}{
		{
			name: "rejects pending request successfully",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "req-1",
				AdminRejectReason: "Item does not meet listing guidelines.",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest(),
				},
			},
			wantResp: RejectProductRequestResponse{},
			assertRepo: func(t *testing.T, prRepo *mockRejectProductRequestRepository, tx *mockTX) {
				req := prRepo.getByIDResult
				if req.Status() != valueobject.ProductRequestStatusRejected {
					t.Errorf("Status = %v, want %v", req.Status(), valueobject.ProductRequestStatusRejected)
				}
				if req.AdminRejectReason() == nil {
					t.Error("expected AdminRejectReason to be set")
				}
				if *req.AdminRejectReason() != "Item does not meet listing guidelines." {
					t.Errorf("AdminRejectReason = %q, want %q", *req.AdminRejectReason(), "Item does not meet listing guidelines.")
				}
				if !tx.commitCalled {
					t.Error("expected Commit to have been called")
				}
			},
		},
		{
			name: "returns ErrProductRequestNotFound when request does not exist",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "nonexistent",
				AdminRejectReason: "Some reason",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: caterrors.ErrProductRequestNotFound,
				},
			},
			wantErr: caterrors.ErrProductRequestNotFound,
		},
		{
			name: "returns ErrInternal on unexpected GetByID error",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "req-1",
				AdminRejectReason: "Some reason",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDErr: errors.New("unexpected db error"),
				},
			},
			wantErr: caterrors.ErrInternal,
		},
		{
			name: "returns ErrProductRequestNotEditable when request is already approved",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "req-2",
				AdminRejectReason: "Some reason",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: approvedRequest,
				},
			},
			wantErr: caterrors.ErrProductRequestNotEditable,
		},
		{
			name: "returns ErrProductRequestNotEditable when request is already rejected",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "req-3",
				AdminRejectReason: "Some reason",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: rejectedRequest,
				},
			},
			wantErr: caterrors.ErrProductRequestNotEditable,
		},
		{
			name: "returns ErrInternal on unexpected Update error",
			cmd: RejectProductRequestCommand{
				ProductRequestID:  "req-1",
				AdminRejectReason: "Some reason",
			},
			prRepo: &mockRejectProductRequestRepository{
				mockProductRequestRepository: mockProductRequestRepository{
					getByIDResult: newPendingRequest(),
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
			h := NewRejectProductRequestHandler(tc.prRepo, db, pub)

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

func TestRejectProductRequestResponse(t *testing.T) {
	t.Parallel()

	got := RejectProductRequestResponse{}
	want := RejectProductRequestResponse{}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("RejectProductRequestResponse mismatch (-want +got):\n%s", diff)
	}
}

// Ensure mockProductRequestRepository (embedded by mockRejectProductRequestRepository) implements repository.ProductRequestRepository.
var _ repository.ProductRequestRepository = (*mockProductRequestRepository)(nil)