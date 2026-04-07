package application

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/query"
)

type Application struct {
	Commands      Commands
	Queries       Queries
	EventHandlers EventHandlers
}

type Commands struct {
	AddToCart       command.AddToCartHandler
	RemoveFromCart  command.RemoveFromCartHandler
}

type Queries struct {
	GetCart query.GetCartHandler
}

// EventHandlers holds all event-driven (RabbitMQ consumer) handlers.
type EventHandlers struct {
}
