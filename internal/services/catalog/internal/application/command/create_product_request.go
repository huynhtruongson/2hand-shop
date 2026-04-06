package command

import (
	"context"

	"github.com/google/uuid"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/aggregate"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// CreateProductRequestHandler defines the command handler interface for seller product requests.
type CreateProductRequestHandler cqrs.CommandHandler[CreateProductRequestCommand, CreateProductRequestResponse]

// CreateProductRequestCommand represents the input for creating a product request.
// SellerID is injected by the HTTP handler from the authenticated user context.
type CreateProductRequestCommand struct {
	SellerID      string
	CategoryID    string
	Title         string
	Description   string
	Brand         *string
	ExpectedPrice customtypes.Price
	Condition     string
	Images        customtypes.Attachments
	ContactInfo   string
	AdminNote     *string
}

type CreateProductRequestResponse struct {
	ProductRequestID string
}

type createProductRequestHandler struct {
	productRequestRepo repository.ProductRequestRepository
	db                 postgressqlx.DB
	publisher          publisher
}

func NewCreateProductRequestHandler(
	productRequestRepo repository.ProductRequestRepository,
	db postgressqlx.DB,
	publisher publisher,
) CreateProductRequestHandler {
	return &createProductRequestHandler{
		productRequestRepo: productRequestRepo,
		db:                 db,
		publisher:          publisher,
	}
}

// Handle processes CreateProductRequestCommand by creating a pending ProductRequest
// and publishing a ProductRequestCreated domain event.
func (h *createProductRequestHandler) Handle(ctx context.Context, cmd CreateProductRequestCommand) (CreateProductRequestResponse, error) {
	id := uuid.NewString()

	cond, err := valueobject.NewConditionFromString(cmd.Condition)
	if err != nil {
		return CreateProductRequestResponse{}, err
	}

	pr, err := aggregate.NewProductRequest(
		id,
		cmd.SellerID,
		cmd.CategoryID,
		cmd.Title,
		cmd.Description,
		cmd.Brand,
		cmd.ExpectedPrice,
		cond,
		cmd.Images,
		cmd.ContactInfo,
		cmd.AdminNote,
	)
	if err != nil {
		return CreateProductRequestResponse{}, err
	}

	err = postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		if err := h.productRequestRepo.Save(ctx, tx, pr); err != nil {
			return err
		}
		// return h.publisher.PublishMessage(ctx, event.NewProductRequestCreatedEvent(pr))
		return nil
	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return CreateProductRequestResponse{}, err
		}
		return CreateProductRequestResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return CreateProductRequestResponse{ProductRequestID: pr.ID()}, nil
}
