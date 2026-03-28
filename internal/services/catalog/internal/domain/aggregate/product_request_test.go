package aggregate

import (
	"errors"
	"testing"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
	apperrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
)

// ── Test helpers ─────────────────────────────────────────────────────────────

func makeValidNewProductRequestParams() (
	id, sellerID, categoryID, title, description string,
	expectedPrice customtypes.Price, currency customtypes.Currency,
	condition valueobject.Condition, images customtypes.Attachments, contactInfo string,
) {
	return "req-001", "seller-001", "cat-001", "Vintage Lamp", "A beautiful vintage desk lamp",
		customtypes.MustNewPrice("30.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionGood, makeAttachments(), "seller@example.com"
}

func rebuildProductRequest(
	id, sellerID, categoryID, title, description string,
	expectedPrice customtypes.Price, currency customtypes.Currency,
	condition valueobject.Condition, status valueobject.ProductRequestStatus,
	images customtypes.Attachments, contactInfo, adminRejectReason, adminNote string,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *ProductRequest {
	return UnmarshalProductRequestFromDB(
		id, sellerID, categoryID, title, description,
		expectedPrice, currency, condition, status,
		images, contactInfo, adminRejectReason, adminNote,
		createdAt, updatedAt, deletedAt,
	)
}

// ── NewProductRequest ─────────────────────────────────────────────────────────

func TestNewProductRequest_Valid(t *testing.T) {
	t.Parallel()
	id, sellerID, categoryID, title, description, expPrice, currency, condition, images, contact := makeValidNewProductRequestParams()

	pr, err := NewProductRequest(id, sellerID, categoryID, title, description,
		expPrice, currency, condition, images, contact)

	if err != nil {
		t.Fatalf("NewProductRequest() unexpected error: %v", err)
	}
	if pr.ID() != id {
		t.Errorf("ID() = %q, want %q", pr.ID(), id)
	}
	if pr.SellerID() != sellerID {
		t.Errorf("SellerID() = %q, want %q", pr.SellerID(), sellerID)
	}
	if pr.CategoryID() != categoryID {
		t.Errorf("CategoryID() = %q, want %q", pr.CategoryID(), categoryID)
	}
	if pr.Title() != title {
		t.Errorf("Title() = %q, want %q", pr.Title(), title)
	}
	if pr.Description() != description {
		t.Errorf("Description() = %q, want %q", pr.Description(), description)
	}
	if !pr.ExpectedPrice().Equal(expPrice) {
		t.Errorf("ExpectedPrice() = %v, want %v", pr.ExpectedPrice(), expPrice)
	}
	if pr.Currency() != currency {
		t.Errorf("Currency() = %v, want %v", pr.Currency(), currency)
	}
	if pr.Condition() != condition {
		t.Errorf("Condition() = %v, want %v", pr.Condition(), condition)
	}
	if pr.Status() != valueobject.ProductRequestStatusPending {
		t.Errorf("Status() = %v, want %v", pr.Status(), valueobject.ProductRequestStatusPending)
	}
	if pr.ContactInfo() != contact {
		t.Errorf("ContactInfo() = %q, want %q", pr.ContactInfo(), contact)
	}
	if pr.AdminRejectReason() != "" {
		t.Errorf("AdminRejectReason() = %q, want empty string", pr.AdminRejectReason())
	}
	if pr.AdminNote() != "" {
		t.Errorf("AdminNote() = %q, want empty string", pr.AdminNote())
	}
	if pr.DeletedAt() != nil {
		t.Error("DeletedAt() = not nil, want nil for a new product request")
	}
}

func TestNewProductRequest_ValidationErrors(t *testing.T) {
	t.Parallel()
	id, sellerID, categoryID, title, description, expPrice, currency, condition, images, contact := makeValidNewProductRequestParams()

	tests := []struct {
		name    string
		modify  func()
		wantErr string // ae.Details map key; empty = no error expected
	}{
		{
			name:    "empty id",
			modify:  func() { id = "  " },
			wantErr: "id",
		},
		{
			name:    "empty seller_id",
			modify:  func() { sellerID = "  " },
			wantErr: "seller_id",
		},
		{
			name:    "empty category_id",
			modify:  func() { categoryID = "  " },
			wantErr: "category_id",
		},
		{
			name:    "empty title",
			modify:  func() { title = "  " },
			wantErr: "title",
		},
		{
			name:    "zero expected_price",
			modify:  func() { expPrice = customtypes.MustNewPrice("0") },
			wantErr: "expected_price",
		},
		{
			name: "negative expected_price",
			modify: func() {
				neg, _ := customtypes.NewPrice("-5.00")
				expPrice = neg
			},
			wantErr: "expected_price",
		},
		{
			name:    "empty currency",
			modify:  func() { currency = "" },
			wantErr: "currency",
		},
		{
			name: "invalid condition",
			modify: func() {
				bad, _ := valueobject.NewConditionFromString("invalid")
				condition = bad
			},
			wantErr: "condition",
		},
		{
			name:    "images may be nil",
			modify:  func() {},
			wantErr: "",
		},
		{
			name:    "contact_info may be empty",
			modify:  func() {},
			wantErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Reset to defaults before each iteration.
			id, sellerID, categoryID, title, description, expPrice, currency, condition, images, contact =
				"req-001", "seller-001", "cat-001", "Vintage Lamp", "A beautiful vintage desk lamp",
				customtypes.MustNewPrice("30.00"), customtypes.MustCurrency("USD"),
				valueobject.ConditionGood, makeAttachments(), "seller@example.com"

			tc.modify()

			_, err := NewProductRequest(id, sellerID, categoryID, title, description,
				expPrice, currency, condition, images, contact)

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var ae *apperrors.AppError
			if !errors.As(err, &ae) {
				t.Fatalf("error does not implement *AppError: %T", err)
			}
			ue := ae.UserFacing()
			if _, ok := ue.Details[tc.wantErr]; !ok {
				t.Errorf("UserFacing().Details[%q] missing; full details = %v", tc.wantErr, ue.Details)
			}
		})
	}
}

// ── Getters ───────────────────────────────────────────────────────────────────

func TestProductRequest_Getters(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	pr := rebuildProductRequest(
		"req-getter", "seller-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("50.00"), customtypes.MustCurrency("EUR"),
		valueobject.ConditionFair, valueobject.ProductRequestStatusApproved,
		makeAttachments(), "email@test.com", "not good enough", "internal note",
		now, now, nil,
	)

	if got := pr.ID(); got != "req-getter" {
		t.Errorf("ID() = %q, want %q", got, "req-getter")
	}
	if got := pr.SellerID(); got != "seller-001" {
		t.Errorf("SellerID() = %q, want %q", got, "seller-001")
	}
	if got := pr.CategoryID(); got != "cat-001" {
		t.Errorf("CategoryID() = %q, want %q", got, "cat-001")
	}
	if got := pr.Title(); got != "Title" {
		t.Errorf("Title() = %q, want %q", got, "Title")
	}
	if got := pr.Description(); got != "Desc" {
		t.Errorf("Description() = %q, want %q", got, "Desc")
	}
	if got := pr.Status(); got != valueobject.ProductRequestStatusApproved {
		t.Errorf("Status() = %v, want %v", got, valueobject.ProductRequestStatusApproved)
	}
	if got := pr.ContactInfo(); got != "email@test.com" {
		t.Errorf("ContactInfo() = %q, want %q", got, "email@test.com")
	}
	if got := pr.AdminRejectReason(); got != "not good enough" {
		t.Errorf("AdminRejectReason() = %q, want %q", got, "not good enough")
	}
	if got := pr.AdminNote(); got != "internal note" {
		t.Errorf("AdminNote() = %q, want %q", got, "internal note")
	}
	if got := pr.DeletedAt(); got != nil {
		t.Errorf("DeletedAt() = %v, want nil", got)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestProductRequest_Update(t *testing.T) {
	t.Parallel()
	pr := rebuildProductRequest(
		"req-upd", "seller-001", "cat-001", "Old Title", "Old Desc",
		customtypes.MustNewPrice("10.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionFair, valueobject.ProductRequestStatusPending,
		nil, "old@email.com", "", "",
		time.Now().UTC(), time.Now().UTC(), nil,
	)

	newImages := makeAttachments()
	err := pr.Update(
		"New Title", "New Desc", "cat-002",
		customtypes.MustNewPrice("20.00"), customtypes.MustCurrency("EUR"),
		valueobject.ConditionLikeNew, newImages, "new@email.com",
	)

	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if pr.Title() != "New Title" {
		t.Errorf("Title() = %q, want %q", pr.Title(), "New Title")
	}
	if pr.Description() != "New Desc" {
		t.Errorf("Description() = %q, want %q", pr.Description(), "New Desc")
	}
	if pr.CategoryID() != "cat-002" {
		t.Errorf("CategoryID() = %q, want %q", pr.CategoryID(), "cat-002")
	}
	if !pr.ExpectedPrice().Equal(customtypes.MustNewPrice("20.00")) {
		t.Errorf("ExpectedPrice() = %v, want %v", pr.ExpectedPrice(), customtypes.MustNewPrice("20.00"))
	}
	if pr.Currency() != customtypes.MustCurrency("EUR") {
		t.Errorf("Currency() = %v, want %v", pr.Currency(), customtypes.MustCurrency("EUR"))
	}
	if pr.Condition() != valueobject.ConditionLikeNew {
		t.Errorf("Condition() = %v, want %v", pr.Condition(), valueobject.ConditionLikeNew)
	}
	if pr.ContactInfo() != "new@email.com" {
		t.Errorf("ContactInfo() = %q, want %q", pr.ContactInfo(), "new@email.com")
	}
}

func TestProductRequest_Update_NotEditableWhenApproved(t *testing.T) {
	t.Parallel()
	pr := rebuildProductRequest(
		"req-upd-app", "seller-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, valueobject.ProductRequestStatusApproved,
		nil, "email@test.com", "", "",
		time.Now().UTC(), time.Now().UTC(), nil,
	)

	err := pr.Update("New Title", "New Desc", "cat-001",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, nil, "new@email.com")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !apperrors.IsCode(err, "PRODUCT_REQUEST_NOT_EDITABLE") {
		t.Errorf("error code mismatch: got %v", err)
	}
	// Title must remain unchanged.
	if pr.Title() != "Title" {
		t.Errorf("Title() = %q, want unchanged %q", pr.Title(), "Title")
	}
}

func TestProductRequest_Update_NotEditableWhenRejected(t *testing.T) {
	t.Parallel()
	pr := rebuildProductRequest(
		"req-upd-rej", "seller-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, valueobject.ProductRequestStatusRejected,
		nil, "email@test.com", "reason", "",
		time.Now().UTC(), time.Now().UTC(), nil,
	)

	err := pr.Update("New Title", "New Desc", "cat-001",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, nil, "new@email.com")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !apperrors.IsCode(err, "PRODUCT_REQUEST_NOT_EDITABLE") {
		t.Errorf("error code mismatch: got %v", err)
	}
	if pr.Title() != "Title" {
		t.Errorf("Title() = %q, want unchanged %q", pr.Title(), "Title")
	}
}

func TestProductRequest_Update_StatusUnchanged(t *testing.T) {
	t.Parallel()
	pr := rebuildProductRequest(
		"req-upd-status", "seller-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
		nil, "email@test.com", "", "",
		time.Now().UTC(), time.Now().UTC(), nil,
	)

	err := pr.Update("New Title", "New Desc", "cat-001",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, nil, "new@email.com")

	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if pr.Status() != valueobject.ProductRequestStatusPending {
		t.Errorf("Status() = %v, want %v (status must not change on update)", pr.Status(), valueobject.ProductRequestStatusPending)
	}
}

// ── UnmarshalProductRequestFromDB ─────────────────────────────────────────────

func TestUnmarshalProductRequestFromDB(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	deletedAt := now.Add(-2 * time.Hour)
	images := makeAttachments()

	pr := rebuildProductRequest(
		"req-db", "seller-db", "cat-db", "Stored Title", "Stored Desc",
		customtypes.MustNewPrice("88.00"), customtypes.MustCurrency("GBP"),
		valueobject.ConditionPoor, valueobject.ProductRequestStatusRejected,
		images, "stored@email.com", "too expensive", "check again",
		now, now, &deletedAt,
	)

	if pr.ID() != "req-db" {
		t.Errorf("ID() = %q, want %q", pr.ID(), "req-db")
	}
	if pr.Status() != valueobject.ProductRequestStatusRejected {
		t.Errorf("Status() = %v, want %v", pr.Status(), valueobject.ProductRequestStatusRejected)
	}
	if pr.AdminRejectReason() != "too expensive" {
		t.Errorf("AdminRejectReason() = %q, want %q", pr.AdminRejectReason(), "too expensive")
	}
	if pr.DeletedAt() == nil || !pr.DeletedAt().Equal(deletedAt) {
		t.Errorf("DeletedAt() = %v, want %v", pr.DeletedAt(), deletedAt)
	}
}

// ── imageURLs ─────────────────────────────────────────────────────────────────

func TestProductRequest_imageURLs(t *testing.T) {
	t.Parallel()
	pr := rebuildProductRequest(
		"req-img", "seller-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
		valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
		customtypes.Attachments{
			{Key: "products/img-a.jpg", ContentType: "image/jpeg", Type: customtypes.AttachmentTypeImage},
			{Key: "products/img-b.png", ContentType: "image/png", Type: customtypes.AttachmentTypeImage},
		}, "email@test.com", "", "",
		time.Now().UTC(), time.Now().UTC(), nil,
	)

	urls := pr.imageURLs()
	if len(urls) != 2 {
		t.Fatalf("len(urls) = %d, want 2", len(urls))
	}
	if urls[0] != "products/img-a.jpg" {
		t.Errorf("urls[0] = %q, want %q", urls[0], "products/img-a.jpg")
	}
	if urls[1] != "products/img-b.png" {
		t.Errorf("urls[1] = %q, want %q", urls[1], "products/img-b.png")
	}
}

// ── validate ──────────────────────────────────────────────────────────────────

func TestProductRequest_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func() *ProductRequest
		wantErr bool
	}{
		{
			name: "valid product request passes",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"id", "seller", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: false,
		},
		{
			name: "empty id fails",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"  ", "seller", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "empty seller_id fails",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"id", "  ", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "empty category_id fails",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"id", "seller", "  ", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "empty title fails",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"id", "seller", "cat", "  ", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "zero expected_price fails",
			setup: func() *ProductRequest {
				return UnmarshalProductRequestFromDB(
					"id", "seller", "cat", "title", "desc",
					customtypes.MustNewPrice("0"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "invalid condition fails",
			setup: func() *ProductRequest {
				bad, _ := valueobject.NewConditionFromString("invalid")
				return UnmarshalProductRequestFromDB(
					"id", "seller", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					bad, valueobject.ProductRequestStatusPending,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
		{
			name: "invalid status fails",
			setup: func() *ProductRequest {
				bad, _ := valueobject.NewProductRequestStatusFromString("unknown")
				return UnmarshalProductRequestFromDB(
					"id", "seller", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), customtypes.MustCurrency("USD"),
					valueobject.ConditionNew, bad,
					nil, "email@test.com", "", "",
					time.Now().UTC(), time.Now().UTC(), nil,
				)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pr := tc.setup()
			err := pr.validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
