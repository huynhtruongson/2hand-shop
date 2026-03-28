package aggregate

import (
	"slices"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// ProductRequest is the root aggregate for a client-submitted product sell request.
// It mirrors the Product aggregate with additional seller-supplied fields and a
// client-editable lifecycle controlled by the admin.
//
// Lifecycle:
//   - pending  → approved | rejected  (admin transitions; terminal once reached)
//   - pending  → pending             (seller may update mutable fields)
//   - approved | rejected            → dead end (no further mutations allowed)
//
// All fields are unexported; access is via getter methods.
type ProductRequest struct {
	id                string
	sellerID          string
	categoryID        string
	title             string
	description       string
	currency          customtypes.Currency
	condition         valueobject.Condition
	images            customtypes.Attachments
	status            valueobject.ProductRequestStatus
	expectedPrice     customtypes.Price // price the seller expects to receive
	contactInfo       string            // seller's preferred contact details
	adminRejectReason string            // set by admin when status becomes rejected
	adminNote         string            // internal admin note (not visible to seller)
	createdAt         time.Time
	updatedAt         time.Time
	deletedAt         *time.Time // soft-delete marker; nil means not deleted
}

// ── Constructor ──────────────────────────────────────────────────────────────

// NewProductRequest creates a new ProductRequest in pending status.
// It validates all fields and returns an error if any constraint is violated.
// The returned ProductRequest collects a ProductRequestCreated domain event.
func NewProductRequest(
	id, sellerID, categoryID, title, description string,
	expectedPrice customtypes.Price, currency customtypes.Currency, condition valueobject.Condition,
	images customtypes.Attachments, contactInfo string,
) (*ProductRequest, error) {

	pr := &ProductRequest{
		id:            id,
		sellerID:      sellerID,
		categoryID:    categoryID,
		title:         title,
		description:   description,
		currency:      currency,
		condition:     condition,
		images:        images,
		status:        valueobject.ProductRequestStatusPending,
		expectedPrice: expectedPrice,
		contactInfo:   contactInfo,
		createdAt:     time.Now().UTC(),
		updatedAt:     time.Now().UTC(),
	}
	if err := pr.validate(); err != nil {
		return nil, err
	}

	return pr, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (pr *ProductRequest) ID() string                               { return pr.id }
func (pr *ProductRequest) SellerID() string                         { return pr.sellerID }
func (pr *ProductRequest) CategoryID() string                       { return pr.categoryID }
func (pr *ProductRequest) Title() string                            { return pr.title }
func (pr *ProductRequest) Description() string                      { return pr.description }
func (pr *ProductRequest) Currency() customtypes.Currency           { return pr.currency }
func (pr *ProductRequest) Condition() valueobject.Condition         { return pr.condition }
func (pr *ProductRequest) Images() customtypes.Attachments          { return pr.images }
func (pr *ProductRequest) Status() valueobject.ProductRequestStatus { return pr.status }
func (pr *ProductRequest) ExpectedPrice() customtypes.Price         { return pr.expectedPrice }
func (pr *ProductRequest) ContactInfo() string                      { return pr.contactInfo }
func (pr *ProductRequest) AdminRejectReason() string                { return pr.adminRejectReason }
func (pr *ProductRequest) AdminNote() string                        { return pr.adminNote }
func (pr *ProductRequest) CreatedAt() time.Time                     { return pr.createdAt }
func (pr *ProductRequest) UpdatedAt() time.Time                     { return pr.updatedAt }
func (pr *ProductRequest) DeletedAt() *time.Time                    { return pr.deletedAt }

// ── Domain methods ────────────────────────────────────────────────────────────

// Update updates mutable fields of the ProductRequest.
// Only pending requests may be updated; approved and rejected are dead ends.
// Returns ErrProductRequestNotEditable if the request is not in pending status.
func (pr *ProductRequest) Update(
	title, description string, categoryID string,
	expectedPrice customtypes.Price, currency customtypes.Currency,
	condition valueobject.Condition, images customtypes.Attachments,
	contactInfo string,
) error {
	if pr.status != valueobject.ProductRequestStatusPending {
		return caterrors.ErrProductRequestNotEditable.
			WithMeta("current_status", pr.status.String()).
			WithMeta("action", "update")
	}

	pr.title = title
	pr.description = description
	pr.categoryID = categoryID
	pr.currency = currency
	pr.condition = condition
	pr.images = images
	pr.expectedPrice = expectedPrice
	pr.contactInfo = contactInfo
	pr.updatedAt = time.Now().UTC()

	if err := pr.validate(); err != nil {
		return err
	}

	return nil
}

// imageURLs is a helper that extracts raw URL strings from the Attachments.
func (pr *ProductRequest) imageURLs() []string {
	urls := make([]string, len(pr.images))
	for i, img := range pr.images {
		urls[i] = img.Key
	}
	return urls
}

// ── Factory / DB reconstruction ───────────────────────────────────────────────

// UnmarshalProductRequestFromDB reconstructs a ProductRequest from persistence storage.
// It skips validation so that stored (potentially legacy) data can still be loaded.
func UnmarshalProductRequestFromDB(
	id, sellerID, categoryID, title, description string,
	expectedPrice customtypes.Price, currency customtypes.Currency,
	condition valueobject.Condition, status valueobject.ProductRequestStatus,
	images customtypes.Attachments, contactInfo string,
	adminRejectReason string, adminNote string,
	createdAt, updatedAt time.Time, deletedAt *time.Time,
) *ProductRequest {
	return &ProductRequest{
		id:                id,
		sellerID:          sellerID,
		categoryID:        categoryID,
		title:             title,
		description:       description,
		currency:          currency,
		condition:         condition,
		images:            images,
		status:            status,
		expectedPrice:     expectedPrice,
		contactInfo:       contactInfo,
		adminRejectReason: adminRejectReason,
		adminNote:         adminNote,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		deletedAt:         deletedAt,
	}
}

// ── Validation ────────────────────────────────────────────────────────────────

// validate enforces all invariant constraints on the ProductRequest aggregate.
func (pr *ProductRequest) validate() error {
	switch {
	case strings.TrimSpace(pr.id) == "":
		return caterrors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(pr.sellerID) == "":
		return caterrors.ErrValidation.WithDetail("seller_id", "seller_id is empty")
	case strings.TrimSpace(pr.categoryID) == "":
		return caterrors.ErrValidation.WithDetail("category_id", "category_id is empty")
	case strings.TrimSpace(pr.title) == "":
		return caterrors.ErrValidation.WithDetail("title", "title is empty")
	case !pr.expectedPrice.IsPositive():
		return caterrors.ErrValidation.WithDetail("expected_price", "expected_price must be positive")
	case pr.currency == "":
		return caterrors.ErrValidation.WithDetail("currency", "currency is empty")
	case !pr.condition.IsValid():
		return caterrors.ErrValidation.WithDetail("condition", "condition is not a valid value")
	case !isValidProductRequestStatus(pr.status):
		return caterrors.ErrValidation.WithDetail("status", "status is not a valid value")
	}
	return nil
}

// isValidProductRequestStatus returns true if s is one of the defined product request statuses.
func isValidProductRequestStatus(s valueobject.ProductRequestStatus) bool {
	return slices.Contains(valueobject.AllProductRequestStatuses(), s)
}
