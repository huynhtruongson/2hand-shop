package application

import (
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/eventhandler"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application/query"
)

type Application struct {
	Commands      Commands
	Queries       Queries
	EventHandlers EventHandlers
}

type Commands struct {
	CreateProduct command.CreateProductHandler
	UpdateProduct command.UpdateProductHandler
	DeleteProduct command.DeleteProductHandler
}

type Queries struct {
	ListProduct query.ListProductHandler
}

// EventHandlers holds all event-driven (RabbitMQ consumer) handlers.
type EventHandlers struct {
	OnProductCreated eventhandler.OnProductCreatedHandler
	OnProductUpdated eventhandler.OnProductUpdatedHandler
	OnProductDeleted eventhandler.OnProductDeletedHandler
}
