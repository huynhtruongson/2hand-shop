package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/http/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
)

type SearchProductsHandler cqrs.QueryHandler[SearchProductsQuery, SearchProductsResponse]

type SearchProductsQuery struct {
	Page       int
	Limit      int
	Keyword    string
	Category   *string
	Conditions []string
	Sort       *string
}

type SearchProductsResponse struct {
	Products   []*aggregate.Product
	Pagination types.Pagination
}

type SearchProductsFilter struct {
	Page       int
	Limit      int
	Keyword    string
	Category   *string
	Conditions []string
	Sort       *string
}

type ProductSearcher interface {
	SearchProducts(ctx context.Context, filter SearchProductsFilter) ([]*aggregate.Product, int, error)
}

type searchProductsHandler struct {
	indexer ProductSearcher
}

func NewSearchProductsHandler(indexer ProductSearcher) SearchProductsHandler {
	return &searchProductsHandler{indexer: indexer}
}

func (h *searchProductsHandler) Handle(ctx context.Context, q SearchProductsQuery) (SearchProductsResponse, error) {
	products, total, err := h.indexer.SearchProducts(ctx, SearchProductsFilter{
		Page:       q.Page,
		Limit:      q.Limit,
		Keyword:    q.Keyword,
		Category:   q.Category,
		Conditions: q.Conditions,
		Sort:       q.Sort,
	})
	if err != nil {
		return SearchProductsResponse{}, err
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return SearchProductsResponse{
		Products: products,
		Pagination: types.Pagination{
			Page:       q.Page,
			Limit:      q.Limit,
			TotalPages: totalPages,
			TotalItems: total,
		},
	}, nil
}
