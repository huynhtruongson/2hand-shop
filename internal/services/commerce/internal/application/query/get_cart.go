package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/cqrs"
	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// GetCartHandler is the query handler for retrieving a user's cart.
type GetCartHandler cqrs.QueryHandler[GetCartQuery, GetCartResponse]

// GetCartQuery represents a request to retrieve the cart for a given user.
type GetCartQuery struct {
	UserID string
}

// GetCartResponse is the output of the GetCart query.
type GetCartResponse struct {
	Cart aggregate.Cart
}

// getCartHandler implements GetCartHandler.
type getCartHandler struct {
	cartRepo repository.CartRepository
}

// NewGetCartHandler constructs a GetCartHandler.
func NewGetCartHandler(cartRepo repository.CartRepository) GetCartHandler {
	return &getCartHandler{cartRepo: cartRepo}
}

// Handle processes a GetCartQuery, returning the user's cart.
// If no cart exists for the user, an empty in-memory cart is returned (200).
func (h *getCartHandler) Handle(ctx context.Context, q GetCartQuery) (GetCartResponse, error) {
	cart, err := h.cartRepo.GetByUserID(ctx, nil, q.UserID)
	if err != nil {
		if errpkg.IsKind(err, errpkg.KindNotFound) {
			// No cart exists — return an empty in-memory cart (not persisted).
			newCart, newErr := aggregate.NewCart(uuid.NewString(), q.UserID, nil)
			if newErr != nil {
				return GetCartResponse{}, carterrors.ErrInternal.WithCause(newErr)
			}
			return GetCartResponse{Cart: *newCart}, nil
		}
		if _, ok := errpkg.As(err); ok {
			return GetCartResponse{}, err
		}
		return GetCartResponse{}, carterrors.ErrInternal.WithCause(err)
	}
	return GetCartResponse{Cart: *cart}, nil
}
