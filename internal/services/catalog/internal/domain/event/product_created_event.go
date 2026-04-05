package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

type ProductCreatedEvent struct {
	types.BaseEvent
	ProductPayload
}

type ProductPayload struct {
	ID           string                  `json:"id"`
	CategoryID   string                  `json:"category_id"`
	CategoryName string                  `json:"category_name"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	Brand        *string                 `json:"brand,omitempty"`
	Price        string                  `json:"price"`
	Currency     string                  `json:"currency"`
	Condition    string                  `json:"condition"`
	Status       string                  `json:"status"`
	Images       customtypes.Attachments `json:"images,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdateAt     time.Time               `json:"updated_at"`
}

func NewProductCreatedEvent(domainProduct *aggregate.Product, categoryName string) ProductCreatedEvent {
	return ProductCreatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.created", "catalog.events", uuid.NewString()),
		ProductPayload: ProductPayload{
			ID:           domainProduct.ID(),
			CategoryID:   domainProduct.CategoryID(),
			CategoryName: categoryName,
			Title:        domainProduct.Title(),
			Description:  domainProduct.Description(),
			Brand:        domainProduct.Brand(),
			Price:        domainProduct.Price().String(),
			Currency:     domainProduct.Currency().String(),
			Condition:    domainProduct.Condition().String(),
			Status:       domainProduct.Status().String(),
			Images:       domainProduct.Images(),
			CreatedAt:    domainProduct.CreatedAt(),
			UpdateAt:     domainProduct.UpdatedAt(),
		},
	}
}
