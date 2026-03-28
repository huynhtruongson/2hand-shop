package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

// ProductRequestRepository defines the write-side persistence contract for ProductRequest aggregates.
// The caller supplies a postgressqlx.Querier so that callers in the application layer
// retain control over transactions.
type ProductRequestRepository interface {
	// Save persists a new ProductRequest to the write database.
	// Implementations must return ErrProductRequestAlreadyExists if the ID already exists.
	Save(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error

	// Update updates an existing ProductRequest in the write database.
	// Implementations must return ErrProductRequestNotFound if the request does not exist.
	Update(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error

	// GetByID retrieves a ProductRequest by its aggregate ID.
	// Implementations must return ErrProductRequestNotFound if no request matches.
	GetByID(ctx context.Context, q postgressqlx.Querier, id string) (*aggregate.ProductRequest, error)

	// ListBySellerID retrieves all ProductRequests belonging to a given seller.
	ListBySellerID(ctx context.Context, q postgressqlx.Querier, sellerID string) ([]*aggregate.ProductRequest, error)
}
