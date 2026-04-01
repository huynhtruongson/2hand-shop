package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type ProductCreatedEvent struct {
	types.BaseEvent
	ID           string    `json:"id"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Brand        *string   `json:"brand,omitempty"`
	Price        string    `json:"price"`
	Condition    string    `json:"condition"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdateAt     time.Time `json:"updated_at"`
}

func NewProductCreatedEvent() ProductCreatedEvent {
	return ProductCreatedEvent{
		BaseEvent: types.NewBaseEvent("catalog.product.created", "catalog.events", uuid.NewString()),
	}
}
