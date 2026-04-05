package eventhandler

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductDeletedHandler = dispatcher.TypedHandler[event.ProductDeletedEvent]

type onProductDeletedHandler struct {
	logger         logger.Logger
	productIndexer productIndexer
}

func NewOnProductDeletedHandler(logger logger.Logger, productIndexer productIndexer) OnProductDeletedHandler {
	return &onProductDeletedHandler{logger: logger, productIndexer: productIndexer}
}

func (h *onProductDeletedHandler) Handle(ctx context.Context, ec types.EventContext[event.ProductDeletedEvent]) error {
	h.logger.Info("Processing product deleted event", "payload", ec.Payload())
	payload := ec.Payload()

	return h.productIndexer.DeleteProduct(ctx, payload.ProductID)
}
