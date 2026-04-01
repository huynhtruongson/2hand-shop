package event

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// ProductPayload holds all product fields published as part of a domain event.
type ProductPayload struct {
	ID          string                     `json:"id"`
	CategoryID  string                     `json:"category_id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Brand       *string                    `json:"brand,omitempty"`
	Price       customtypes.Price          `json:"price"`
	Currency    valueobject.Currency        `json:"currency"`
	Condition   valueobject.Condition      `json:"condition"`
	Status      valueobject.ProductStatus   `json:"status"`
	Images      customtypes.Attachments     `json:"images"`
	CreatedAt   time.Time                  `json:"created_at"`
}
