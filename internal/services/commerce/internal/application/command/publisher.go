package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

// publisher is implemented by the RabbitMQ producer and used to publish domain events.
type publisher interface {
	PublishMessage(ctx context.Context, message types.DomainEvent, opts ...types.MessageOption) error
}