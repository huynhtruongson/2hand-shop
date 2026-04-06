package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// PublishProductHandler is the command handler for publishing a draft product.
type PublishProductHandler cqrs.CommandHandler[PublishProductCommand, PublishProductResponse]

// PublishProductCommand is the input DTO for publishing a product.
type PublishProductCommand struct {
	ProductID string
}

// PublishProductResponse is the output DTO for the publish product command.
type PublishProductResponse struct{}

type publishProductHandler struct {
	productRepo repository.ProductRepository
	cateRepo    repository.CategoryRepository
	db          postgressqlx.DB
	publisher   publisher
}

// NewPublishProductHandler constructs a PublishProductHandler.
func NewPublishProductHandler(
	productRepo repository.ProductRepository,
	cateRepo repository.CategoryRepository,
	db postgressqlx.DB,
	publisher publisher,
) PublishProductHandler {
	return &publishProductHandler{
		productRepo: productRepo,
		cateRepo:    cateRepo,
		db:          db,
		publisher:   publisher,
	}
}

// Handle processes a PublishProductCommand, transitioning a draft product to published.
func (h *publishProductHandler) Handle(ctx context.Context, cmd PublishProductCommand) (PublishProductResponse, error) {
	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		product, err := h.productRepo.GetByID(ctx, tx, cmd.ProductID)
		if err != nil {
			return err
		}

		if err := product.Publish(); err != nil {
			return err
		}

		if err := h.productRepo.Update(ctx, tx, product); err != nil {
			return err
		}

		cate, err := h.cateRepo.GetByID(ctx, tx, product.CategoryID())
		if err != nil {
			return err
		}

		return h.publisher.PublishMessage(ctx, event.NewProductUpdatedEvent(product, cate.Name()))
	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return PublishProductResponse{}, err
		}
		return PublishProductResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return PublishProductResponse{}, nil
}
