package types

import (
	"context"
)

type EventHandler interface {
	Handle(ctx context.Context, ec EventContext) error
}
