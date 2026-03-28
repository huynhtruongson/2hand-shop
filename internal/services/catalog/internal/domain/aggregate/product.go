package aggregate

import (
	"slices"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// Product is the root aggregate for the Catalog bounded context.
// It encapsulates all business rules for a second-hand product listing.
// All fields are unexported; access is via getter methods.
type Product struct {
	id          string
	sellerID    string
	categoryID  string
	title       string
	description string
	price       customtypes.Price
	currency    customtypes.Currency
	condition   valueobject.Condition
	status      valueobject.ProductStatus
	images      customtypes.Attachments
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// ── Constructor ─────────────────────────────────────────────────────────────

// NewProduct creates a new Product in draft status.
// It validates all fields and returns an error if any constraint is violated.
// The returned Product collects a ProductCreated domain event.
func NewProduct(
	id, sellerID, categoryID, title, description string,
	price customtypes.Price, currency customtypes.Currency, condition valueobject.Condition,
	images customtypes.Attachments,
) (*Product, error) {
	p := &Product{
		id:          id,
		sellerID:    sellerID,
		categoryID:  categoryID,
		title:       title,
		description: description,
		price:       price,
		currency:    currency,
		condition:   condition,
		status:      valueobject.ProductStatusDraft,
		images:      images,
		createdAt:   time.Now().UTC(),
		updatedAt:   time.Now().UTC(),
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (p *Product) ID() string                        { return p.id }
func (p *Product) SellerID() string                  { return p.sellerID }
func (p *Product) CategoryID() string                { return p.categoryID }
func (p *Product) Title() string                     { return p.title }
func (p *Product) Description() string               { return p.description }
func (p *Product) Price() customtypes.Price          { return p.price }
func (p *Product) Currency() customtypes.Currency    { return p.currency }
func (p *Product) Condition() valueobject.Condition  { return p.condition }
func (p *Product) Status() valueobject.ProductStatus { return p.status }
func (p *Product) Images() customtypes.Attachments   { return p.images }
func (p *Product) CreatedAt() time.Time              { return p.createdAt }
func (p *Product) UpdatedAt() time.Time              { return p.updatedAt }
func (p *Product) DeletedAt() *time.Time             { return p.deletedAt }

// MarkDeleted soft-deletes the product by recording the current UTC time
// in the deletedAt field. Soft-deleted products must not appear in query results.
func (p *Product) MarkDeleted() {
	now := time.Now().UTC()
	p.deletedAt = &now
	p.updatedAt = now
}

func (p *Product) Publish() error {
	if !p.status.CanTransitionTo(valueobject.ProductStatusPublished) {
		return caterrors.ErrProductInvalidStatusTransition.
			WithMeta("current_status", p.status.String()).
			WithMeta("target_status", "active")
	}
	p.status = valueobject.ProductStatusPublished
	p.updatedAt = time.Now().UTC()

	return nil
}

// MarkSold transitions the product from active to sold.
// This is called by the CommerceService event handler when an order is placed.
// Returns ErrProductInvalidStatusTransition if the product is not active.
func (p *Product) MarkSold(orderID string) error {
	if !p.status.CanTransitionTo(valueobject.ProductStatusSold) {
		return caterrors.ErrProductInvalidStatusTransition.
			WithMeta("current_status", p.status.String()).
			WithMeta("target_status", "sold")
	}
	p.status = valueobject.ProductStatusSold
	p.updatedAt = time.Now().UTC()
	return nil
}

// Archive transitions the product from active to archived.
// Only the owning seller may archive a product.
// Returns ErrProductInvalidStatusTransition if the product is not active.
func (p *Product) Archive(actorID string) error {
	if !p.status.CanTransitionTo(valueobject.ProductStatusArchived) {
		return caterrors.ErrProductInvalidStatusTransition.
			WithMeta("current_status", p.status.String()).
			WithMeta("target_status", "archived")
	}
	p.status = valueobject.ProductStatusArchived
	p.updatedAt = time.Now().UTC()
	return nil
}

// ── Business logic — field mutations ─────────────────────────────────────────

// Update updates mutable fields of the product.
// Only draft or active products may be updated; sold and archived are terminal.
func (p *Product) Update(title, description string, price customtypes.Price, currency customtypes.Currency, condition valueobject.Condition, images customtypes.Attachments) error {
	// Only draft and active products can be updated.
	if p.status == valueobject.ProductStatusSold || p.status == valueobject.ProductStatusArchived {
		return caterrors.ErrProductInvalidStatusTransition.
			WithMeta("current_status", p.status.String()).
			WithMeta("action", "update")
	}

	p.title = title
	p.description = description
	p.price = price
	p.currency = currency
	p.condition = condition
	p.images = images
	p.updatedAt = time.Now().UTC()
	if err := p.validate(); err != nil {
		return err
	}
	return nil
}

// ── Factory / DB reconstruction ─────────────────────────────────────────────

// UnmarshalProductFromDB reconstructs a Product from persistence storage.
// It skips validation so that stored (potentially legacy) data can still be loaded.
func UnmarshalProductFromDB(
	id, sellerID, categoryID, title, description string,
	price customtypes.Price, currency customtypes.Currency,
	condition valueobject.Condition, status valueobject.ProductStatus,
	images customtypes.Attachments,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		sellerID:    sellerID,
		categoryID:  categoryID,
		title:       title,
		description: description,
		price:       price,
		currency:    currency,
		condition:   condition,
		status:      status,
		images:      images,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

// validate enforces all invariant constraints on the Product aggregate.
func (p *Product) validate() error {
	switch {
	case strings.TrimSpace(p.id) == "":
		return caterrors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(p.categoryID) == "":
		return caterrors.ErrValidation.WithDetail("category_id", "category_id is empty")
	case strings.TrimSpace(p.title) == "":
		return caterrors.ErrValidation.WithDetail("title", "title is empty")
	case !p.price.IsPositive():
		return caterrors.ErrValidation.WithDetail("price", "price must be positive")
	case p.currency == "":
		return caterrors.ErrValidation.WithDetail("currency", "currency is empty")
	case !p.condition.IsValid():
		return caterrors.ErrValidation.WithDetail("condition", "condition is not a valid value")
	case !isValidProductStatus(p.status):
		return caterrors.ErrValidation.WithDetail("status", "status is not a valid value")
	}
	return nil
}

// isValidProductStatus returns true if s is one of the defined product statuses.
func isValidProductStatus(s valueobject.ProductStatus) bool {
	return slices.Contains(valueobject.AllProductStatuses(), s)
}
