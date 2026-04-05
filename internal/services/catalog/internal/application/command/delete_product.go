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

// DeleteProductHandler defines the command handler interface for soft-deleting a product.
type DeleteProductHandler cqrs.CommandHandler[DeleteProductCommand, DeleteProductResponse]

// DeleteProductCommand carries the ID of the product to soft-delete.
type DeleteProductCommand struct {
	ProductID string
}

// DeleteProductResponse is returned after a successful product deletion.
// It is empty because the product ID is already known from the command.
type DeleteProductResponse struct{}

type deleteProductHandler struct {
	repo      repository.ProductRepository
	db        postgressqlx.DB
	publisher publisher
}

// NewDeleteProductHandler constructs a DeleteProductHandler.
func NewDeleteProductHandler(repo repository.ProductRepository, db postgressqlx.DB, publisher publisher) DeleteProductHandler {
	return &deleteProductHandler{repo: repo, db: db, publisher: publisher}
}

// Handle processes DeleteProductCommand.
// It fetches the product, marks it as deleted (soft-delete), persists the change,
// and returns an empty response on success.
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
