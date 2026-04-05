package eventhandler

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductCreatedHandler = dispatcher.TypedHandler[event.ProductPayload]

type onProductCreatedHandler struct {
	logger         logger.Logger
	productIndexer productIndexer
}

type productIndexer interface {
	IndexProduct(ctx context.Context, payload event.ProductPayload) error
	DeleteProduct(ctx context.Context, productID string) error
}

func NewOnProductCreatedHandler(logger logger.Logger, productIndexer productIndexer) OnProductCreatedHandler {
	return &onProductCreatedHandler{logger: logger, productIndexer: productIndexer}
}

func (h *onProductCreatedHandler) Handle(ctx context.Context, ec types.EventContext[event.ProductPayload]) error {
	h.logger.Info("Processing product created event", "payload", ec.Payload())
	payload := ec.Payload()

	return h.productIndexer.IndexProduct(ctx, payload)
}
