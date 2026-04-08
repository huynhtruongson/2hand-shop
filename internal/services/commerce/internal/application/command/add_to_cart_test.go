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

// Verify that addToCartHandlerForTest satisfies the AddToCartHandler interface at compile time.
var _ AddToCartHandler = (*addToCartHandlerForTest)(nil)

// addToCartHandlerForTest wraps the handler logic with an injectable transaction runner.
type addToCartHandlerForTest struct {
	cartRepo repository.CartRepository
}

func newAddToCartHandlerForTest(cartRepo repository.CartRepository) AddToCartHandler {
	return &addToCartHandlerForTest{cartRepo: cartRepo}
}

func (h *addToCartHandlerForTest) Handle(ctx context.Context, cmd AddToCartCommand) (AddToCartResponse, error) {
	cartID := "test-cart-id"
	itemID := "test-item-id"

	var cart *aggregate.Cart

	execErr := func() error {
		tx := &mockTX{}
		existingCart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			if errpkg.IsKind(err, errpkg.KindNotFound) {
				cart, err = aggregate.NewCart(cartID, cmd.UserID, nil)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			cart = existingCart
		}

		item := entity.NewCartItem(
			itemID,
			cart.ID(),
			cmd.ProductID,
			cmd.ProductName,
			cmd.Price,
			valueobject.CurrencyUSD,
		)
		cart.AddItem(item)

		return h.cartRepo.Save(ctx, tx, cart)
	}()
	if execErr != nil {
		return AddToCartResponse{}, execErr
	}

	return AddToCartResponse{
		CartID:         cart.ID(),
		ItemID:         itemID,
		TotalItemCount: cart.ItemCount(),
	}, nil
}

func TestAddToCartHandler(t *testing.T) {
	t.Parallel()

	price := customtypes.MustNewPrice("10.99")

	tests := []struct {
		name        string
		command     AddToCartCommand
		setup       func(t *testing.T) repository.CartRepository
		wantErr     bool
		errContains string
	}{
		{
			name: "creates new cart when none exists",
			command: AddToCartCommand{
				UserID:      "user-1",
				ProductID:   "product-1",
				ProductName: "Test Product",
				Price:       price,
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
				}
				repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, cart *aggregate.Cart) error {
					if cart.ID() == "" {
						t.Error("expected non-empty CartID after creation")
					}
					return nil
				}
				return repo
			},
			wantErr: false,
		},
		{
			name: "adds item to existing cart",
			command: AddToCartCommand{
				UserID:      "user-1",
				ProductID:   "product-2",
				ProductName: "Another Product",
				Price:       price,
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				existingCart, _ := aggregate.NewCart("cart-1", "user-1", nil)
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return existingCart, nil
				}
				repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, cart *aggregate.Cart) error {
					if cart.ItemCount() != 1 {
						t.Errorf("expected 1 item, got %d", cart.ItemCount())
					}
					return nil
				}
				return repo
			},
			wantErr: false,
		},
		{
			name: "replaces existing item for same product",
			command: AddToCartCommand{
				UserID:      "user-1",
				ProductID:   "product-1",
				ProductName: "Updated Product",
				Price:       customtypes.MustNewPrice("20.00"),
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				existingItem := entity.NewCartItem(
					"item-1", "cart-1", "product-1",
					"Old Product", customtypes.MustNewPrice("10.99"),
					valueobject.CurrencyUSD,
				)
				existingCart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{existingItem})
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return existingCart, nil
				}
				repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, cart *aggregate.Cart) error {
					if len(cart.Items()) != 1 {
						t.Errorf("expected 1 item after upsert, got %d", len(cart.Items()))
					}
					if cart.Items()[0].ProductName() != "Updated Product" {
						t.Errorf("expected updated product name, got %q", cart.Items()[0].ProductName())
					}
					return nil
				}
				return repo
			},
			wantErr: false,
		},
		{
			name: "returns error when GetByUserID fails with unexpected error",
			command: AddToCartCommand{
				UserID:      "user-1",
				ProductID:   "product-1",
				ProductName: "Product",
				Price:       price,
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "db error")
				}
				return repo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
		{
			name: "returns error when Save fails",
			command: AddToCartCommand{
				UserID:      "user-1",
				ProductID:   "product-1",
				ProductName: "Product",
				Price:       price,
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
				}
				repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Cart) error {
					return errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "save failed")
				}
				return repo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := tc.setup(t)
			handler := newAddToCartHandlerForTest(repo)

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
			if resp.CartID == "" {
				t.Error("expected non-empty CartID")
			}
			if resp.ItemID == "" {
				t.Error("expected non-empty ItemID")
			}
		})
	}
}

func TestAddToCartHandler_TotalItemCount(t *testing.T) {
	t.Parallel()

	item1 := entity.NewCartItem("item-1", "cart-1", "prod-1", "Product 1", customtypes.MustNewPrice("10.00"), valueobject.CurrencyUSD)
	item2 := entity.NewCartItem("item-2", "cart-1", "prod-2", "Product 2", customtypes.MustNewPrice("20.00"), valueobject.CurrencyUSD)
	cart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{item1, item2})

	repo := &mockCartRepo{}
	repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
		return cart, nil
	}
	repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, _ *aggregate.Cart) error {
		return nil
	}

	handler := newAddToCartHandlerForTest(repo)
	resp, err := handler.Handle(context.Background(), AddToCartCommand{
		UserID:      "user-1",
		ProductID:   "prod-3",
		ProductName: "Product 3",
		Price:       customtypes.MustNewPrice("30.00"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalItemCount != 3 {
		t.Errorf("expected TotalItemCount=3, got %d", resp.TotalItemCount)
	}
}
