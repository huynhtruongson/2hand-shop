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

type DeleteProductHandler cqrs.CommandHandler[DeleteProductCommand, DeleteProductResponse]

type DeleteProductCommand struct {
	ProductID string
}

type DeleteProductResponse struct{}

type deleteProductHandler struct {
	repo      repository.ProductRepository
	db        postgressqlx.DB
	publisher publisher
}

func NewDeleteProductHandler(repo repository.ProductRepository, db postgressqlx.DB, publisher publisher) DeleteProductHandler {
	return &deleteProductHandler{repo: repo, db: db, publisher: publisher}
}

func (h *deleteProductHandler) Handle(ctx context.Context, cmd DeleteProductCommand) (DeleteProductResponse, error) {
	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		product, err := h.repo.GetByID(ctx, tx, cmd.ProductID)
		if err != nil {
			return err // ErrProductNotFound or wrapped ErrInternal — pass through
		}

		product.MarkDeleted()

		if err := h.repo.Delete(ctx, tx, product.ID()); err != nil {
			return err
		}

		return h.publisher.PublishMessage(ctx, event.NewProductDeletedEvent(product.ID()))

	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return DeleteProductResponse{}, err
		}
		return DeleteProductResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return DeleteProductResponse{}, nil
}
