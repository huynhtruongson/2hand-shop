package event

import (
	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

// ProductRequestDeletedEvent is published when a seller deletes their product request.
type ProductRequestDeletedEvent struct {
	types.BaseEvent
	ProductRequestID string `json:"product_request_id"`
}

// NewProductRequestDeletedEvent constructs a ProductRequestDeletedEvent.
func NewProductRequestDeletedEvent(id string) ProductRequestDeletedEvent {
	return ProductRequestDeletedEvent{
		BaseEvent:        types.NewBaseEvent("catalog.product_request.deleted", "catalog.events", uuid.NewString()),
		ProductRequestID: id,
	}
}
