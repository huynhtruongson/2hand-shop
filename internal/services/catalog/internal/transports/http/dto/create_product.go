package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
)

type CreateProductRequest struct {
	CategoryID  string                  `json:"category_id" binding:"required"`
	Title       string                  `json:"title" binding:"required,max=200"`
	Description string                  `json:"description" binding:"required,max=500"`
	Brand       *string                 `json:"brand,omitempty"`
	Price       customtypes.Price       `json:"price" binding:"required"`
	Condition   string                  `json:"condition" binding:"required,oneof=new like_new good fair poor"`
	Images      customtypes.Attachments `json:"images" binding:"required,min=1,dive"`
}

func (req CreateProductRequest) ToCreateProductCommand() command.CreateProductCommand {
	return command.CreateProductCommand{
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Brand:       req.Brand,
		Price:       req.Price,
		Condition:   req.Condition,
		Images:      req.Images,
	}
}

type CreateProductResponse struct {
	ProductID string `json:"product_id"`
}
