package dto

// RemoveFromCartRequestDTO is the HTTP request for removing an item from the cart.
// The product_id comes from the URI path: DELETE /cart/:product_id.
type RemoveFromCartRequestDTO struct {
	ProductID string `uri:"product_id" binding:"required"`
}

// RemoveFromCartResponseDTO is the HTTP response after removing an item from the cart.
type RemoveFromCartResponseDTO struct {
	CartID         string `json:"cart_id"`
	TotalItemCount int    `json:"total_item_count"`
}