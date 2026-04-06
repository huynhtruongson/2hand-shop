package dto

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// SearchProductsRequest is the HTTP request DTO for the search products endpoint.
type SearchProductsRequest struct {
	Keyword    string   `form:"keyword"`
	Category   *string  `form:"category"`
	Conditions []string `form:"conditions" binding:"omitempty,dive,oneof=new like_new good fair poor"`
	Page       int      `form:"page,default=1" binding:"min=1"`
	Limit      int      `form:"limit,default=20" binding:"min=1,max=100"`
	Sort       *string  `form:"sort" binding:"omitempty,oneof=created_at -created_at price -price"`
}

// ToSearchProductsQuery converts the HTTP request DTO to an application-layer query.
func (req SearchProductsRequest) ToSearchProductsQuery() query.SearchProductsQuery {
	return query.SearchProductsQuery{
		Page:       req.Page,
		Limit:      req.Limit,
		Keyword:    req.Keyword,
		Category:   req.Category,
		Conditions: req.Conditions,
		Sort:       req.Sort,
	}
}

// ToSearchProductsDTO converts a slice of domain Product aggregates to the HTTP response DTO.
func ToSearchProductsDTO(products []*aggregate.Product) []Product {
	if products == nil {
		return nil
	}
	out := make([]Product, len(products))
	for i, p := range products {
		out[i] = Product{
			ID:          p.ID(),
			CategoryID:  p.CategoryID(),
			Title:       p.Title(),
			Description: p.Description(),
			Brand:       p.Brand(),
			Price:       p.Price(),
			Currency:    p.Currency().String(),
			Condition:   p.Condition().String(),
			Images:      p.Images(),
			CreatedAt:   p.CreatedAt(),
			UpdatedAt:   p.UpdatedAt(),
		}
	}
	return out
}
