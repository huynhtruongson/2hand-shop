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

	// Delete soft-deletes a ProductRequest by setting its deleted_at timestamp.
	// Implementations must return ErrProductRequestNotFound if no request matches.
	Delete(ctx context.Context, q postgressqlx.Querier, id string) error

	// List returns product requests matching filter, paginated, plus the total count.
	// Must only return non-deleted records (deleted_at IS NULL).
	// If filter.SellerID is non-nil the query is automatically scoped to that seller.
	List(ctx context.Context, q postgressqlx.Querier, filter ListProductRequestsFilter, page postgressqlx.Page) ([]*aggregate.ProductRequest, int, error)
}

// ListProductRequestsFilter carries optional filters for a paginated product-request listing.
type ListProductRequestsFilter struct {
	Category   *string  // filters by category slug via JOIN — nil means no filter
	SellerID   *string  // forces the query to a single seller's requests; nil means any seller
	Statuses   []string // e.g. []string{"pending", "approved"} — empty means all
	Conditions []string // e.g. []string{"new", "like_new"} — empty means all
	Sort       *string  // "created_at", "-created_at", "expected_price", "-expected_price"
}
