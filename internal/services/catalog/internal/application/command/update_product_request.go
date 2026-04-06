package command

import (
	"context"

	"github.com/LukaGiorgadze/gonull/v2"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	caterrors "github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/domain/valueobject"
)

// UpdateProductRequestHandler defines the command handler interface for updating product requests.
type UpdateProductRequestHandler cqrs.CommandHandler[UpdateProductRequestCommand, UpdateProductRequestResponse]

// UpdateProductRequestCommand represents the input for updating a product request.
// SellerID is injected by the HTTP handler from the authenticated user context.
type UpdateProductRequestCommand struct {
	ProductRequestID string
	SellerID         string
	CategoryID       gonull.Nullable[string]
	Title            gonull.Nullable[string]
	Description      gonull.Nullable[string]
	Brand            gonull.Nullable[string]
	ExpectedPrice    gonull.Nullable[customtypes.Price]
	Condition        gonull.Nullable[string]
	Images           gonull.Nullable[customtypes.Attachments]
	ContactInfo      gonull.Nullable[string]
}

// UpdateProductRequestResponse is returned on a successful update.
type UpdateProductRequestResponse struct{}

type updateProductRequestHandler struct {
	productRequestRepo repository.ProductRequestRepository
	db                 postgressqlx.DB
	publisher          publisher
}

// NewUpdateProductRequestHandler constructs an UpdateProductRequestHandler.
func NewUpdateProductRequestHandler(
	productRequestRepo repository.ProductRequestRepository,
	db postgressqlx.DB,
	publisher publisher,
) UpdateProductRequestHandler {
	return &updateProductRequestHandler{
		productRequestRepo: productRequestRepo,
		db:                 db,
		publisher:          publisher,
	}
}

// Handle processes UpdateProductRequestCommand.
// Authorization and status (pending-only) checks are delegated to the domain layer.
func (h *updateProductRequestHandler) Handle(ctx context.Context, cmd UpdateProductRequestCommand) (UpdateProductRequestResponse, error) {
	var result UpdateProductRequestResponse

	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		pr, err := h.productRequestRepo.GetByID(ctx, tx, cmd.ProductRequestID)
		if err != nil {
			return err
		}

		title := pr.Title()
		if cmd.Title.Present {
			title = cmd.Title.Val
		}

		description := pr.Description()
		if cmd.Description.Present {
			description = cmd.Description.Val
		}

		categoryID := pr.CategoryID()
		if cmd.CategoryID.Present {
			categoryID = cmd.CategoryID.Val
		}

		brand := pr.Brand()
		if cmd.Brand.Present {
			brand = &cmd.Brand.Val
		}

		expectedPrice := pr.ExpectedPrice()
		if cmd.ExpectedPrice.Present {
			expectedPrice = cmd.ExpectedPrice.Val
		}

		condition := pr.Condition()
		if cmd.Condition.Present {
			condition, err = valueobject.NewConditionFromString(cmd.Condition.Val)
			if err != nil {
				return caterrors.ErrValidation.WithCause(err)
			}
		}

		images := pr.Images()
		if cmd.Images.Present {
			images = cmd.Images.Val
		}

		contactInfo := pr.ContactInfo()
		if cmd.ContactInfo.Present {
			contactInfo = cmd.ContactInfo.Val
		}

		if err := pr.Update(cmd.SellerID, title, description, categoryID, brand, expectedPrice, condition, images, contactInfo); err != nil {
			return err
		}

		if err := h.productRequestRepo.Update(ctx, tx, pr); err != nil {
			return err
		}

		// return h.publisher.PublishMessage(ctx, event.NewProductRequestUpdatedEvent(pr))
		return nil
	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return result, err
		}
		return result, caterrors.ErrInternal.WithCause(err)
	}

	return result, nil
}
