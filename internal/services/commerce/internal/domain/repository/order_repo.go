package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
)

// OrderRepository defines the persistence contract for Order aggregates.
type OrderRepository interface {
	// Save persists a new Order to the write database.
	Save(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error

	// Update updates an existing Order in the write database.
	Update(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error

	// GetByID retrieves an order by its aggregate ID.
	GetByID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Order, error)

	// List returns orders matching the given filter and pagination, plus the total count.
	List(ctx context.Context, q postgressqlx.Querier, filter ListOrdersFilter, page postgressqlx.Page) ([]aggregate.Order, int, error)
}

type ListOrdersFilter struct {
	BuyerID *string
	Statuses []string
}