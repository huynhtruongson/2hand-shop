package command

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// RemoveFromCartHandler is the command handler for removing an item from a user's cart.
type RemoveFromCartHandler cqrs.CommandHandler[RemoveFromCartCommand, RemoveFromCartResponse]

// RemoveFromCartCommand is the input DTO for removing an item from the cart.
type RemoveFromCartCommand struct {
	UserID    string
	ProductID string
}

// RemoveFromCartResponse is the output DTO after removing an item from the cart.
type RemoveFromCartResponse struct {
	CartID         string
	TotalItemCount int
}

// removeFromCartHandler implements RemoveFromCartHandler.
type removeFromCartHandler struct {
	cartRepo repository.CartRepository
	db       postgressqlx.DB
}

// NewRemoveFromCartHandler constructs a RemoveFromCartHandler.
func NewRemoveFromCartHandler(cartRepo repository.CartRepository, db postgressqlx.DB) RemoveFromCartHandler {
	return &removeFromCartHandler{
		cartRepo: cartRepo,
		db:       db,
	}
}

// Handle processes a RemoveFromCartCommand, removing the item from the cart if it exists.
func (h *removeFromCartHandler) Handle(ctx context.Context, cmd RemoveFromCartCommand) (RemoveFromCartResponse, error) {
	var resp RemoveFromCartResponse

	err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		cart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			return err
		}

		if !cart.RemoveItem(cmd.ProductID) {
			return errors.ErrCartItemNotFound.WithMeta("product_id", cmd.ProductID)
		}

		if err := h.cartRepo.Save(ctx, tx, cart); err != nil {
			return err
		}

		resp = RemoveFromCartResponse{
			CartID:         cart.ID(),
			TotalItemCount: cart.ItemCount(),
		}
		return nil
	})

	return resp, err
}
