package dto

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
)

// CartItemDTO is the JSON-friendly representation of a cart item.
// The internal cart_id field is omitted.
type CartItemDTO struct {
	ID          string            `json:"id"`
	ProductID   string            `json:"product_id"`
	ProductName string            `json:"product_name"`
	Price       customtypes.Price `json:"price"`
	Currency    string            `json:"currency"` // ISO 4217 code, always "USD"
	AddedAt     time.Time         `json:"added_at"`
}

// GetCartResponseDTO is the HTTP response body for GET /cart.
type GetCartResponseDTO struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	Items       []CartItemDTO      `json:"items"`
	ItemCount   int                `json:"item_count"`
	TotalAmount customtypes.Price   `json:"total_amount"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}
