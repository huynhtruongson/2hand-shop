package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
)

// CartRepository defines the persistence contract for Cart aggregates.
type CartRepository interface {
	// GetByUserID retrieves a cart by user ID.
	GetByUserID(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error)

	// Save persists or updates a Cart.
	Save(ctx context.Context, q postgressqlx.Querier, cart *aggregate.Cart) error

	// Delete removes a cart by user ID.
	Delete(ctx context.Context, q postgressqlx.Querier, userID string) error
}