package aggregate

import (
	"slices"
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

type Product struct {
	id          string
	categoryID  string
	title       string
	description string
	brand       *string
	price       customtypes.Price
	currency    valueobject.Currency
	condition   valueobject.Condition
	status      valueobject.ProductStatus
	images      customtypes.Attachments
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

func NewProduct(
	id, categoryID, title, description string,
	price customtypes.Price, condition valueobject.Condition,
	images customtypes.Attachments,
	brand *string,
) (*Product, error) {
	p := &Product{
		id:          id,
		categoryID:  categoryID,
		title:       title,
		description: description,
		brand:       brand,
		price:       price,
		currency:    valueobject.CurrencyUSD,
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
func (p *Product) CategoryID() string                { return p.categoryID }
func (p *Product) Title() string                     { return p.title }
func (p *Product) Description() string               { return p.description }
func (p *Product) Brand() *string                    { return p.brand }
func (p *Product) Price() customtypes.Price          { return p.price }
func (p *Product) Currency() valueobject.Currency    { return p.currency }
func (p *Product) Condition() valueobject.Condition  { return p.condition }
func (p *Product) Status() valueobject.ProductStatus { return p.status }
func (p *Product) Images() customtypes.Attachments   { return p.images }
func (p *Product) CreatedAt() time.Time              { return p.createdAt }
func (p *Product) UpdatedAt() time.Time              { return p.updatedAt }
func (p *Product) DeletedAt() *time.Time             { return p.deletedAt }

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

func (p *Product) Update(title, description string, price customtypes.Price, condition valueobject.Condition, images customtypes.Attachments, brand *string) error {
	// Only draft and active products can be updated.
	if p.status == valueobject.ProductStatusSold || p.status == valueobject.ProductStatusArchived {
		return caterrors.ErrProductInvalidStatusTransition.
			WithMeta("current_status", p.status.String()).
			WithMeta("action", "update")
	}

	p.title = title
	p.description = description
	p.price = price
	p.condition = condition
	p.images = images
	if brand != nil {
		p.brand = brand
	}
	p.updatedAt = time.Now().UTC()
	if err := p.validate(); err != nil {
		return err
	}
	return nil
}

func UnmarshalProductFromDB(
	id, categoryID, title, description string,
	price customtypes.Price, currency valueobject.Currency,
	condition valueobject.Condition, status valueobject.ProductStatus,
	images customtypes.Attachments,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	brand *string,
) *Product {
	return &Product{
		id:          id,
		categoryID:  categoryID,
		title:       title,
		description: description,
		brand:       brand,
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
	case !p.currency.IsValid():
		return caterrors.ErrValidation.WithDetail("currency", "currency is not a valid value")
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
