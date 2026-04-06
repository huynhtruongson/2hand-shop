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

// DeleteProductRequestHandler defines the command handler interface for deleting product requests.
type DeleteProductRequestHandler cqrs.CommandHandler[DeleteProductRequestCommand, DeleteProductRequestResponse]

// DeleteProductRequestCommand represents the input for deleting a product request.
type DeleteProductRequestCommand struct {
	ProductRequestID string
	SellerID         string
}

// DeleteProductRequestResponse is returned after a successful deletion.
type DeleteProductRequestResponse struct{}

type deleteProductRequestHandler struct {
	productRequestRepo repository.ProductRequestRepository
	db                 postgressqlx.DB
	publisher          publisher
}

// NewDeleteProductRequestHandler constructs a DeleteProductRequestHandler.
func NewDeleteProductRequestHandler(
	productRequestRepo repository.ProductRequestRepository,
	db postgressqlx.DB,
	publisher publisher,
) DeleteProductRequestHandler {
	return &deleteProductRequestHandler{
		productRequestRepo: productRequestRepo,
		db:                 db,
		publisher:          publisher,
	}
}

// Handle processes DeleteProductRequestCommand.
// The domain layer enforces ownership. Deletion succeeds regardless of status.
func (h *deleteProductRequestHandler) Handle(ctx context.Context, cmd DeleteProductRequestCommand) (DeleteProductRequestResponse, error) {
	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		pr, err := h.productRequestRepo.GetByID(ctx, tx, cmd.ProductRequestID)
		if err != nil {
			return err
		}

		if err := pr.Delete(cmd.SellerID); err != nil {
			return err
		}

		if err := h.productRequestRepo.Delete(ctx, tx, pr.ID()); err != nil {
			return err
		}

		return h.publisher.PublishMessage(ctx, event.NewProductRequestDeletedEvent(pr.ID()))
	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return DeleteProductRequestResponse{}, err
		}
		return DeleteProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return DeleteProductRequestResponse{}, nil
}
