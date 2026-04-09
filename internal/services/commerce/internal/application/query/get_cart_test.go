package query

import (
	"context"
	"strings"
	"testing"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// mockCartRepo is a test double for repository.CartRepository.
type mockGetCartRepo struct {
	getByUserIDFn func(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error)
}

func (m *mockGetCartRepo) GetByUserID(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, q, userID)
	}
	return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
}

func (m *mockGetCartRepo) Save(ctx context.Context, q postgressqlx.Querier, cart *aggregate.Cart) error {
	return nil
}

func (m *mockGetCartRepo) Delete(ctx context.Context, q postgressqlx.Querier, userID string) error {
	return nil
}

func TestGetCartHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		query       GetCartQuery
		setup       func(t *testing.T) *mockGetCartRepo
		wantErr     bool
		errContains string
		assertResp  func(t *testing.T, resp GetCartResponse)
	}{
		{
			name:  "returns existing cart when found",
			query: GetCartQuery{UserID: "user-1"},
			setup: func(t *testing.T) *mockGetCartRepo {
				repo := &mockGetCartRepo{}
				item := entity.NewCartItem("item-1", "cart-1", "prod-1", "Product 1",
					customtypes.MustNewPrice("10.00"), valueobject.CurrencyUSD)
				existingCart, _ := aggregate.NewCart("cart-1", "user-1", []entity.CartItem{item})
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, userID string) (*aggregate.Cart, error) {
					if userID != "user-1" {
						t.Errorf("expected userID=user-1, got %q", userID)
					}
					return existingCart, nil
				}
				return repo
			},
			wantErr: false,
			assertResp: func(t *testing.T, resp GetCartResponse) {
				if resp.Cart.ID() != "cart-1" {
					t.Errorf("expected cart ID cart-1, got %q", resp.Cart.ID())
				}
				if resp.Cart.UserID() != "user-1" {
					t.Errorf("expected user ID user-1, got %q", resp.Cart.UserID())
				}
				if resp.Cart.ItemCount() != 1 {
					t.Errorf("expected 1 item, got %d", resp.Cart.ItemCount())
				}
			},
		},
		{
			name:  "returns empty in-memory cart when not found",
			query: GetCartQuery{UserID: "user-2"},
			setup: func(t *testing.T) *mockGetCartRepo {
				repo := &mockGetCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
				}
				return repo
			},
			wantErr: false,
			assertResp: func(t *testing.T, resp GetCartResponse) {
				if resp.Cart.UserID() != "user-2" {
					t.Errorf("expected user ID user-2, got %q", resp.Cart.UserID())
				}
				if resp.Cart.ItemCount() != 0 {
					t.Errorf("expected 0 items for empty cart, got %d", resp.Cart.ItemCount())
				}
			},
		},
		{
			name:  "returns wrapped internal error when GetByUserID fails unexpectedly",
			query: GetCartQuery{UserID: "user-3"},
			setup: func(t *testing.T) *mockGetCartRepo {
				repo := &mockGetCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, errpkg.NewAppError(errpkg.KindInternal, "INTERNAL_ERROR", "db connection error")
				}
				return repo
			},
			wantErr:     true,
			errContains: "INTERNAL_ERROR",
		},
		{
			name:  "returns wrapped internal error when GetByUserID returns plain error",
			query: GetCartQuery{UserID: "user-4"},
			setup: func(t *testing.T) *mockGetCartRepo {
				repo := &mockGetCartRepo{}
				repo.getByUserIDFn = func(_ context.Context, _ postgressqlx.Querier, _ string) (*aggregate.Cart, error) {
					return nil, context.DeadlineExceeded
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
			handler := NewGetCartHandler(repo, nil)

			resp, err := handler.Handle(context.Background(), tc.query)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}


