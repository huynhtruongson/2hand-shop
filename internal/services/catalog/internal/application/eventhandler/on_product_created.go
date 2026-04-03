package eventhandler

import (
	"context"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductCreatedHandler = dispatcher.TypedHandler[event.ProductCreatedEvent]

type onProductCreatedHandler struct {
	// TODO: inject read-side dependencies here (e.g., elasticsearch client, notification service)
}

// NewOnProductCreatedHandler constructs a new OnProductCreatedHandler.
func NewOnProductCreatedHandler() OnProductCreatedHandler {
	return &onProductCreatedHandler{}
}

//	processes the incoming product.created event.
//
// ec provides access to the typed payload, correlation_id, timestamp, and RabbitMQ metadata.
func (h *onProductCreatedHandler) Handle(ctx context.Context, ec types.EventContext[event.ProductCreatedEvent]) error {
	payload := ec.Payload()
	fmt.Printf("=========ID=%v", payload.ID)
	fmt.Printf("=========Title=%v", payload.Title)

	return nil
}
