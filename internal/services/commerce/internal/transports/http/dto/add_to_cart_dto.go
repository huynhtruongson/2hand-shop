package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
)

// AddToCartRequestDTO is the HTTP request body for adding an item to the cart.
type AddToCartRequestDTO struct {
	ProductID   string            `json:"product_id" binding:"required"`
	ProductName string            `json:"product_name" binding:"required"`
	Price       customtypes.Price `json:"price" binding:"required"`
	// Currency is always USD — no field on the request.
}

// AddToCartResponseDTO is the HTTP response body after adding an item to the cart.
type AddToCartResponseDTO struct {
	CartID         string `json:"cart_id"`
	ItemID         string `json:"item_id"`
	TotalItemCount int    `json:"total_item_count"`
}
