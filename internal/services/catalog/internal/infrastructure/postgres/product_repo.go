package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

type productModel struct {
	ID          string                  `db:"id"`
	CategoryID  string                  `db:"category_id"`
	Title       string                  `db:"title"`
	Description string                  `db:"description"`
	Brand       sql.NullString          `db:"brand"`
	Price       string                  `db:"price"` // TEXT — customtypes.Price.String()
	Currency    string                  `db:"currency"`
	Condition   string                  `db:"condition"`
	Status      string                  `db:"status"`
	Images      customtypes.Attachments `db:"images"` // JSONB
	CreatedAt   time.Time               `db:"created_at"`
	UpdatedAt   time.Time               `db:"updated_at"`
	DeletedAt   sql.NullTime            `db:"deleted_at"`
}

func toProductModel(p *aggregate.Product) *productModel {
	return &productModel{
		ID:          p.ID(),
		CategoryID:  p.CategoryID(),
		Title:       p.Title(),
		Description: p.Description(),
		Brand:       utils.StringPtrToNullString(p.Brand()),
		Price:       p.Price().String(),
		Currency:    p.Currency().String(),
		Condition:   p.Condition().String(),
		Status:      p.Status().String(),
		Images:      p.Images(),
		CreatedAt:   p.CreatedAt(),
		UpdatedAt:   p.UpdatedAt(),
		DeletedAt:   utils.TimePtrToNullTime(p.DeletedAt()),
	}
}

func (m productModel) toAggregate() (*aggregate.Product, error) {
	var price customtypes.Price
	if err := price.Scan(m.Price); err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productModel.toAggregate: scan price")
	}

	currency, err := valueobject.NewCurrencyFromString(m.Currency)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productModel.toAggregate: parse currency")
	}

	condition, err := valueobject.NewConditionFromString(m.Condition)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productModel.toAggregate: parse condition")
	}

	status, err := valueobject.NewProductStatusFromString(m.Status)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productModel.toAggregate: parse status")
	}

	return aggregate.UnmarshalProductFromDB(
		m.ID,
		m.CategoryID,
		m.Title,
		m.Description,
		price,
		currency,
		condition,
		status,
		m.Images,
		m.CreatedAt,
		m.UpdatedAt,
		utils.NullTimeToPtr(m.DeletedAt),
		utils.NullStringToStringPtr(m.Brand),
	), nil
}

// ProductRepo implements repository.ProductRepository using PostgreSQL.
type ProductRepo struct{}

func NewProductRepo() *ProductRepo {
	return &ProductRepo{}
}

var _ repository.ProductRepository = (*ProductRepo)(nil)

// Save persists a new Product. Returns ErrProductAlreadyExists on unique-constraint violation.
func (r *ProductRepo) Save(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	const query = `
		INSERT INTO products (id, category_id, title, description, brand, price, currency,
		                      condition, status, images, created_at, updated_at, deleted_at)
		VALUES (:id, :category_id, :title, :description, :brand, :price, :currency,
		        :condition, :status, :images, :created_at, :updated_at, :deleted_at)`

	_, err := q.NamedExecContext(ctx, query, toProductModel(product))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return caterrors.ErrProductAlreadyExists
		}
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.Save")
	}
	return nil
}

// Update persists changes to an existing Product. Returns ErrProductNotFound if no row matches.
func (r *ProductRepo) Update(ctx context.Context, q postgressqlx.Querier, product *aggregate.Product) error {
	const query = `
		UPDATE products
		SET category_id = $1, title = $2, description = $3, brand = $4,
		    price = $5, currency = $6, condition = $7, status = $8,
		    images = $9, updated_at = $10
		WHERE id = $11 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query,
		product.CategoryID(),
		product.Title(),
		product.Description(),
		utils.StringPtrToNullString(product.Brand()),
		product.Price().String(),
		product.Currency().String(),
		product.Condition().String(),
		product.Status().String(),
		product.Images(),
		product.UpdatedAt(),
		product.ID(),
	)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.Update")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.Update")
	}
	if rows == 0 {
		return caterrors.ErrProductNotFound
	}
	return nil
}

// Delete soft-deletes a Product. Returns ErrProductNotFound if no active product matches.
func (r *ProductRepo) Delete(ctx context.Context, q postgressqlx.Querier, productID string) error {
	now := time.Now().UTC()
	const query = `
		UPDATE products
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query, now, now, productID)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.Delete")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.Delete")
	}
	if rows == 0 {
		return caterrors.ErrProductNotFound
	}
	return nil
}

// GetByID retrieves a Product by its ID. Returns ErrProductNotFound if no row matches.
func (r *ProductRepo) GetByID(ctx context.Context, q postgressqlx.Querier, productID string) (*aggregate.Product, error) {
	const query = `
		SELECT id, category_id, title, description, brand, price, currency,
		       condition, status, images, created_at, updated_at, deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m productModel
	err := q.QueryRowxContext(ctx, query, productID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, caterrors.ErrProductNotFound
		}
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.GetByID")
	}

	product, err := m.toAggregate()
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.GetByID")
	}
	return product, nil
}

// List returns products matching the given filter and pagination, plus the total count.
func (r *ProductRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListProductsFilter, page postgressqlx.Page) ([]aggregate.Product, int, error) {
	var (
		args   []any
		join   string
		where  = "WHERE p.deleted_at IS NULL"
		argIdx = 1
	)
	if filter.Category != nil {
		join = " JOIN categories c ON c.id = p.category_id"
		where += " AND c.slug = $" + utils.Itoa(argIdx)
		args = append(args, *filter.Category)
		argIdx++
	}

	if len(filter.Conditions) > 0 {
		where += " AND p.condition = ANY($" + utils.Itoa(argIdx) + ")"
		args = append(args, pq.Array(filter.Conditions))
		argIdx++
	}

	if len(filter.Statuses) > 0 {
		where += " AND p.status = ANY($" + utils.Itoa(argIdx) + ")"
		args = append(args, pq.Array(filter.Statuses))
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM products p" + join + " " + where
	var total int
	if err := q.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.List: count")
	}

	orderBy := buildSortClause(filter.Sort)

	// Data query
	dataQuery := `
		SELECT p.id, p.category_id, p.title, p.description, p.brand, p.price, p.currency,
		       p.condition, p.status, p.images, p.created_at, p.updated_at, p.deleted_at
		FROM products p` + join + " " + where + `
		ORDER BY ` + orderBy + `
		` + page.SQL()

	var models []productModel
	if err := q.SelectContext(ctx, &models, dataQuery, args...); err != nil {
		return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.List: select")
	}

	products := make([]aggregate.Product, 0, len(models))
	for _, m := range models {
		p, err := m.toAggregate()
		if err != nil {
			return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRepo.List: toAggregate")
		}
		products = append(products, *p)
	}

	return products, total, nil
}

// buildSortClause returns a safe ORDER BY expression from a sort param.
// Accepts: "created_at", "-created_at", "price", "-price".
// Invalid/nil values default to "created_at DESC".
func buildSortClause(s *string) string {
	if s == nil || *s == "" {
		return "created_at DESC"
	}
	switch *s {
	case "created_at":
		return "created_at ASC"
	case "-created_at":
		return "created_at DESC"
	case "price":
		return "price::numeric ASC"
	case "-price":
		return "price::numeric DESC"
	default:
		return "created_at DESC"
	}
}
