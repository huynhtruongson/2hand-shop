package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

// RejectProductRequestHandler defines the command handler interface for rejecting product requests.
type RejectProductRequestHandler cqrs.CommandHandler[RejectProductRequestCommand, RejectProductRequestResponse]

// RejectProductRequestCommand represents the input for an admin to reject a product request.
type RejectProductRequestCommand struct {
	ProductRequestID  string
	AdminRejectReason string
}

// RejectProductRequestResponse is returned on a successful rejection.
type RejectProductRequestResponse struct{}

type rejectProductRequestHandler struct {
	productRequestRepo repository.ProductRequestRepository
	db                 postgressqlx.DB
	publisher          publisher
}

// NewRejectProductRequestHandler creates a new RejectProductRequestHandler.
func NewRejectProductRequestHandler(
	productRequestRepo repository.ProductRequestRepository,
	db postgressqlx.DB,
	publisher publisher,
) RejectProductRequestHandler {
	return &rejectProductRequestHandler{
		productRequestRepo: productRequestRepo,
		db:                 db,
		publisher:          publisher,
	}
}

// Handle processes RejectProductRequestCommand.
// It transitions the product request to rejected status and publishes a domain event.
func (h *rejectProductRequestHandler) Handle(ctx context.Context, cmd RejectProductRequestCommand) (RejectProductRequestResponse, error) {
	pr, err := h.productRequestRepo.GetByID(ctx, h.db, cmd.ProductRequestID)
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return RejectProductRequestResponse{}, err
		}
		return RejectProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	if err := pr.Reject(cmd.AdminRejectReason); err != nil {
		return RejectProductRequestResponse{}, err
	}

	err = postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		if err := h.productRequestRepo.Update(ctx, tx, pr); err != nil {
			return err
		}

		// return h.publisher.PublishMessage(ctx, event.NewProductRequestRejectedEvent(pr))
		return nil
	})

	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return RejectProductRequestResponse{}, err
		}
		return RejectProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return RejectProductRequestResponse{}, nil
}
