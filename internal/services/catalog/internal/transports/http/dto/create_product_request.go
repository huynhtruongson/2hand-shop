package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
)

type CreateProductRequestDTO struct {
	SellerID      string                  `json:"seller_id"` // injected by handler from auth context — not from request body
	CategoryID    string                  `json:"category_id" binding:"required"`
	Title         string                  `json:"title" binding:"required,max=200"`
	Description   string                  `json:"description" binding:"required,max=500"`
	Brand         *string                 `json:"brand,omitempty"`
	ExpectedPrice customtypes.Price       `json:"expected_price" binding:"required"`
	Condition     string                  `json:"condition" binding:"required,oneof=new like_new good fair poor"`
	Images        customtypes.Attachments `json:"images" binding:"required,min=1,dive"`
	ContactInfo   string                  `json:"contact_info" binding:"required"`
	AdminNote     *string                 `json:"admin_note,omitempty"`
}

func (dto CreateProductRequestDTO) ToCreateProductRequestCommand() command.CreateProductRequestCommand {
	return command.CreateProductRequestCommand{
		SellerID:      dto.SellerID,
		CategoryID:    dto.CategoryID,
		Title:         dto.Title,
		Description:   dto.Description,
		Brand:         dto.Brand,
		ExpectedPrice: dto.ExpectedPrice,
		Condition:     dto.Condition,
		Images:        dto.Images,
		ContactInfo:   dto.ContactInfo,
		AdminNote:     dto.AdminNote,
	}
}

type CreateProductRequestResponseDTO struct {
	ProductRequestID string `json:"product_request_id"`
}
