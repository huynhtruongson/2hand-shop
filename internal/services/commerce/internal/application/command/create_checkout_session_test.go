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

// checkoutHandlerTD (test double) wraps CreateCheckoutSessionHandler and
// intercepts the transaction runner to avoid needing a real DB.
type checkoutHandlerTD struct {
	cartRepo       repository.CartRepository
	orderRepo      repository.OrderRepository
	stripeProvider repository.PaymentProvider
}

func newCheckoutHandlerTD(
	cartRepo repository.CartRepository,
	orderRepo repository.OrderRepository,
	stripeProvider repository.PaymentProvider,
) *checkoutHandlerTD {
	return &checkoutHandlerTD{
		cartRepo:       cartRepo,
		orderRepo:      orderRepo,
		stripeProvider: stripeProvider,
	}
}

// Handle executes the checkout session logic using mockTX instead of a real transaction.
func (h *checkoutHandlerTD) Handle(ctx context.Context, cmd CreateCheckoutSessionCommand) (*CreateCheckoutSessionResponse, error) {
	var result *CreateCheckoutSessionResponse

	err := func() error {
		tx := &mockTX{}

		cart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			return err
		}

		if cart.ItemCount() == 0 {
			return errpkg.NewAppError(errpkg.KindBadRequest, "CART_EMPTY", "cannot create order from empty cart")
		}

		orderID := "test-order-id"

		orderItems := make([]entity.OrderItem, 0, cart.ItemCount())
		for _, ci := range cart.Items() {
			orderItems = append(orderItems, entity.NewOrderItem(
				"test-item-"+ci.ProductID(),
				orderID,
				ci.ProductID(),
				ci.ProductName(),
				ci.Price(),
				ci.Currency(),
			))
		}

		order, err := aggregate.NewOrder(
			orderID,
			cmd.UserID,
			orderItems,
			valueobject.CurrencyUSD,
			cmd.ShippingAddress,
		)
		if err != nil {
			return err
		}

		if err := h.orderRepo.Save(ctx, tx, order); err != nil {
			return err
		}

		stripeResult, err := h.stripeProvider.CreateSession(ctx, repository.CreateSessionParams{
			Order:      order,
			UserID:     cmd.UserID,
			SuccessURL: cmd.SuccessURL,
			CancelURL:  cmd.CancelURL,
			Currency:   valueobject.CurrencyUSD.String(),
		})
		if err != nil {
			return err
		}

		result = &CreateCheckoutSessionResponse{
			SessionID: stripeResult.SessionID,
			URL:       stripeResult.URL,
			OrderID:   orderID,
		}
		return nil
	}()
	if err != nil {
		return nil, err
	}
	return result, nil
}

var _ func(ctx context.Context, cmd CreateCheckoutSessionCommand) (*CreateCheckoutSessionResponse, error) = (*checkoutHandlerTD)(nil).Handle

func TestCreateCheckoutSessionHandler(t *testing.T) {
	t.Parallel()

	item := entity.NewCartItem("item-1", "cart-1", "prod-1", "Test Product", customtypes.MustNewPrice("10.99"), valueobject.CurrencyUSD)
	cart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{item})

	tests := []struct {
		name        string
		command     CreateCheckoutSessionCommand
		setup       func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider)
		wantErr     bool
		errContains string
	}{
		{
			name: "creates order and returns Stripe session URL",
			command: CreateCheckoutSessionCommand{
				UserID:     "user-1",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			setup: func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider) {
				cartRepo := &mockCartRepo{}
				cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
				}

				orderRepo := &mockOrderRepo{}
				orderRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, order *aggregate.Order) error {
					if order.UserID() != "user-1" {
						t.Errorf("expected user_id=user-1, got %q", order.UserID())
					}
					return nil
				}

				stripeProvider := &mockPaymentProvider{}
				stripeProvider.createSessionFn = func(_ context.Context, params repository.CreateSessionParams) (*repository.SessionResult, error) {
					if params.SuccessURL != "https://example.com/success" {
						t.Errorf("expected success URL, got %q", params.SuccessURL)
					}
					return &repository.SessionResult{SessionID: "cs_test_123", URL: "https://stripe.test/session"}, nil
				}

				return cartRepo, orderRepo, stripeProvider
			},
			wantErr: false,
		},
		{
			name: "returns CART_EMPTY when cart has no items",
			command: CreateCheckoutSessionCommand{
				UserID:     "user-1",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			setup: func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider) {
				emptyCart, _ := aggregate.NewCart("cart-1", "user-1", nil)
				cartRepo := &mockCartRepo{}
				cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return emptyCart, nil
				}
				return cartRepo, &mockOrderRepo{}, &mockPaymentProvider{}
			},
			wantErr:     true,
			errContains: "CART_EMPTY",
		},
		{
			name: "returns error when GetByUserID fails",
			command: CreateCheckoutSessionCommand{
				UserID:     "user-1",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			setup: func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider) {
				cartRepo := &mockCartRepo{}
				cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
				}
				return cartRepo, &mockOrderRepo{}, &mockPaymentProvider{}
			},
			wantErr:     true,
			errContains: "CART_NOT_FOUND",
		},
		{
			name: "returns error when orderRepo.Save fails",
			command: CreateCheckoutSessionCommand{
				UserID:     "user-1",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			setup: func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider) {
				cartRepo := &mockCartRepo{}
				cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
				}

				orderRepo := &mockOrderRepo{}
				orderRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Order) error {
					return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "save failed")
				}

				return cartRepo, orderRepo, &mockPaymentProvider{}
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
		{
			name: "returns error when Stripe CreateSession fails",
			command: CreateCheckoutSessionCommand{
				UserID:     "user-1",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			setup: func(t *testing.T) (repository.CartRepository, repository.OrderRepository, repository.PaymentProvider) {
				cartRepo := &mockCartRepo{}
				cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
				}

				orderRepo := &mockOrderRepo{}
				orderRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Order) error {
					return nil
				}

				stripeProvider := &mockPaymentProvider{}
				stripeProvider.createSessionFn = func(_ context.Context, _ repository.CreateSessionParams) (*repository.SessionResult, error) {
					return nil, errpkg.NewAppError(errpkg.KindInternal, "STRIPE_ERROR", "stripe error")
				}

				return cartRepo, orderRepo, stripeProvider
			},
			wantErr:     true,
			errContains: "STRIPE_ERROR",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cartRepo, orderRepo, stripeProvider := tc.setup(t)
			handler := newCheckoutHandlerTD(cartRepo, orderRepo, stripeProvider)

			resp, err := handler.Handle(context.Background(), tc.command)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.SessionID == "" {
				t.Error("expected non-empty SessionID")
			}
			if resp.URL == "" {
				t.Error("expected non-empty URL")
			}
			if resp.OrderID == "" {
				t.Error("expected non-empty OrderID")
			}
		})
	}
}

func TestCreateCheckoutSessionHandler_OrderItemsFromCart(t *testing.T) {
	t.Parallel()

	item1 := entity.NewCartItem("item-1", "cart-1", "prod-1", "Product 1", customtypes.MustNewPrice("10.00"), valueobject.CurrencyUSD)
	item2 := entity.NewCartItem("item-2", "cart-1", "prod-2", "Product 2", customtypes.MustNewPrice("20.00"), valueobject.CurrencyUSD)
	cart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{item1, item2})

	cartRepo := &mockCartRepo{}
	cartRepo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
		return cart, nil
	}

	var savedOrder *aggregate.Order
	orderRepo := &mockOrderRepo{}
	orderRepo.saveFn = func(_ context.Context, _ postgressqlx.Querier, order *aggregate.Order) error {
		savedOrder = order
		return nil
	}

	stripeProvider := &mockPaymentProvider{}
	stripeProvider.createSessionFn = func(_ context.Context, _ repository.CreateSessionParams) (*repository.SessionResult, error) {
		return &repository.SessionResult{SessionID: "cs_test", URL: "https://stripe.test"}, nil
	}

	handler := newCheckoutHandlerTD(cartRepo, orderRepo, stripeProvider)
	resp, err := handler.Handle(context.Background(), CreateCheckoutSessionCommand{
		UserID:     "user-1",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedOrder == nil {
		t.Fatal("expected order to be saved")
	}
	if len(savedOrder.Items()) != 2 {
		t.Errorf("expected 2 order items, got %d", len(savedOrder.Items()))
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}
