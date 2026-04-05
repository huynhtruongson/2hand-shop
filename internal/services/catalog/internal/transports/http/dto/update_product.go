package dto

import (
	"github.com/LukaGiorgadze/gonull/v2"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
)

type ProductRequestID struct {
	ProductID string `uri:"product_id" binding:"required"`
}

type UpdateProductRequest struct {
	ProductID   string
	Title       gonull.Nullable[string]                  `json:"title" binding:"omitempty,max=200"`
	Description gonull.Nullable[string]                  `json:"description" binding:"omitempty,max=500"`
	Price       gonull.Nullable[customtypes.Price]       `json:"price" binding:"omitempty"`
	Condition   gonull.Nullable[string]                  `json:"condition" binding:"omitempty,oneof=new like_new good fair poor"`
	Images      gonull.Nullable[customtypes.Attachments] `json:"images" binding:"omitempty,min=1"`
	Brand       gonull.Nullable[string]                  `json:"brand" binding:"omitempty,max=100"`
}

func (req UpdateProductRequest) ToUpdateProductCommand() command.UpdateProductCommand {
	return command.UpdateProductCommand{
		ProductID:   req.ProductID,
		Title:       req.Title,
		Description: req.Description,
		Brand:       req.Brand,
		Price:       req.Price,
		Condition:   req.Condition,
		Images:      req.Images,
	}
}
