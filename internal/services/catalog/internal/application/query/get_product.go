package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// GetProductHandler is the interface for handling GetProductQuery.
type GetProductHandler cqrs.QueryHandler[GetProductQuery, *aggregate.Product]

// GetProductQuery represents a request to retrieve a single product by ID.
type GetProductQuery struct {
	ProductID string
}

// getProductHandler implements GetProductHandler.
type getProductHandler struct {
	db   postgressqlx.DB
	repo repository.ProductRepository
}

// NewGetProductHandler returns a GetProductHandler that uses repo for data retrieval.
func NewGetProductHandler(repo repository.ProductRepository, db postgressqlx.DB) GetProductHandler {
	return &getProductHandler{repo: repo, db: db}
}

// Handle processes GetProductQuery and returns the product or an error.
func (h *getProductHandler) Handle(ctx context.Context, q GetProductQuery) (*aggregate.Product, error) {
	product, err := h.repo.GetByID(ctx, h.db, q.ProductID)
	if err != nil {
		if errpkg.IsKind(err, errpkg.KindNotFound) {
			return nil, caterrors.ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}
