package dto

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

type ListProductRequest struct {
	Category   *string  `form:"category"`
	Conditions []string `form:"conditions" binding:"omitempty,dive,oneof=new like_new good fair poor"`
	Statuses   []string `form:"statuses" binding:"omitempty,dive,oneof=draft published sold archived"`
	Page       int      `form:"page,default=1" binding:"min=1"`
	Limit      int      `form:"limit,default=20" binding:"min=1,max=100"`
	Sort       *string  `form:"sort" binding:"omitempty,oneof=created_at -created_at price -price"`
}

func (req ListProductRequest) ToListProductQuery() query.ListProductQuery {
	return query.ListProductQuery{
		Page:       req.Page,
		Limit:      req.Limit,
		Category:   req.Category,
		Conditions: req.Conditions,
		Statuses:   req.Statuses,
		Sort:       req.Sort,
	}
}

type Product struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	CategoryID  string                  `json:"category_id"`
	Description string                  `json:"description"`
	Brand       *string                 `json:"brand,omitempty"`
	Price       customtypes.Price       `json:"price"`
	Currency    string                  `json:"currency"`
	Condition   string                  `json:"condition"`
	Images      customtypes.Attachments `json:"images"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

func ToProductsDTO(domainProducts []aggregate.Product) []Product {
	products := make([]Product, len(domainProducts))
	for i, p := range domainProducts {
		products[i] = Product{
			ID:          p.ID(),
			Title:       p.Title(),
			CategoryID:  p.CategoryID(),
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
	return products
}
