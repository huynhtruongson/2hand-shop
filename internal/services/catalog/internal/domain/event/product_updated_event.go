package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

type ProductUpdatedEvent struct {
	types.BaseEvent
	ProductPayload
}

func NewProductUpdatedEvent(domainProduct *aggregate.Product, categoryName string) ProductUpdatedEvent {
	return ProductUpdatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.updated", "catalog.events", uuid.NewString()),
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
