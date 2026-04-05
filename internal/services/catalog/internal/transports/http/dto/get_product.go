package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// GetProductRequest binds the product_id path parameter.
type GetProductRequest struct {
	ProductID string `uri:"product_id" binding:"required"`
}

// ToGetProductQuery converts the request into a GetProductQuery.
func (req GetProductRequest) ToGetProductQuery() query.GetProductQuery {
	return query.GetProductQuery{ProductID: req.ProductID}
}

// ToProductDTO converts a domain Product aggregate into a Product response DTO.
func ToProductDTO(p aggregate.Product) Product {
	return Product{
		ID:          p.ID(),
		Title:       p.Title(),
		CategoryID:  p.CategoryID(),
		Description: p.Description(),
		Brand:       p.Brand(),
		Price:       p.Price(),
		Currency:    p.Currency().String(),
		Condition:   p.Condition().String(),
		Images:      p.Images(),
	}
}
