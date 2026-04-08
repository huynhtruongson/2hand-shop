package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// mockCreateProductRequestRepository satisfies repository.ProductRequestRepository for testing.
type mockCreateProductRequestRepository struct {
	mockProductRequestRepository
	saveErr error
}

func (m *mockCreateProductRequestRepository) Save(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	m.savedProductRequest = pr
	return m.saveErr
}

func TestCreateProductRequestHandler_Handle(t *testing.T) {
	t.Parallel()

	baseCmd := CreateProductRequestCommand{
		SellerID:      "seller-001",
		CategoryID:    "cat-123",
		Title:         "Vintage Jacket",
		Description:   "A nice vintage jacket",
		Brand:         nil,
		ExpectedPrice: customtypes.MustNewPrice("5000"),
		Condition:     "good",
		Images:        customtypes.Attachments{},
		ContactInfo:   "contact@example.com",
		AdminNote:     nil,
	}

	tests := []struct {
		name      string
		cmd       CreateProductRequestCommand
		prRepo    *mockCreateProductRequestRepository
		wantResp  CreateProductRequestResponse
		wantErr   error
		assertion func(t *testing.T, repo *mockCreateProductRequestRepository, tx *mockTX)
	}{
		{
			name:   "creates pending product request and returns its ID",
			cmd:    baseCmd,
			prRepo: &mockCreateProductRequestRepository{},
			wantResp: CreateProductRequestResponse{},
			assertion: func(t *testing.T, repo *mockCreateProductRequestRepository, tx *mockTX) {
				if repo.savedProductRequest == nil {
					t.Fatal("expected product request to be saved, got nil")
				}
				if repo.savedProductRequest.ID() == "" {
					t.Error("expected non-empty ProductRequestID")
				}
				if repo.savedProductRequest.SellerID() != "seller-001" {
					t.Errorf("SellerID = %q, want %q", repo.savedProductRequest.SellerID(), "seller-001")
				}
				if repo.savedProductRequest.CategoryID() != "cat-123" {
					t.Errorf("CategoryID = %q, want %q", repo.savedProductRequest.CategoryID(), "cat-123")
				}
				if repo.savedProductRequest.Title() != "Vintage Jacket" {
					t.Errorf("Title = %q, want %q", repo.savedProductRequest.Title(), "Vintage Jacket")
				}
				if repo.savedProductRequest.Status() != valueobject.ProductRequestStatusPending {
					t.Errorf("Status = %v, want %v", repo.savedProductRequest.Status(), valueobject.ProductRequestStatusPending)
				}
			},
		},
		{
			name: "returns ErrValidation when title is empty",
			cmd: func() CreateProductRequestCommand {
				c := baseCmd
				c.Title = "  "
				return c
			}(),
			prRepo:  &mockCreateProductRequestRepository{},
			wantErr: caterrors.ErrValidation,
		},
		{
			name: "returns ErrValidation when categoryID is empty",
			cmd: func() CreateProductRequestCommand {
				c := baseCmd
				c.CategoryID = "  "
				return c
			}(),
			prRepo:  &mockCreateProductRequestRepository{},
			wantErr: caterrors.ErrValidation,
		},
		{
			name: "returns ErrValidation when expected price is zero",
			cmd: func() CreateProductRequestCommand {
				c := baseCmd
				c.ExpectedPrice = customtypes.MustNewPrice("0")
				return c
			}(),
			prRepo:  &mockCreateProductRequestRepository{},
			wantErr: caterrors.ErrValidation,
		},
		{
			name: "returns ErrProductConditionInvalid when condition is invalid",
			cmd: func() CreateProductRequestCommand {
				c := baseCmd
				c.Condition = "invalid_condition"
				return c
			}(),
			prRepo:  &mockCreateProductRequestRepository{},
			wantErr: caterrors.ErrProductConditionInvalid,
		},
		{
			name:   "returns ErrProductRequestAlreadyExists on repo conflict",
			cmd:    baseCmd,
			prRepo: &mockCreateProductRequestRepository{saveErr: caterrors.ErrProductRequestAlreadyExists},
			wantErr: caterrors.ErrProductRequestAlreadyExists,
		},
		{
			name: "returns ErrInternal on unexpected repo error",
			cmd:    baseCmd,
			prRepo: &mockCreateProductRequestRepository{saveErr: errors.New("unexpected db error")},
			wantErr: caterrors.ErrInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &mockDB{}
			pub := &mockPublisher{}
			h := NewCreateProductRequestHandler(tc.prRepo, db, pub)

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
				if resp.ProductRequestID == "" {
					t.Error("ProductRequestID is empty, want non-empty UUID")
				}
			}

			if tc.assertion != nil {
				tc.assertion(t, tc.prRepo, db.tx)
			}
		})
	}
}

func TestCreateProductRequestResponse(t *testing.T) {
	t.Parallel()

	got := CreateProductRequestResponse{ProductRequestID: "req-uuid-123"}
	want := CreateProductRequestResponse{ProductRequestID: "req-uuid-123"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CreateProductRequestResponse mismatch (-want +got):\n%s", diff)
	}
}
