package eventhandler

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
)

type OnProductRequestCreatedHandler = dispatcher.TypedHandler[event.ProductRequestPayload]

type onProductRequestCreatedHandler struct {
	logger logger.Logger
}

func NewOnProductRequestCreatedHandler(logger logger.Logger) OnProductRequestCreatedHandler {
	return &onProductRequestCreatedHandler{logger: logger}
}

func (h *onProductRequestCreatedHandler) Handle(ctx context.Context, ec types.EventContext[event.ProductRequestPayload]) error {
	h.logger.Info("Processing product request created event", "payload", ec.Payload())
	// TODO: index product request into Elasticsearch read model
	return nil
}
