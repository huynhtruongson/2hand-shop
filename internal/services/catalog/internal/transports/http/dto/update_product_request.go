package dto

import (
	"github.com/LukaGiorgadze/gonull/v2"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
)

// UpdateProductRequestDTO is the HTTP request body for updating a product request.
// SellerID is injected by the handler from the authenticated user context.
type UpdateProductRequestDTO struct {
	CategoryID    gonull.Nullable[string]                  `json:"category_id" binding:"omitempty"`
	Title         gonull.Nullable[string]                  `json:"title" binding:"omitempty,max=200"`
	Description   gonull.Nullable[string]                  `json:"description" binding:"omitempty,max=500"`
	Brand         gonull.Nullable[string]                  `json:"brand" binding:"omitempty,max=100"`
	ExpectedPrice gonull.Nullable[customtypes.Price]       `json:"expected_price" binding:"omitempty"`
	Condition     gonull.Nullable[string]                  `json:"condition" binding:"omitempty,oneof=new like_new good fair poor"`
	Images        gonull.Nullable[customtypes.Attachments] `json:"images" binding:"omitempty,min=1,dive"`
	ContactInfo   gonull.Nullable[string]                  `json:"contact_info" binding:"omitempty"`
}

// ToUpdateProductRequestCommand converts the DTO to the application-layer command.
func (req UpdateProductRequestDTO) ToUpdateProductRequestCommand(productRequestID, sellerID string) command.UpdateProductRequestCommand {
	return command.UpdateProductRequestCommand{
		ProductRequestID: productRequestID,
		SellerID:          sellerID,
		CategoryID:         req.CategoryID,
		Title:              req.Title,
		Description:        req.Description,
		Brand:              req.Brand,
		ExpectedPrice:      req.ExpectedPrice,
		Condition:          req.Condition,
		Images:             req.Images,
		ContactInfo:        req.ContactInfo,
	}
}