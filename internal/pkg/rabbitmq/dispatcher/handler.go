package dispatcher

import (
	"context"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type Handler interface {
	Handle(ctx context.Context, msg *types.DeliveryMessage) error
}

type eventHandler struct {
	handler types.EventHandler
}

// NewEventHandler wraps an EventHandler function.
func NewEventHandler(handler types.EventHandler) eventHandler {
	return eventHandler{handler: handler}
}

// Handle decodes the delivery into EventEnvelope[T] and invokes the wrapped handler
// with the resulting EventContext.
func (h eventHandler) Handle(ctx context.Context, msg *types.DeliveryMessage) error {
	env, meta, err := types.DecodeEnvelope(msg)
	if err != nil {
		return fmt.Errorf("dispatcher: decode event (type=%s): %w", msg.Metadata().Type, err)
	}
	return h.handler.Handle(ctx, types.NewEventContext(env, meta))
}
