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
	CreateProduct        command.CreateProductHandler
	DeleteProduct        command.DeleteProductHandler
	UpdateProduct        command.UpdateProductHandler
	PublishProduct       command.PublishProductHandler
	CreateProductRequest command.CreateProductRequestHandler
	UpdateProductRequest command.UpdateProductRequestHandler
	DeleteProductRequest command.DeleteProductRequestHandler
	AcceptProductRequest command.AcceptProductRequestHandler
	RejectProductRequest command.RejectProductRequestHandler
}

type Queries struct {
	SearchProducts    query.SearchProductsHandler
	ListProduct       query.ListProductHandler
	GetProduct        query.GetProductHandler
	ListProductRequests query.ListProductRequestsHandler
}

// EventHandlers holds all event-driven (RabbitMQ consumer) handlers.
type EventHandlers struct {
	OnProductCreated        eventhandler.OnProductCreatedHandler
	OnProductUpdated        eventhandler.OnProductUpdatedHandler
	OnProductDeleted        eventhandler.OnProductDeletedHandler
	OnProductRequestCreated eventhandler.OnProductRequestCreatedHandler
}
