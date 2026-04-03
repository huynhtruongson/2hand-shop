package dispatcher

import (
	"context"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

// Handler is the non-generic interface stored in the dispatcher's registry.
// Application-layer typed handlers satisfy this by implementing TypedHandler[T]
// (renaming their Handle method to Handle).
type Handler interface {
	Handle(ctx context.Context, msg *types.DeliveryMessage) error
}

// TypedHandler is implemented by application-layer typed event handlers.
// Each concrete handler (e.g. *OnProductCreatedHandler) satisfies this interface
// by providing a Handle method that receives the fully-typed event context.
type TypedHandler[T any] interface {
	Handle(ctx context.Context, ec types.EventContext[T]) error
}

// typedHandler[T] wraps a TypedHandler[T] and implements the non-generic Handler
// by decoding the delivery into the concrete type T before delegating.
type typedHandler[T any] struct {
	fn TypedHandler[T]
}

// NewTypedHandler wraps a TypedHandler[T] into a Handler suitable for registration
// with the EventDispatcher. T is inferred from fn's type.
func NewTypedHandler[T any](fn TypedHandler[T]) Handler {
	return typedHandler[T]{fn: fn}
}

func (h typedHandler[T]) Handle(ctx context.Context, msg *types.DeliveryMessage) error {
	env, meta, err := types.DecodeEnvelope[T](msg)
	if err != nil {
		return fmt.Errorf("dispatcher: decode event (type=%s): %w", msg.Metadata().Type, err)
	}
	return h.fn.Handle(ctx, types.NewEventContext(env, meta))
}
