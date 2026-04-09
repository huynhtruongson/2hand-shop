package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// AddToCartHandler is the command handler for adding an item to a user's cart.
type AddToCartHandler cqrs.CommandHandler[AddToCartCommand, AddToCartResponse]

// AddToCartCommand is the input DTO for adding an item to the cart.
type AddToCartCommand struct {
	UserID      string
	ProductID   string
	ProductName string
	Price       customtypes.Price
}

// AddToCartResponse is the output DTO after adding an item to the cart.
type AddToCartResponse struct {
	CartID         string
	ItemID         string
	TotalItemCount int
}

// addToCartHandler implements AddToCartHandler.
type addToCartHandler struct {
	cartRepo repository.CartRepository
	db       postgressqlx.DB
}

// NewAddToCartHandler constructs an AddToCartHandler.
func NewAddToCartHandler(cartRepo repository.CartRepository, db postgressqlx.DB) AddToCartHandler {
	return &addToCartHandler{
		cartRepo: cartRepo,
		db:       db,
	}
}

// Handle processes an AddToCartCommand, creating the cart if it does not exist.
func (h *addToCartHandler) Handle(ctx context.Context, cmd AddToCartCommand) (AddToCartResponse, error) {
	cartID := uuid.NewString()
	itemID := uuid.NewString()

	var cart *aggregate.Cart

	if err := postgressqlx.ExecTx(ctx, h.db, func(ctx context.Context, tx postgressqlx.TX) error {
		existingCart, err := h.cartRepo.GetByUserID(ctx, tx, cmd.UserID)
		if err != nil {
			if errpkg.IsKind(err, errpkg.KindNotFound) {
				// No cart exists for this user — create one.
				cart, err = aggregate.NewCart(cartID, cmd.UserID, nil)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			cart = existingCart
		}

		// Build the cart item with USD as the default currency.
		item := entity.NewCartItem(
			itemID,
			cart.ID(),
			cmd.ProductID,
			cmd.ProductName,
			cmd.Price,
			valueobject.CurrencyUSD,
		)

		// AddItem upserts by ProductID (existing items are replaced).
		cart.AddItem(item)
		// Persist the cart and its items.
		return h.cartRepo.Save(ctx, tx, cart)
	}); err != nil {
		return AddToCartResponse{}, err
	}

	return AddToCartResponse{
		CartID:         cart.ID(),
		ItemID:         itemID,
		TotalItemCount: cart.ItemCount(),
	}, nil
}
