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

type productRequestModel struct {
	ID                string                  `db:"id"`
	SellerID          string                  `db:"seller_id"`
	CategoryID        string                  `db:"category_id"`
	Title             string                  `db:"title"`
	Description       string                  `db:"description"`
	Brand             sql.NullString          `db:"brand"`
	ExpectedPrice     customtypes.Price       `db:"expected_price"` // TEXT — customtypes.Price.String()
	Currency          string                  `db:"currency"`
	Condition         string                  `db:"condition"`
	Status            string                  `db:"status"`
	Images            customtypes.Attachments `db:"images"` // JSONB
	ContactInfo       string                  `db:"contact_info"`
	AdminRejectReason *string                 `db:"admin_reject_reason"`
	AdminNote         *string                 `db:"admin_note"`
	CreatedAt         time.Time               `db:"created_at"`
	UpdatedAt         time.Time               `db:"updated_at"`
	DeletedAt         sql.NullTime            `db:"deleted_at"`
}

func toProductRequestModel(pr *aggregate.ProductRequest) *productRequestModel {
	return &productRequestModel{
		ID:                pr.ID(),
		SellerID:          pr.SellerID(),
		CategoryID:        pr.CategoryID(),
		Title:             pr.Title(),
		Description:       pr.Description(),
		Brand:             utils.StringPtrToNullString(pr.Brand()),
		ExpectedPrice:     pr.ExpectedPrice(),
		Currency:          pr.Currency().String(),
		Condition:         pr.Condition().String(),
		Status:            pr.Status().String(),
		Images:            pr.Images(),
		ContactInfo:       pr.ContactInfo(),
		AdminRejectReason: pr.AdminRejectReason(),
		AdminNote:         pr.AdminNote(),
		CreatedAt:         pr.CreatedAt(),
		UpdatedAt:         pr.UpdatedAt(),
		DeletedAt:         utils.TimePtrToNullTime(pr.DeletedAt()),
	}
}

func (m productRequestModel) toAggregate() (*aggregate.ProductRequest, error) {
	currency, err := valueobject.NewCurrencyFromString(m.Currency)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productRequestModel.toAggregate: parse currency")
	}

	condition, err := valueobject.NewConditionFromString(m.Condition)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productRequestModel.toAggregate: parse condition")
	}

	status, err := valueobject.NewProductRequestStatusFromString(m.Status)
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("productRequestModel.toAggregate: parse status")
	}

	return aggregate.UnmarshalProductRequestFromDB(
		m.ID,
		m.SellerID,
		m.CategoryID,
		m.Title,
		m.Description,
		m.ExpectedPrice,
		currency,
		utils.NullStringToStringPtr(m.Brand),
		condition,
		status,
		m.Images,
		m.ContactInfo,
		m.AdminRejectReason,
		m.AdminNote,
		m.CreatedAt,
		m.UpdatedAt,
		utils.NullTimeToPtr(m.DeletedAt),
	), nil
}

// ProductRequestRepo implements repository.ProductRequestRepository using PostgreSQL.
type ProductRequestRepo struct{}

func NewProductRequestRepo() *ProductRequestRepo {
	return &ProductRequestRepo{}
}

var _ repository.ProductRequestRepository = (*ProductRequestRepo)(nil)

// Save persists a new ProductRequest. Returns ErrProductRequestAlreadyExists on unique-constraint violation.
func (r *ProductRequestRepo) Save(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	const query = `
		INSERT INTO product_requests (id, seller_id, category_id, title, description, brand,
		                               expected_price, currency, condition, status, images,
		                               contact_info, admin_reject_reason, admin_note,
		                               created_at, updated_at, deleted_at)
		VALUES (:id, :seller_id, :category_id, :title, :description, :brand,
		        :expected_price, :currency, :condition, :status, :images,
		        :contact_info, :admin_reject_reason, :admin_note,
		        :created_at, :updated_at, :deleted_at)`

	_, err := q.NamedExecContext(ctx, query, toProductRequestModel(pr))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return caterrors.ErrProductRequestAlreadyExists
		}
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.Save")
	}
	return nil
}

// Update persists changes to an existing ProductRequest. Returns ErrProductRequestNotFound if no row matches.
func (r *ProductRequestRepo) Update(ctx context.Context, q postgressqlx.Querier, pr *aggregate.ProductRequest) error {
	const query = `
		UPDATE product_requests
		SET category_id = $1, title = $2, description = $3, brand = $4,
		    expected_price = $5, currency = $6, condition = $7, status = $8,
		    images = $9, contact_info = $10, admin_reject_reason = $11,
		    admin_note = $12, updated_at = $13, deleted_at = $14
		WHERE id = $15 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query,
		pr.CategoryID(),
		pr.Title(),
		pr.Description(),
		utils.StringPtrToNullString(pr.Brand()),
		pr.ExpectedPrice().String(),
		pr.Currency().String(),
		pr.Condition().String(),
		pr.Status().String(),
		pr.Images(),
		pr.ContactInfo(),
		pr.AdminRejectReason(),
		pr.AdminNote(),
		pr.UpdatedAt(),
		utils.TimePtrToNullTime(pr.DeletedAt()),
		pr.ID(),
	)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.Update")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.Update")
	}
	if rows == 0 {
		return caterrors.ErrProductRequestNotFound
	}
	return nil
}

// GetByID retrieves a ProductRequest by its ID. Returns ErrProductRequestNotFound if no row matches.
func (r *ProductRequestRepo) GetByID(ctx context.Context, q postgressqlx.Querier, id string) (*aggregate.ProductRequest, error) {
	const query = `
		SELECT id, seller_id, category_id, title, description, brand, expected_price,
		       currency, condition, status, images, contact_info,
		       admin_reject_reason, admin_note, created_at, updated_at, deleted_at
		FROM product_requests
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m productRequestModel
	err := q.QueryRowxContext(ctx, query, id).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, caterrors.ErrProductRequestNotFound
		}
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.GetByID")
	}

	pr, err := m.toAggregate()
	if err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.GetByID")
	}
	return pr, nil
}

// ListBySellerID retrieves all non-deleted ProductRequests for a seller.
func (r *ProductRequestRepo) ListBySellerID(ctx context.Context, q postgressqlx.Querier, sellerID string) ([]*aggregate.ProductRequest, error) {
	const query = `
		SELECT id, seller_id, category_id, title, description, brand, expected_price,
		       currency, condition, status, images, contact_info,
		       admin_reject_reason, admin_note, created_at, updated_at, deleted_at
		FROM product_requests
		WHERE seller_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var models []productRequestModel
	if err := q.SelectContext(ctx, &models, query, sellerID); err != nil {
		return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.ListBySellerID")
	}

	result := make([]*aggregate.ProductRequest, 0, len(models))
	for _, m := range models {
		pr, err := m.toAggregate()
		if err != nil {
			return nil, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.ListBySellerID")
		}
		result = append(result, pr)
	}
	return result, nil
}

// Delete soft-deletes a ProductRequest by setting its deleted_at timestamp.
// Returns ErrProductRequestNotFound if no row matches.
func (r *ProductRequestRepo) Delete(ctx context.Context, q postgressqlx.Querier, id string) error {
	const query = `
		UPDATE product_requests
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.Delete")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.Delete")
	}
	if rows == 0 {
		return caterrors.ErrProductRequestNotFound
	}
	return nil
}

// List returns product requests matching the given filter and pagination, plus the total count.
func (r *ProductRequestRepo) List(ctx context.Context, q postgressqlx.Querier,
	filter repository.ListProductRequestsFilter, page postgressqlx.Page) ([]*aggregate.ProductRequest, int, error) {

	var (
		args   []any
		join   string
		where  = "WHERE pr.deleted_at IS NULL"
		argIdx = 1
	)

	if filter.Category != nil {
		join += " JOIN categories c ON c.id = pr.category_id"
		where += " AND c.slug = $" + utils.Itoa(argIdx)
		args = append(args, *filter.Category)
		argIdx++
	}

	if filter.SellerID != nil {
		where += " AND pr.seller_id = $" + utils.Itoa(argIdx)
		args = append(args, *filter.SellerID)
		argIdx++
	}

	if len(filter.Conditions) > 0 {
		where += " AND pr.condition = ANY($" + utils.Itoa(argIdx) + ")"
		args = append(args, pq.Array(filter.Conditions))
		argIdx++
	}

	if len(filter.Statuses) > 0 {
		where += " AND pr.status = ANY($" + utils.Itoa(argIdx) + ")"
		args = append(args, pq.Array(filter.Statuses))
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM product_requests pr" + join + " " + where
	var total int
	if err := q.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.List: count")
	}

	orderBy := buildProductRequestSortClause(filter.Sort)
	dataQuery := `
		SELECT pr.id, pr.seller_id, pr.category_id, pr.title, pr.description, pr.brand,
		       pr.expected_price, pr.currency, pr.condition, pr.status, pr.images,
		       pr.contact_info, pr.admin_reject_reason, pr.admin_note,
		       pr.created_at, pr.updated_at, pr.deleted_at
		FROM product_requests pr` + join + " " + where + `
		ORDER BY ` + orderBy + ` ` + page.SQL()

	var models []productRequestModel
	if err := q.SelectContext(ctx, &models, dataQuery, args...); err != nil {
		return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.List: select")
	}

	result := make([]*aggregate.ProductRequest, 0, len(models))
	for _, m := range models {
		pr, err := m.toAggregate()
		if err != nil {
			return nil, 0, caterrors.ErrInternal.WithCause(err).WithInternal("ProductRequestRepo.List: toAggregate")
		}
		result = append(result, pr)
	}

	return result, total, nil
}

// buildProductRequestSortClause returns a safe ORDER BY expression from a sort param.
// Accepts: "created_at", "-created_at", "expected_price", "-expected_price".
// Invalid/nil values default to "created_at DESC".
func buildProductRequestSortClause(s *string) string {
	switch {
	case s == nil || *s == "":
		return "created_at DESC"
	case *s == "created_at":
		return "created_at ASC"
	case *s == "-created_at":
		return "created_at DESC"
	case *s == "expected_price":
		return "expected_price::numeric ASC"
	case *s == "-expected_price":
		return "expected_price::numeric DESC"
	default:
		return "created_at DESC"
	}
}
