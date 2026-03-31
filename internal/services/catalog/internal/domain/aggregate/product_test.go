package aggregate

import (
	"errors"
	"testing"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
	apperrors "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

// ── Test helpers ─────────────────────────────────────────────────────────────

func makeAttachments() customtypes.Attachments {
	return customtypes.Attachments{
		{Key: "products/img-1.jpg", ContentType: "image/jpeg", Type: customtypes.AttachmentTypeImage},
	}
}

func rebuildProduct(
	id, categoryID, title, description string,
	price customtypes.Price, currency valueobject.Currency,
	condition valueobject.Condition, status valueobject.ProductStatus,
	images customtypes.Attachments,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
	brand *string,
) *Product {
	return UnmarshalProductFromDB(
		id, categoryID, title, description,
		price, currency, condition, status,
		images, createdAt, updatedAt, deletedAt, brand,
	)
}

// ── NewProduct ───────────────────────────────────────────────────────────────

func TestNewProduct_Valid(t *testing.T) {
	t.Parallel()
	p, err := NewProduct(
		"prod-001", "cat-001",
		"Vintage Lamp", "A beautiful vintage desk lamp",
		customtypes.MustNewPrice("25.50"),
		valueobject.ConditionGood, makeAttachments(),
		nil,
	)
	if err != nil {
		t.Fatalf("NewProduct() unexpected error: %v", err)
	}
	if p.ID() != "prod-001" {
		t.Errorf("ID() = %q, want %q", p.ID(), "prod-001")
	}
	if p.CategoryID() != "cat-001" {
		t.Errorf("CategoryID() = %q, want %q", p.CategoryID(), "cat-001")
	}
	if p.Title() != "Vintage Lamp" {
		t.Errorf("Title() = %q, want %q", p.Title(), "Vintage Lamp")
	}
	if p.Description() != "A beautiful vintage desk lamp" {
		t.Errorf("Description() mismatch")
	}
	if !p.Price().Equal(customtypes.MustNewPrice("25.50")) {
		t.Errorf("Price() = %v, want %v", p.Price(), customtypes.MustNewPrice("25.50"))
	}
	if p.Currency() != valueobject.CurrencyUSD {
		t.Errorf("Currency() = %v, want %v", p.Currency(), valueobject.CurrencyUSD)
	}
	if p.Condition() != valueobject.ConditionGood {
		t.Errorf("Condition() = %v, want %v", p.Condition(), valueobject.ConditionGood)
	}
	if p.Status() != valueobject.ProductStatusDraft {
		t.Errorf("Status() = %v, want %v", p.Status(), valueobject.ProductStatusDraft)
	}
	if len(p.Images()) != 1 {
		t.Errorf("Images() len = %d, want 1", len(p.Images()))
	}
	if p.DeletedAt() != nil {
		t.Error("DeletedAt() = not nil, want nil for a new product")
	}
}

func TestNewProduct_ValidationErrors(t *testing.T) {
	// wantErr is the ae.Details map key; "" = no error expected.
	//
	// Each test case calls NewProduct directly so no default-overwrite logic
	// accidentally masks the field under test.
	t.Run("empty id", func(t *testing.T) {
		_, err := NewProduct("  ", "cat-001",
			"Vintage Lamp", "A beautiful vintage desk lamp",
			customtypes.MustNewPrice("25.50"),
			valueobject.ConditionGood, makeAttachments(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		assertAppErrorDetail(t, err, "id")
	})
	t.Run("empty category_id", func(t *testing.T) {
		_, err := NewProduct("prod-001", "  ",
			"Vintage Lamp", "A beautiful vintage desk lamp",
			customtypes.MustNewPrice("25.50"),
			valueobject.ConditionGood, makeAttachments(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		assertAppErrorDetail(t, err, "category_id")
	})
	t.Run("empty title", func(t *testing.T) {
		_, err := NewProduct("prod-001", "cat-001",
			"  ", "A beautiful vintage desk lamp",
			customtypes.MustNewPrice("25.50"),
			valueobject.ConditionGood, makeAttachments(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		assertAppErrorDetail(t, err, "title")
	})
	t.Run("zero price", func(t *testing.T) {
		_, err := NewProduct("prod-001", "cat-001",
			"Vintage Lamp", "A beautiful vintage desk lamp",
			customtypes.MustNewPrice("0"),
			valueobject.ConditionGood, makeAttachments(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		assertAppErrorDetail(t, err, "price")
	})
	t.Run("negative price", func(t *testing.T) {
		// NewPrice rejects negative strings, so construct via the embedded decimal.
		negPrice := customtypes.Price{Decimal: decimal.NewFromFloat(-5.0)}
		_, err := NewProduct("prod-001", "cat-001",
			"Vintage Lamp", "A beautiful vintage desk lamp",
			negPrice,
			valueobject.ConditionGood, makeAttachments(), nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		assertAppErrorDetail(t, err, "price")
	})
	t.Run("images may be nil", func(t *testing.T) {
		_, err := NewProduct("prod-001", "cat-001",
			"Vintage Lamp", "A beautiful vintage desk lamp",
			customtypes.MustNewPrice("25.50"),
			valueobject.ConditionGood, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func assertAppErrorDetail(t *testing.T, err error, fieldKey string) {
	t.Helper()
	var ae *apperrors.AppError
	if !errors.As(err, &ae) {
		t.Fatalf("error does not implement *AppError: %T", err)
	}
	ue := ae.UserFacing()
	if _, ok := ue.Details[fieldKey]; !ok {
		t.Errorf("UserFacing().Details[%q] missing; full details = %v", fieldKey, ue.Details)
	}
}

// ── Getters ───────────────────────────────────────────────────────────────────

func TestProduct_Getters(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	p := rebuildProduct(
		"prod-getter-001", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("10.00"), valueobject.CurrencyEUR,
		valueobject.ConditionLikeNew, valueobject.ProductStatusPublished,
		makeAttachments(),
		now, now, nil,
		nil,
	)

	if got := p.ID(); got != "prod-getter-001" {
		t.Errorf("ID() = %q, want %q", got, "prod-getter-001")
	}
	if got := p.CategoryID(); got != "cat-001" {
		t.Errorf("CategoryID() = %q, want %q", got, "cat-001")
	}
	if got := p.Title(); got != "Title" {
		t.Errorf("Title() = %q, want %q", got, "Title")
	}
	if got := p.Description(); got != "Desc" {
		t.Errorf("Description() = %q, want %q", got, "Desc")
	}
	if got := p.Status(); got != valueobject.ProductStatusPublished {
		t.Errorf("Status() = %v, want %v", got, valueobject.ProductStatusPublished)
	}
	if got := p.CreatedAt(); !got.Equal(now) {
		t.Errorf("CreatedAt() = %v, want %v", got, now)
	}
	if got := p.UpdatedAt(); !got.Equal(now) {
		t.Errorf("UpdatedAt() = %v, want %v", got, now)
	}
	if got := p.DeletedAt(); got != nil {
		t.Errorf("DeletedAt() = %v, want nil", got)
	}
}

// ── MarkDeleted ───────────────────────────────────────────────────────────────

func TestProduct_MarkDeleted(t *testing.T) {
	t.Parallel()
	p := rebuildProduct(
		"prod-del", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusDraft,
		nil,
		time.Now().UTC(), time.Now().UTC(), nil,
		nil,
	)

	p.MarkDeleted()

	if p.DeletedAt() == nil {
		t.Fatal("DeletedAt() = nil, want non-nil")
	}
	if p.UpdatedAt().Before(p.DeletedAt().Add(-time.Second)) {
		t.Errorf("UpdatedAt() = %v, want >= %v", p.UpdatedAt(), p.DeletedAt().Add(-time.Second))
	}
}

// ── Publish ───────────────────────────────────────────────────────────────────

func TestProduct_Publish(t *testing.T) {
	t.Parallel()
	p := rebuildProduct(
		"prod-pub", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusDraft,
		nil,
		time.Now().UTC(), time.Now().UTC(), nil,
		nil,
	)

	err := p.Publish()
	if err != nil {
		t.Fatalf("Publish() unexpected error: %v", err)
	}
	if p.Status() != valueobject.ProductStatusPublished {
		t.Errorf("Status() = %v, want %v", p.Status(), valueobject.ProductStatusPublished)
	}
}

func TestProduct_Publish_InvalidTransition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status valueobject.ProductStatus
	}{
		{"published_already", valueobject.ProductStatusPublished},
		{"sold_is_terminal", valueobject.ProductStatusSold},
		{"archived_is_terminal", valueobject.ProductStatusArchived},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := rebuildProduct(
				"prod-pub-err", "cat-001", "Title", "Desc",
				customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
				valueobject.ConditionNew, tc.status,
				nil,
				time.Now().UTC(), time.Now().UTC(), nil,
				nil,
			)

			err := p.Publish()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !apperrors.IsCode(err, "PRODUCT_INVALID_STATUS_TRANSITION") {
				t.Errorf("error code mismatch: got %v", err)
			}
			if p.Status() != tc.status {
				t.Errorf("Status() changed to %v, want unchanged %v", p.Status(), tc.status)
			}
		})
	}
}

// ── MarkSold ──────────────────────────────────────────────────────────────────

func TestProduct_MarkSold(t *testing.T) {
	t.Parallel()
	p := rebuildProduct(
		"prod-sold", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusPublished,
		nil,
		time.Now().UTC(), time.Now().UTC(), nil,
		nil,
	)

	err := p.MarkSold("order-001")
	if err != nil {
		t.Fatalf("MarkSold() unexpected error: %v", err)
	}
	if p.Status() != valueobject.ProductStatusSold {
		t.Errorf("Status() = %v, want %v", p.Status(), valueobject.ProductStatusSold)
	}
}

func TestProduct_MarkSold_InvalidTransition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status valueobject.ProductStatus
	}{
		{"draft", valueobject.ProductStatusDraft},
		{"sold_already", valueobject.ProductStatusSold},
		{"archived", valueobject.ProductStatusArchived},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := rebuildProduct(
				"prod-sold-err", "cat-001", "Title", "Desc",
				customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
				valueobject.ConditionNew, tc.status,
				nil,
				time.Now().UTC(), time.Now().UTC(), nil,
				nil,
			)

			err := p.MarkSold("order-001")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !apperrors.IsCode(err, "PRODUCT_INVALID_STATUS_TRANSITION") {
				t.Errorf("error code mismatch: got %v", err)
			}
			if p.Status() != tc.status {
				t.Errorf("Status() changed to %v, want unchanged %v", p.Status(), tc.status)
			}
		})
	}
}

// ── Archive ────────────────────────────────────────────────────────────────────

func TestProduct_Archive(t *testing.T) {
	t.Parallel()
	p := rebuildProduct(
		"prod-arch", "cat-001", "Title", "Desc",
		customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
		valueobject.ConditionNew, valueobject.ProductStatusPublished,
		nil,
		time.Now().UTC(), time.Now().UTC(), nil,
		nil,
	)

	err := p.Archive("seller-001")
	if err != nil {
		t.Fatalf("Archive() unexpected error: %v", err)
	}
	if p.Status() != valueobject.ProductStatusArchived {
		t.Errorf("Status() = %v, want %v", p.Status(), valueobject.ProductStatusArchived)
	}
}

func TestProduct_Archive_InvalidTransition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status valueobject.ProductStatus
	}{
		{"draft", valueobject.ProductStatusDraft},
		{"sold", valueobject.ProductStatusSold},
		{"archived_already", valueobject.ProductStatusArchived},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := rebuildProduct(
				"prod-arch-err", "cat-001", "Title", "Desc",
				customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
				valueobject.ConditionNew, tc.status,
				nil,
				time.Now().UTC(), time.Now().UTC(), nil,
				nil,
			)

			err := p.Archive("seller-001")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !apperrors.IsCode(err, "PRODUCT_INVALID_STATUS_TRANSITION") {
				t.Errorf("error code mismatch: got %v", err)
			}
			if p.Status() != tc.status {
				t.Errorf("Status() changed to %v, want unchanged %v", p.Status(), tc.status)
			}
		})
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestProduct_Update(t *testing.T) {
	t.Parallel()
	p := rebuildProduct(
		"prod-upd", "cat-001", "Old Title", "Old Desc",
		customtypes.MustNewPrice("5.00"), valueobject.CurrencyUSD,
		valueobject.ConditionFair, valueobject.ProductStatusDraft,
		nil,
		time.Now().UTC(), time.Now().UTC(), nil,
		nil,
	)

	newImages := makeAttachments()
	err := p.Update("New Title", "New Desc",
		customtypes.MustNewPrice("15.00"),
		valueobject.ConditionLikeNew, newImages, nil)

	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if p.Title() != "New Title" {
		t.Errorf("Title() = %q, want %q", p.Title(), "New Title")
	}
	if p.Description() != "New Desc" {
		t.Errorf("Description() = %q, want %q", p.Description(), "New Desc")
	}
	if !p.Price().Equal(customtypes.MustNewPrice("15.00")) {
		t.Errorf("Price() = %v, want %v", p.Price(), customtypes.MustNewPrice("15.00"))
	}
	if p.Currency() != valueobject.CurrencyUSD {
		t.Errorf("Currency() = %v, want %v", p.Currency(), valueobject.CurrencyUSD)
	}
	if p.Condition() != valueobject.ConditionLikeNew {
		t.Errorf("Condition() = %v, want %v", p.Condition(), valueobject.ConditionLikeNew)
	}
}

func TestProduct_Update_TerminalStatusBlocked(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status valueobject.ProductStatus
	}{
		{"sold", valueobject.ProductStatusSold},
		{"archived", valueobject.ProductStatusArchived},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := rebuildProduct(
				"prod-upd-err", "cat-001", "Title", "Desc",
				customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
				valueobject.ConditionNew, tc.status,
				nil,
				time.Now().UTC(), time.Now().UTC(), nil,
				nil,
			)

			err := p.Update("New Title", "New Desc",
				customtypes.MustNewPrice("1.00"),
				valueobject.ConditionNew, nil, nil)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !apperrors.IsCode(err, "PRODUCT_INVALID_STATUS_TRANSITION") {
				t.Errorf("error code mismatch: got %v", err)
			}
			if p.Title() != "Title" {
				t.Errorf("Title() = %q, want unchanged %q", p.Title(), "Title")
			}
		})
	}
}

// ── UnmarshalProductFromDB ────────────────────────────────────────────────────

func TestUnmarshalProductFromDB(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	deletedAt := now.Add(-time.Hour)
	images := makeAttachments()

	p := rebuildProduct(
		"prod-db", "cat-001", "Stored Title", "Stored Desc",
		customtypes.MustNewPrice("99.99"), valueobject.CurrencyGBP,
		valueobject.ConditionGood, valueobject.ProductStatusArchived,
		images,
		now, now, &deletedAt,
		nil,
	)

	if p.ID() != "prod-db" {
		t.Errorf("ID() = %q, want %q", p.ID(), "prod-db")
	}
	if p.Status() != valueobject.ProductStatusArchived {
		t.Errorf("Status() = %v, want %v", p.Status(), valueobject.ProductStatusArchived)
	}
	if p.DeletedAt() == nil || !p.DeletedAt().Equal(deletedAt) {
		t.Errorf("DeletedAt() = %v, want %v", p.DeletedAt(), deletedAt)
	}
}

// ── validate ──────────────────────────────────────────────────────────────────

func TestProduct_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func() *Product
		wantErr bool
	}{
		{
			name: "valid product passes",
			setup: func() *Product {
				return UnmarshalProductFromDB(
					"id", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: false,
		},
		{
			name: "empty id fails",
			setup: func() *Product {
				return UnmarshalProductFromDB(
					"  ", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "empty category_id fails",
			setup: func() *Product {
				return UnmarshalProductFromDB(
					"id", "  ", "title", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "empty title fails",
			setup: func() *Product {
				return UnmarshalProductFromDB(
					"id", "cat", "  ", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "zero price fails",
			setup: func() *Product {
				return UnmarshalProductFromDB(
					"id", "cat", "title", "desc",
					customtypes.MustNewPrice("0"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "invalid condition fails",
			setup: func() *Product {
				bad, _ := valueobject.NewConditionFromString("invalid")
				return UnmarshalProductFromDB(
					"id", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					bad, valueobject.ProductStatusDraft,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "invalid status fails",
			setup: func() *Product {
				bad, _ := valueobject.NewProductStatusFromString("unknown")
				return UnmarshalProductFromDB(
					"id", "cat", "title", "desc",
					customtypes.MustNewPrice("1.00"), valueobject.CurrencyUSD,
					valueobject.ConditionNew, bad,
					nil, time.Now().UTC(), time.Now().UTC(), nil, nil,
				)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := tc.setup()
			err := p.validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
