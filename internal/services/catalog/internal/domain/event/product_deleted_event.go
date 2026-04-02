package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type ProductDeletedEvent struct {
	types.BaseEvent
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}

func NewProductDeletedEvent() ProductDeletedEvent {
	return ProductDeletedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.deleted", "catalog.events", uuid.NewString()),
	}
}
