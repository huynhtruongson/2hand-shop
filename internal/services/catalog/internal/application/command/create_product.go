package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/event"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

type CreateProductHandler cqrs.CommandHandler[CreateProductCommand, CreateProductResponse]

type CreateProductCommand struct {
	CategoryID  string
	Title       string
	Description string
	Price       customtypes.Price
	Condition   string
	Images      customtypes.Attachments
	Brand       *string
}

type CreateProductResponse struct {
	ProductID string
}

type createProductHandler struct {
	repo      repository.ProductRepository
	db        postgressqlx.DB
	publisher interface {
		PublishMessage(ctx context.Context, message types.DomainEvent, opts ...types.MessageOption) error
	}
}

func NewCreateProductHandler(repo repository.ProductRepository, db postgressqlx.DB) CreateProductHandler {
	return &createProductHandler{repo: repo, db: db}
}

func (h *createProductHandler) Handle(ctx context.Context, cmd CreateProductCommand) (CreateProductResponse, error) {
	id := uuid.NewString()

	cond, err := valueobject.NewConditionFromString(cmd.Condition)
	if err != nil {
		return CreateProductResponse{}, err
	}
	product, err := aggregate.NewProduct(
		id, cmd.CategoryID, cmd.Title, cmd.Description,
		cmd.Price, cond, cmd.Images, cmd.Brand,
	)
	if err != nil {
		return CreateProductResponse{}, err
	}

	err = postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		if err := h.repo.Save(ctx, tx, product); err != nil {
			return err
		}
		return h.publisher.PublishMessage(ctx, event.NewProductCreatedEvent(product.ID(), product.Title()))

	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return CreateProductResponse{}, err
		}
		return CreateProductResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return CreateProductResponse{ProductID: product.ID()}, nil
}
