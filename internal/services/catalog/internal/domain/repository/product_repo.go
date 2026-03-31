package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductRepository defines the write-side persistence contract for Product aggregates.
// The caller supplies a postgressqlx.Querier so that callers in the application layer
// retain control over transactions.

type ProductRepository interface {
	// Save persists a new Product to the write database.
	// Implementations must return ErrProductAlreadyExists if the ID already exists.
	Save(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error

	// Update updates an existing Product in the write database.
	// Implementations must return ErrProductNotFound if the product does not exist.
	Update(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error

	// Delete soft-deletes or hard-deletes a product by ID.
	// Implementations must return ErrProductNotFound if the product does not exist.
	Delete(ctx context.Context, q postgressqlx.Querier, productID string) error

	// GetByID retrieves a product by its aggregate ID.
	// Implementations must return ErrProductNotFound if no product matches.
	GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error)

	List(ctx context.Context, q postgressqlx.Querier, filter ListProductsFilter, page postgressqlx.Page) ([]aggregate.Product, int, error)
}

type ListProductsFilter struct {
	Category  *string  // nil means no category filter
	Condition *string  // nil means no condition filter
	Statuses  []string // e.g. []string{"published"}
	Sort      *string  // e.g. "price", "-price"
}
