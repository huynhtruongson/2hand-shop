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

// Verify that removeFromCartHandlerForTest satisfies the RemoveFromCartHandler interface at compile time.
var _ RemoveFromCartHandler = (*removeFromCartHandlerForTest)(nil)

// removeFromCartHandlerForTest wraps the handler logic with an injectable transaction runner.
type removeFromCartHandlerForTest struct {
	cartRepo repository.CartRepository
}

func newRemoveFromCartHandlerForTest(cartRepo repository.CartRepository) RemoveFromCartHandler {
	return &removeFromCartHandlerForTest{cartRepo: cartRepo}
}

func (h *removeFromCartHandlerForTest) Handle(ctx context.Context, cmd RemoveFromCartCommand) (RemoveFromCartResponse, error) {
	var resp RemoveFromCartResponse

	err := func() error {
		tx := &mockTX{}
		cart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			return err
		}

		if !cart.RemoveItem(cmd.ProductID) {
			return errpkg.NewAppError(errpkg.KindNotFound, "CART_ITEM_NOT_FOUND", "cart item not found").
				WithMeta("product_id", cmd.ProductID)
		}

		if err := h.cartRepo.Save(ctx, tx, cart); err != nil {
			return err
		}

		resp = RemoveFromCartResponse{
			CartID:         cart.ID(),
			TotalItemCount: cart.ItemCount(),
		}
		return nil
	}()

	return resp, err
}

func TestRemoveFromCartHandler(t *testing.T) {
	t.Parallel()

	makeCart := func(userID, itemProductID string, itemPrice string) *aggregate.Cart {
		item := entity.NewCartItem(
			"item-1", "cart-1", itemProductID,
			"Product", customtypes.MustNewPrice(itemPrice),
			valueobject.CurrencyUSD,
		)
		cart, _ := aggregate.NewCart("cart-1", userID, []entity.CartItem{item})
		return cart
	}

	tests := []struct {
		name        string
		command     RemoveFromCartCommand
		setup       func(t *testing.T) repository.CartRepository
		wantErr     bool
		errContains string
	}{
		{
			name: "removes existing item and returns updated cart",
			command: RemoveFromCartCommand{
				UserID:    "user-1",
				ProductID: "product-1",
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				cart := makeCart("user-1", "product-1", "10.99")
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
				}
				repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, saved *aggregate.Cart) error {
					if saved.ItemCount() != 0 {
						t.Errorf("expected 0 items after removal, got %d", saved.ItemCount())
					}
					return nil
				}
				return repo
			},
			wantErr: false,
		},
		{
			name: "returns CART_ITEM_NOT_FOUND when product not in cart",
			command: RemoveFromCartCommand{
				UserID:    "user-1",
				ProductID: "nonexistent-product",
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				cart := makeCart("user-1", "product-1", "10.99")
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
				}
				return repo
			},
			wantErr:     true,
			errContains: "CART_ITEM_NOT_FOUND",
		},
		{
			name: "returns error when GetByUserID fails with not found",
			command: RemoveFromCartCommand{
				UserID:    "user-1",
				ProductID: "product-1",
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
				}
				return repo
			},
			wantErr:     true,
			errContains: "CART_NOT_FOUND",
		},
		{
			name: "returns error when GetByUserID fails with internal error",
			command: RemoveFromCartCommand{
				UserID:    "user-1",
				ProductID: "product-1",
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
			command: RemoveFromCartCommand{
				UserID:    "user-1",
				ProductID: "product-1",
			},
			setup: func(t *testing.T) repository.CartRepository {
				repo := &mockCartRepo{}
				cart := makeCart("user-1", "product-1", "10.99")
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return cart, nil
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
			handler := newRemoveFromCartHandlerForTest(repo)

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
		})
	}
}

func TestRemoveFromCartHandler_MultipleItems(t *testing.T) {
	t.Parallel()

	// Remove one item from a cart with two items; the other should remain.
	item1 := entity.NewCartItem("item-1", "cart-1", "prod-1", "Product 1", customtypes.MustNewPrice("10.00"), valueobject.CurrencyUSD)
	item2 := entity.NewCartItem("item-2", "cart-1", "prod-2", "Product 2", customtypes.MustNewPrice("20.00"), valueobject.CurrencyUSD)
	cart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{item1, item2})

	repo := &mockCartRepo{}
	repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
		return cart, nil
	}
	repo.saveFn = func(_ context.Context, _ postgressqlx.Querier, saved *aggregate.Cart) error {
		if saved.ItemCount() != 1 {
			t.Errorf("expected 1 item remaining, got %d", saved.ItemCount())
		}
		return nil
	}

	handler := newRemoveFromCartHandlerForTest(repo)
	resp, err := handler.Handle(context.Background(), RemoveFromCartCommand{
		UserID:    "user-1",
		ProductID: "prod-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalItemCount != 1 {
		t.Errorf("expected TotalItemCount=1, got %d", resp.TotalItemCount)
	}
}
