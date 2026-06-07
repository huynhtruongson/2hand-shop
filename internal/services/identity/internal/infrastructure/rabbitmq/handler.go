package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type typedHandler[T any] struct {
	fn dispatcher.TypedHandler[T]
}

func newTypedHandler[T any](fn dispatcher.TypedHandler[T]) dispatcher.Handler {
	return typedHandler[T]{fn: fn}
}

func (h typedHandler[T]) Handle(ctx context.Context, msg *types.DeliveryMessage) error {
	fmt.Printf("================msg: %v\n", string(msg.Raw().Body))
	var payload T
	if err := json.Unmarshal(msg.Raw().Body, &payload); err != nil {
		return fmt.Errorf("message: decode envelope (type=%s): %w", msg.Raw().Type, err)
	}
	env := types.EventEnvelope[T]{Payload: payload}
	meta := types.Metadata{
		Exchange:    msg.Raw().Exchange,
		RoutingKey:  msg.Raw().RoutingKey,
		DeliveryTag: msg.Raw().DeliveryTag,
	}
	return h.fn.Handle(ctx, types.NewEventContext(env, meta))
}
