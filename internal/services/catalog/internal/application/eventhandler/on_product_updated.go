package eventhandler

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductUpdatedHandler = dispatcher.TypedHandler[event.ProductPayload]

type onProductUpdatedHandler struct {
	logger         logger.Logger
	productIndexer productIndexer
}

func NewOnProductUpdatedHandler(logger logger.Logger, productIndexer productIndexer) OnProductUpdatedHandler {
	return &onProductUpdatedHandler{logger: logger, productIndexer: productIndexer}
}

func (h *onProductUpdatedHandler) Handle(ctx context.Context, ec types.EventContext[event.ProductPayload]) error {
	h.logger.Info("Processing product updated event", "payload", ec.Payload())
	payload := ec.Payload()

	return h.productIndexer.IndexProduct(ctx, payload)
}
