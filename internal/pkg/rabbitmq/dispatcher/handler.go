package dispatcher

import (
	"context"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

// EventHandler is a generic function type for handling a fully-decoded domain event.
// T is the concrete event struct (e.g. ProductCreatedEvent, OrderPlacedEvent).
//
// The dispatcher decodes the message into Envelope[T] before calling fn,
// so the handler receives a strongly-typed event with no additional unmarshalling needed.
type EventHandler[T any] func(ctx context.Context, event T) error

// TypedHandler is the interface implemented by event handlers that can be registered
// with the dispatcher. Each handler is responsible for decoding the message body
// into its expected event type before processing.
//
// Implementations are typically created via NewTypedHandler.
type TypedHandler interface {
	// Handle processes the given delivery message. The implementation must be
	// safe to call concurrently from multiple goroutines.
	Handle(ctx context.Context, msg *types.DeliveryMessage) error
}

// typedHandler is a generic adapter that implements TypedHandler.
// It decodes the incoming DeliveryMessage into Envelope[T] and invokes
// the user-supplied EventHandler[T] on the decoded payload.
type typedHandler[T any] struct {
	fn EventHandler[T]
}

// NewTypedHandler wraps a generic EventHandler[T] into a TypedHandler.
// The returned adapter is safe to register under one or more routing keys.
func NewTypedHandler[T any](fn EventHandler[T]) TypedHandler {
	return &typedHandler[T]{fn: fn}
}

// Handle decodes the message body into Envelope[T] and invokes the wrapped handler.
func (h *typedHandler[T]) Handle(ctx context.Context, msg *types.DeliveryMessage) error {
	env, err := types.Decode[T](msg)
	if err != nil {
		return fmt.Errorf("dispatcher: decode event (type=%s): %w", msg.Metadata().Type, err)
	}
	return h.fn(ctx, env.Payload)
}
