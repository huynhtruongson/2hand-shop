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
	User       *auth.User
	Page       int
	Limit      int
	Category   *string
	Conditions []string
	Statuses   []string
	Sort       *string
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
	offset := (q.Page - 1) * q.Limit

	filter := repository.ListProductsFilter{
		Category:   q.Category,
		Statuses:   q.Statuses,
		Conditions: q.Conditions,
		Sort:       q.Sort,
	}
	if q.User != nil && !q.User.IsAdmin() {
		filter.Statuses = []string{"published"}
	}
	products, total, err := h.repo.List(ctx, h.db, filter, postgressqlx.NewPage(q.Limit, offset, maxLimit))
	if err != nil {
		return ListProductResponse{}, err
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return ListProductResponse{
		Products: products,
		Pagination: types.Pagination{
			Page:       q.Page,
			Limit:      q.Limit,
			TotalPages: totalPages,
			TotalItems: total,
		},
	}, nil
}
