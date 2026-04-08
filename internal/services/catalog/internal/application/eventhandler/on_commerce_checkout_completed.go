package eventhandler

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// OnCheckoutCompletedHandler consumes commerce.checkout.completed and marks
// all products in the payload as sold.
type OnCheckoutCompletedHandler = dispatcher.TypedHandler[CheckoutCompletedPayload]

type CheckoutCompletedPayload struct {
	ProductIDs []string `json:"product_ids"`
}

type onCheckoutCompletedHandler struct {
	logger      logger.Logger
	productRepo repository.ProductRepository
	db          postgressqlx.DB
}

// NewOnCheckoutCompletedHandler constructs the handler.
func NewOnCheckoutCompletedHandler(
	logger logger.Logger,
	productRepo repository.ProductRepository,
	db postgressqlx.DB,
) OnCheckoutCompletedHandler {
	return &onCheckoutCompletedHandler{
		logger:      logger,
		productRepo: productRepo,
		db:          db,
	}
}

// Handle marks each product in ProductIDs as sold.
func (h *onCheckoutCompletedHandler) Handle(ctx context.Context, ec types.EventContext[CheckoutCompletedPayload]) error {
	payload := ec.Payload()
	if len(payload.ProductIDs) == 0 {
		return nil
	}

	h.logger.Info("Processing checkout completed event", "product_ids", payload.ProductIDs)

	return postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		for _, productID := range payload.ProductIDs {
			product, err := h.productRepo.GetByID(ctx, tx, productID)
			if err != nil {
				h.logger.Warn("failed to get product for MarkSold", "product_id", productID, "err", err)
				continue
			}

			// orderID intentionally not stored — extend when commerce.checkout.completed carries it
			if err := product.MarkSold(); err != nil {
				h.logger.Warn("failed to mark product sold", "product_id", productID, "err", err)
				continue
			}

			if err := h.productRepo.Update(ctx, tx, product); err != nil {
				return err
			}
		}
		return nil
	})
}
