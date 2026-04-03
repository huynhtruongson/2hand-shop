package types

import (
	"context"
)

type EventHandler[T any] interface {
	Handle(ctx context.Context, ec EventContext[T]) error
}
