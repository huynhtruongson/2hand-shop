package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// categoryModel mirrors the categories table schema.
type categoryModel struct {
	ID          string        `db:"id"`
	Name        string        `db:"name"`
	Description string        `db:"description"`
	Slug        string        `db:"slug"`
	IconURL     sql.NullString `db:"icon_url"`
	SortOrder   int           `db:"sort_order"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
	DeletedAt   sql.NullTime  `db:"deleted_at"`
}

// stringToNullString converts an empty string to a null SQL value.
func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// toCategoryModel converts a Category entity to a persistence model.
func toCategoryModel(c *entity.Category) *categoryModel {
	return &categoryModel{
		ID:          c.ID(),
		Name:        c.Name(),
		Description: c.Description(),
		Slug:        c.Slug(),
		IconURL:     stringToNullString(c.IconURL()),
		SortOrder:   c.SortOrder(),
		CreatedAt:   c.CreatedAt(),
		UpdatedAt:   c.UpdatedAt(),
		DeletedAt:   utils.TimePtrToNullTime(c.DeletedAt()),
	}
}

// toAggregate converts a persistence model back to a Category entity.
func (m categoryModel) toAggregate() *entity.Category {
	return entity.UnmarshalCategoryFromDB(
		m.ID,
		m.Name,
		m.Description,
		m.Slug,
		m.IconURL.String,
		m.SortOrder,
		m.CreatedAt,
		m.UpdatedAt,
		utils.NullTimeToPtr(m.DeletedAt),
	)
}

// CategoryRepo implements repository.CategoryRepository using PostgreSQL.
type CategoryRepo struct{}

func NewCategoryRepo() *CategoryRepo {
	return &CategoryRepo{}
}

var _ repository.CategoryRepository = (*CategoryRepo)(nil)

// Save persists a new Category. Returns ErrCategoryAlreadyExists on unique-constraint violation.
func (r *CategoryRepo) Save(ctx context.Context, q postgressqlx.Querier, cat *entity.Category) error {
	const query = `
		INSERT INTO categories (id, name, description, slug, icon_url, sort_order,
		                        created_at, updated_at, deleted_at)
		VALUES (:id, :name, :description, :slug, :icon_url, :sort_order,
		        :created_at, :updated_at, :deleted_at)`

	_, err := q.NamedExecContext(ctx, query, toCategoryModel(cat))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return caterrors.ErrCategoryAlreadyExists
		}
		return caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.Save")
	}
	return nil
}

// Update persists changes to an existing Category. Returns ErrCategoryNotFound if no row matches.
func (r *CategoryRepo) Update(ctx context.Context, q postgressqlx.Querier, cat *entity.Category) error {
	const query = `
		UPDATE categories
		SET name = $1, description = $2, icon_url = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query,
		cat.Name(),
		cat.Description(),
		stringToNullString(cat.IconURL()),
		cat.UpdatedAt(),
		cat.ID(),
	)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.Update")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.Update")
	}
	if rows == 0 {
		return caterrors.ErrCategoryNotFound
	}
	return nil
}

// Delete soft-deletes a Category. Returns ErrCategoryNotFound if no active category matches.
func (r *CategoryRepo) Delete(ctx context.Context, q postgressqlx.Querier, categoryID string) error {
	now := time.Now().UTC()
	const query = `
		UPDATE categories
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query, now, now, categoryID)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.Delete")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.Delete")
	}
	if rows == 0 {
		return caterrors.ErrCategoryNotFound
	}
	return nil
}

// GetByID retrieves a Category by its ID. Returns ErrCategoryNotFound if no row matches.
func (r *CategoryRepo) GetByID(ctx context.Context, q postgressqlx.Querier, categoryID string) (*entity.Category, error) {
	const query = `
		SELECT id, name, description, slug, icon_url, sort_order,
		       created_at, updated_at, deleted_at
		FROM categories
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m categoryModel
	err := q.QueryRowxContext(ctx, query, categoryID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, caterrors.ErrCategoryNotFound
		}
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.GetByID")
	}
	return m.toAggregate(), nil
}

// GetBySlug retrieves a Category by its slug. Returns ErrCategoryNotFound if no row matches.
func (r *CategoryRepo) GetBySlug(ctx context.Context, q postgressqlx.Querier, slug string) (*entity.Category, error) {
	const query = `
		SELECT id, name, description, slug, icon_url, sort_order,
		       created_at, updated_at, deleted_at
		FROM categories
		WHERE slug = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m categoryModel
	err := q.QueryRowxContext(ctx, query, slug).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, caterrors.ErrCategoryNotFound
		}
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.GetBySlug")
	}
	return m.toAggregate(), nil
}

// ListAll retrieves all active categories ordered by sort_order.
func (r *CategoryRepo) ListAll(ctx context.Context, q postgressqlx.Querier) ([]*entity.Category, error) {
	const query = `
		SELECT id, name, description, slug, icon_url, sort_order,
		       created_at, updated_at, deleted_at
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY sort_order ASC`

	var models []categoryModel
	err := q.SelectContext(ctx, &models, query)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("CategoryRepo.ListAll")
	}

	categories := make([]*entity.Category, 0, len(models))
	for _, m := range models {
		categories = append(categories, m.toAggregate())
	}
	return categories, nil
}
