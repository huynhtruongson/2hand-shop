package command

import (
	"context"

	"github.com/google/uuid"

	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/event"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	commercevo "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// CompleteCheckoutCommand is issued by the Stripe webhook handler when
// a checkout.session.completed event is received.
type CompleteCheckoutCommand struct {
	StripeSessionID string
	StripeEventID   string
	OrderID         string
	UserID          string
	AmountCents     int64
	Currency        string
}

// CompleteCheckoutHandler handles the post-payment completion flow:
// - Confirms the pre-created pending order
// - Creates and confirms the Payment aggregate
// - Clears the user's cart
// - Publishes a checkout.completed domain event.
type CompleteCheckoutHandler struct {
	db          postgressqlx.DB
	cartRepo    repository.CartRepository
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
	publisher   publisher
	logger      logger.Logger
}

// NewCompleteCheckoutHandler returns a new CompleteCheckoutHandler.
func NewCompleteCheckoutHandler(
	cartRepo repository.CartRepository,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
	db postgressqlx.DB,
	pub publisher,
	log logger.Logger,
) *CompleteCheckoutHandler {
	return &CompleteCheckoutHandler{
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		db:          db,
		publisher:   pub,
		logger:      log,
	}
}

// Handle executes the CompleteCheckoutCommand inside a database transaction.
// It is idempotent: if the payment already exists, it returns nil without error.
// After the transaction commits, a checkout.completed event is published to RabbitMQ.
func (h *CompleteCheckoutHandler) Handle(ctx context.Context, cmd CompleteCheckoutCommand) error {

	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		// 1. Idempotency check: skip if payment already exists.
		existing, err := h.paymentRepo.GetByStripeSessionID(ctx, tx, cmd.StripeSessionID)
		if err == nil && existing != nil {
			// Payment already created — ack the webhook without error.
			return nil
		}
		// Only proceed if the error was "not found"; bubble up unexpected DB errors.
		if !errpkg.IsKind(err, errpkg.KindNotFound) {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: GetByStripeSessionID")
		}

		// 2. Load the pre-created pending order.
		order, err := h.orderRepo.GetByID(ctx, tx, cmd.OrderID)
		if err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: OrderRepo.GetByID")
		}

		// 3. Confirm the order (pending → confirmed).
		if err := order.Confirm(); err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: Order.Confirm")
		}

		// 4. Persist confirmed order.
		if err := h.orderRepo.Update(ctx, tx, order); err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: OrderRepo.Update")
		}

		// 5. Create and confirm the payment.
		currency, err := commercevo.NewCurrencyFromString(cmd.Currency)
		if err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: NewCurrencyFromString")
		}

		payment, err := aggregate.NewPayment(
			uuid.New().String(),
			order.ID(),
			cmd.StripeSessionID,
			order.TotalAmount(),
			currency,
		)
		if err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: NewPayment")
		}

		if err := payment.Confirm(); err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: Payment.Confirm")
		}

		if err := h.paymentRepo.Save(ctx, tx, payment); err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: PaymentRepo.Save")
		}

		// 6. Clear the user's cart.
		if err := h.cartRepo.Delete(ctx, tx, cmd.UserID); err != nil {
			return carterrors.ErrInternal.WithCause(err).WithInternal("CompleteCheckoutHandler: CartRepo.Delete")
		}

		productIDs := make([]string, len(order.Items()))
		for i, item := range order.Items() {
			productIDs[i] = item.ProductID()
		}
		if err := h.publisher.PublishMessage(ctx, event.NewCheckoutCompleted(productIDs)); err != nil {
			h.logger.Error("failed to publish checkout.completed",
				"error", err,
				"order_id", order.ID(),
			)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
