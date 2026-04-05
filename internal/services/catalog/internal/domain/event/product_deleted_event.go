package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type ProductDeletedEvent struct {
	types.BaseEvent
	ProductID string `json:"product_id"`
}

func NewProductDeletedEvent(productID string) ProductDeletedEvent {
	return ProductDeletedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.deleted", "catalog.events", uuid.NewString()),
		ProductID: productID,
	}
}
