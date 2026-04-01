package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type ProductCreatedEvent struct {
	types.BaseEvent
}

func NewProductCreatedEvent() ProductCreatedEvent {
	return ProductCreatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.created", "catalog.events", uuid.NewString()),
	}
}
