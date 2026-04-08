package command

import (
	"context"
	"testing"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// completeCheckoutHandlerTestDouble shadows CompleteCheckoutHandler and overrides Handle.
type completeCheckoutHandlerTestDouble struct {
	cartRepo    repository.CartRepository
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
}

func newCompleteCheckoutHandlerTestDouble(
	cartRepo repository.CartRepository,
	orderRepo repository.OrderRepository,
	paymentRepo repository.PaymentRepository,
) *completeCheckoutHandlerTestDouble {
	return &completeCheckoutHandlerTestDouble{
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
	}
}

func (h *completeCheckoutHandlerTestDouble) Handle(ctx context.Context, cmd CompleteCheckoutCommand) error {
	tx := &mockTX{}

	// 1. Idempotency check
	existing, err := h.paymentRepo.GetByStripeSessionID(ctx, tx, cmd.StripeSessionID)
	if err == nil && existing != nil {
		return nil
	}
	if !errpkg.IsKind(err, errpkg.KindNotFound) {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "GetByStripeSessionID failed").
			WithCause(err)
	}

	// 2. Load order
	order, err := h.orderRepo.GetByID(ctx, tx, cmd.OrderID)
	if err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "OrderRepo.GetByID failed").
			WithCause(err)
	}

	// 3. Confirm order
	if err := order.Confirm(); err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "Order.Confirm failed").
			WithCause(err)
	}
	if err := h.orderRepo.Update(ctx, tx, order); err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "OrderRepo.Update failed").
			WithCause(err)
	}

	// 4. Create and confirm payment
	currency, err := valueobject.NewCurrencyFromString(cmd.Currency)
	if err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "NewCurrencyFromString failed").
			WithCause(err)
	}

	payment, err := aggregate.NewPayment(
		"test-payment-id",
		order.ID(),
		cmd.StripeSessionID,
		order.TotalAmount(),
		currency,
	)
	if err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "NewPayment failed").
			WithCause(err)
	}
	if err := payment.Confirm(); err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "Payment.Confirm failed").
			WithCause(err)
	}
	if err := h.paymentRepo.Save(ctx, tx, payment); err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "PaymentRepo.Save failed").
			WithCause(err)
	}

	// 5. Clear cart
	if err := h.cartRepo.Delete(ctx, tx, cmd.UserID); err != nil {
		return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "CartRepo.Delete failed").
			WithCause(err)
	}

	return nil
}

func makeOrderItem(id, orderID, productID, productName, price string) entity.OrderItem {
	return entity.NewOrderItem(
		id, orderID, productID, productName,
		customtypes.MustNewPrice(price),
		valueobject.CurrencyUSD,
	)
}

func TestCompleteCheckoutHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		command     CompleteCheckoutCommand
		setup       func(t *testing.T) (
			repository.CartRepository,
			repository.OrderRepository,
			repository.PaymentRepository,
		)
		wantErr     bool
		errContains string
	}{
		{
			name: "confirms order and creates payment successfully",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test_123",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				order, _ := aggregate.NewOrder(
					"order-1", "user-1",
					[]entity.OrderItem{makeOrderItem("item-1", "order-1", "prod-1", "Product", "10.99")},
					valueobject.CurrencyUSD, nil,
				)

				orderRepo := &mockOrderRepo{}
				orderRepo.getByIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Order, error) {
					return order, nil
				}
				orderRepo.updateFn = func(_ context.Context, _ postgressqlx.Querier, o *aggregate.Order) error {
					if o.Status() != valueobject.OrderStatusConfirmed {
						t.Errorf("expected confirmed status, got %v", o.Status())
					}
					return nil
				}

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
				}
				paymentRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Payment) error {
					return nil
				}

				cartRepo := &mockCartRepo{}
				cartRepo.deleteFn = func(_ context.Context, _ postgressqlx.Querier, _ string) error {
					return nil
				}

				return cartRepo, orderRepo, paymentRepo
			},
			wantErr: false,
		},
		{
			name: "is idempotent when payment already exists",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_existing",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				existingPayment, _ := aggregate.NewPayment(
					"existing-payment", "order-1", "cs_existing",
					customtypes.MustNewPrice("10.99"), valueobject.CurrencyUSD,
				)

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return existingPayment, nil
				}

				// Order and cart repos should NOT be called.
				return &mockCartRepo{}, &mockOrderRepo{}, paymentRepo
			},
			wantErr: false,
		},
		{
			name: "returns error when GetByStripeSessionID fails unexpectedly",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "db error")
				}
				return &mockCartRepo{}, &mockOrderRepo{}, paymentRepo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
		{
			name: "returns error when order not found",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test",
				OrderID:         "nonexistent-order",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				orderRepo := &mockOrderRepo{}
				orderRepo.getByIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Order, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "ORDER_NOT_FOUND", "order not found")
				}

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
				}

				return &mockCartRepo{}, orderRepo, paymentRepo
			},
			wantErr:     true,
			errContains: "ORDER_NOT_FOUND",
		},
		{
			name: "returns error when Order.Confirm fails due to invalid status transition",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				// Already-confirmed order — confirming again should fail.
				order, _ := aggregate.NewOrder(
					"order-1", "user-1",
					[]entity.OrderItem{makeOrderItem("item-1", "order-1", "prod-1", "Product", "10.99")},
					valueobject.CurrencyUSD, nil,
				)
				_ = order.Confirm() // Move to confirmed first

				orderRepo := &mockOrderRepo{}
				orderRepo.getByIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Order, error) {
					return order, nil
				}

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
				}

				return &mockCartRepo{}, orderRepo, paymentRepo
			},
			wantErr:     true,
			errContains: "ORDER_INVALID_STATUS_TRANSITION",
		},
		{
			name: "returns error when CartRepo.Delete fails",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "USD",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				order, _ := aggregate.NewOrder(
					"order-1", "user-1",
					[]entity.OrderItem{makeOrderItem("item-1", "order-1", "prod-1", "Product", "10.99")},
					valueobject.CurrencyUSD, nil,
				)

				orderRepo := &mockOrderRepo{}
				orderRepo.getByIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Order, error) {
					return order, nil
				}
				orderRepo.updateFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Order) error {
					return nil
				}

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
				}
				paymentRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Payment) error {
					return nil
				}

				cartRepo := &mockCartRepo{}
				cartRepo.deleteFn = func(_ context.Context, _ postgressqlx.Querier, _ string) error {
					return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "delete failed")
				}

				return cartRepo, orderRepo, paymentRepo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
		{
			name: "returns error for invalid currency",
			command: CompleteCheckoutCommand{
				StripeSessionID: "cs_test",
				OrderID:         "order-1",
				UserID:          "user-1",
				AmountCents:     1099,
				Currency:        "INVALID",
			},
			setup: func(t *testing.T) (
				repository.CartRepository,
				repository.OrderRepository,
				repository.PaymentRepository,
			) {
				order, _ := aggregate.NewOrder(
					"order-1", "user-1",
					[]entity.OrderItem{makeOrderItem("item-1", "order-1", "prod-1", "Product", "10.99")},
					valueobject.CurrencyUSD, nil,
				)

				orderRepo := &mockOrderRepo{}
				orderRepo.getByIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Order, error) {
					return order, nil
				}

				paymentRepo := &mockPaymentRepo{}
				paymentRepo.getByStripeSessionFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Payment, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
				}

				return &mockCartRepo{}, orderRepo, paymentRepo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cartRepo, orderRepo, paymentRepo := tc.setup(t)
			handler := newCompleteCheckoutHandlerTestDouble(cartRepo, orderRepo, paymentRepo)

			err := handler.Handle(context.Background(), tc.command)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
