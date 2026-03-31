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

type UpdateProductHandler cqrs.CommandHandler[UpdateProductCommand, UpdateProductResponse]

type UpdateProductCommand struct {
	ProductID   string
	Title       gonull.Nullable[string]
	Description gonull.Nullable[string]
	Price       gonull.Nullable[customtypes.Price]
	Condition   gonull.Nullable[string]
	Images      gonull.Nullable[customtypes.Attachments]
	Brand       gonull.Nullable[string]
}

type UpdateProductResponse struct{}

type updateProductHandler struct {
	repo repository.ProductRepository
	db   postgressqlx.DB
}

func NewUpdateProductHandler(repo repository.ProductRepository, db postgressqlx.DB) UpdateProductHandler {
	return &updateProductHandler{repo: repo, db: db}
}

func (h *updateProductHandler) Handle(ctx context.Context, cmd UpdateProductCommand) (UpdateProductResponse, error) {
	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		product, err := h.repo.GetByID(ctx, tx, cmd.ProductID)
		if err != nil {
			return err
		}

		title := product.Title()
		if cmd.Title.Present {
			title = cmd.Title.Val
		}
		description := product.Description()
		if cmd.Description.Present {
			description = cmd.Description.Val
		}
		newPrice := product.Price()
		if cmd.Price.Present {
			newPrice = cmd.Price.Val
		}

		newCondition := product.Condition()
		if cmd.Condition.Present {
			newCondition, err = valueobject.NewConditionFromString(cmd.Condition.Val)
			if err != nil {
				return err
			}
		}
		newImages := product.Images()
		if cmd.Images.Present {
			newImages = cmd.Images.Val
		}
		newBrand := product.Brand()
		if cmd.Brand.Present {
			newBrand = &cmd.Brand.Val
		}

		if err := product.Update(title, description, newPrice, newCondition, newImages, newBrand); err != nil {
			return err
		}

		if err := h.repo.Update(ctx, tx, product); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if _, ok := errpkg.As(err); ok {
			return UpdateProductResponse{}, err
		}
		return UpdateProductResponse{}, caterrors.ErrInternal.WithCause(err)
	}

	return UpdateProductResponse{}, nil
}
