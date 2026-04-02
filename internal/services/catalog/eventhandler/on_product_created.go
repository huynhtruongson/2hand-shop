package eventhandler

import (
	"context"

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
		return nil
	}
	// TODO: implement — e.g., index product into Elasticsearch, send seller confirmation.
	_ = ec.CorrelationID()
	_ = ec.Timestamp()
	_ = ec.Metadata()
	_ = p
	return nil
}
