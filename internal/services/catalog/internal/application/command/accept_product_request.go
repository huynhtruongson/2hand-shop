package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
)

type AcceptProductRequestHandler cqrs.CommandHandler[AcceptProductRequestCommand, AcceptProductRequestResponse]

type AcceptProductRequestCommand struct {
	ProductRequestID string
}

type AcceptProductRequestResponse struct {
	ProductID string
}

type acceptProductRequestHandler struct {
	productRequestRepo repository.ProductRequestRepository
	productRepo        repository.ProductRepository
	cateRepo           repository.CategoryRepository
	db                 postgressqlx.DB
	publisher          publisher
}

func NewAcceptProductRequestHandler(
	productRequestRepo repository.ProductRequestRepository,
	productRepo repository.ProductRepository,
	cateRepo repository.CategoryRepository,
	db postgressqlx.DB,
	publisher publisher,
) AcceptProductRequestHandler {
	return &acceptProductRequestHandler{
		productRequestRepo: productRequestRepo,
		productRepo:        productRepo,
		cateRepo:           cateRepo,
		db:                 db,
		publisher:          publisher,
	}
}

func (h *acceptProductRequestHandler) Handle(ctx context.Context, cmd AcceptProductRequestCommand) (AcceptProductRequestResponse, error) {
	pr, err := h.productRequestRepo.GetByID(ctx, h.db, cmd.ProductRequestID)
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return AcceptProductRequestResponse{}, err
		}
		return AcceptProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	if err := pr.Approve(); err != nil {
		return AcceptProductRequestResponse{}, err
	}

	var productID string
	err = postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		if err := h.productRequestRepo.Update(ctx, tx, pr); err != nil {
			return err
		}

		product, err := aggregate.NewProduct(
			uuid.NewString(),
			pr.CategoryID(),
			pr.Title(),
			pr.Description(),
			pr.ExpectedPrice(),
			pr.Condition(),
			pr.Images(),
			pr.Brand(),
		)
		if err != nil {
			return err
		}

		if err := h.productRepo.Save(ctx, tx, product); err != nil {
			return err
		}

		cate, err := h.cateRepo.GetByID(ctx, tx, product.CategoryID())
		if err != nil {
			return err
		}

		if err := h.publisher.PublishMessage(ctx, event.NewProductCreatedEvent(product, cate.Name())); err != nil {
			return err
		}

		productID = product.ID()
		return nil
	})

	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return AcceptProductRequestResponse{}, err
		}
		return AcceptProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return AcceptProductRequestResponse{ProductID: productID}, nil
}
