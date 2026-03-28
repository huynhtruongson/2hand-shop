package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/entity"
)

// CategoryRepository defines the persistence contract for Category entities.
type CategoryRepository interface {
	// Save persists a new Category.
	// Implementations must return ErrCategoryAlreadyExists if the ID already exists.
	Save(ctx context.Context, q postgressqlx.Querier, cat *entity.Category) error

	// Update updates an existing Category.
	// Implementations must return ErrCategoryNotFound if the category does not exist.
	Update(ctx context.Context, q postgressqlx.Querier, cat *entity.Category) error

	// Delete removes a Category by ID.
	// Implementations must return ErrCategoryNotFound if the category does not exist.
	Delete(ctx context.Context, q postgressqlx.Querier, categoryID string) error

	// GetByID retrieves a category by its ID.
	// Implementations must return ErrCategoryNotFound if no category matches.
	GetByID(ctx context.Context, q postgressqlx.Querier, categoryID string) (*entity.Category, error)

	// GetBySlug retrieves a category by its slug.
	// Implementations must return ErrCategoryNotFound if no category matches.
	GetBySlug(ctx context.Context, q postgressqlx.Querier, slug string) (*entity.Category, error)

	// ListAll retrieves all active categories ordered by sort_order.
	ListAll(ctx context.Context, q postgressqlx.Querier) ([]*entity.Category, error)
}
