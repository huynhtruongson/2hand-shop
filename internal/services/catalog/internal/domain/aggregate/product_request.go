package aggregate

import (
	"slices"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

type ProductRequest struct {
	id                string
	sellerID          string
	categoryID        string
	title             string
	description       string
	brand             *string
	currency          valueobject.Currency
	condition         valueobject.Condition
	status            valueobject.ProductRequestStatus
	images            customtypes.Attachments
	expectedPrice     customtypes.Price
	contactInfo       string
	adminRejectReason string
	adminNote         string
	createdAt         time.Time
	updatedAt         time.Time
	deletedAt         *time.Time
}

func NewProductRequest(
	id, sellerID, categoryID, title, description string, brand *string,
	expectedPrice customtypes.Price, condition valueobject.Condition,
	images customtypes.Attachments, contactInfo string,
) (*ProductRequest, error) {

	pr := &ProductRequest{
		id:            id,
		sellerID:      sellerID,
		categoryID:    categoryID,
		title:         title,
		description:   description,
		brand:         brand,
		currency:      valueobject.CurrencyUSD,
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

func (pr *ProductRequest) ID() string                               { return pr.id }
func (pr *ProductRequest) SellerID() string                         { return pr.sellerID }
func (pr *ProductRequest) CategoryID() string                       { return pr.categoryID }
func (pr *ProductRequest) Title() string                            { return pr.title }
func (pr *ProductRequest) Description() string                      { return pr.description }
func (pr *ProductRequest) Brand() *string                           { return pr.brand }
func (pr *ProductRequest) Currency() valueobject.Currency           { return pr.currency }
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

func (pr *ProductRequest) MarkDeleted() {
	now := time.Now().UTC()
	pr.deletedAt = &now
	pr.updatedAt = now
}

func (pr *ProductRequest) Update(
	title, description string, categoryID string, brand *string,
	expectedPrice customtypes.Price, condition valueobject.Condition, images customtypes.Attachments,
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
	pr.condition = condition
	pr.images = images
	pr.brand = brand
	pr.expectedPrice = expectedPrice
	pr.contactInfo = contactInfo
	pr.updatedAt = time.Now().UTC()

	if err := pr.validate(); err != nil {
		return err
	}

	return nil
}

func UnmarshalProductRequestFromDB(
	id, sellerID, categoryID, title, description string,
	expectedPrice customtypes.Price, currency valueobject.Currency, brand *string,
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
		brand:             brand,
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
	case !pr.currency.IsValid():
		return caterrors.ErrValidation.WithDetail("currency", "currency is not a valid value")
	case !pr.condition.IsValid():
		return caterrors.ErrValidation.WithDetail("condition", "condition is not a valid value")
	case !isValidProductRequestStatus(pr.status):
		return caterrors.ErrValidation.WithDetail("status", "status is not a valid value")
	}
	return nil
}

func isValidProductRequestStatus(s valueobject.ProductRequestStatus) bool {
	return slices.Contains(valueobject.AllProductRequestStatuses(), s)
}
