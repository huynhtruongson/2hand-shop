package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/http/types"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

type ListProductHandler cqrs.QueryHandler[ListProductQuery, ListProductResponse]

type ListProductQuery struct {
	User     auth.User
	Page     int
	Limit    int
	Category *string
	Statuses []string
}

type ListProductResponse struct {
	Products   []aggregate.Product
	Pagination types.Pagination
}

const (
	defaultLimit = 20
	maxLimit     = 100
)

// listProductHandler implements ListProductHandler by delegating to the
// read-side ProductReadRepository.
type listProductHandler struct {
	db   postgressqlx.DB
	repo repository.ProductRepository
}

// NewListProductHandler returns a ListProductHandler that uses repo for
// data retrieval.
func NewListProductHandler(repo repository.ProductRepository, db postgressqlx.DB) ListProductHandler {
	return &listProductHandler{repo: repo, db: db}
}

// Handle processes ListProductQuery and returns a paginated product listing.
// Only published products are returned. Category filter is optional.
func (h *listProductHandler) Handle(ctx context.Context, q ListProductQuery) (ListProductResponse, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	page := q.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	filter := repository.ListProductsFilter{
		Category: q.Category,
		Statuses: q.Statuses,
	}
	if !q.User.IsAdmin() {
		filter.Statuses = []string{"published"}
	}

	products, total, err := h.repo.List(ctx, h.db, filter, postgressqlx.NewPage(limit, offset, maxLimit))
	if err != nil {
		return ListProductResponse{}, err
	}

	totalPages := (total + limit - 1) / limit

	return ListProductResponse{
		Products: products,
		Pagination: types.Pagination{
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
			TotalItems: total,
		},
	}, nil
}
