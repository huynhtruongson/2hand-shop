package eventhandler

import (
	"context"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductCreatedHandler types.EventHandler

type onProductCreatedHandler struct {
	// TODO: inject read-side dependencies here (e.g., elasticsearch client, notification service)
}

// NewOnProductCreatedHandler constructs a new OnProductCreatedHandler.
func NewOnProductCreatedHandler() *onProductCreatedHandler {
	return &onProductCreatedHandler{}
}

// Handle processes the incoming product.created event.
// ec provides access to the typed payload, correlation_id, timestamp, and RabbitMQ metadata.
func (h *onProductCreatedHandler) Handle(ctx context.Context, ec types.EventContext) error {
	p, ok := ec.Payload().(event.ProductCreatedEvent)
	if !ok {
		fmt.Println("========= not ProductCreatedEvent")
		return nil
	}
	fmt.Printf("========= event context: %+v\n", ec)
	fmt.Printf("========= payload: %+v\n", p)
	fmt.Printf("========= correlation_id: %+v\n", ec.CorrelationID())
	return nil
}
