package http

import (
	"github.com/gin-gonic/gin"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/auth"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/command"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application/query"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/transports/http/dto"
)

type CommerceHandler struct {
	app application.Application
}

func NewCommerceHandler(app application.Application) *CommerceHandler {
	return &CommerceHandler{app: app}
}

// AddToCart handles POST /add_to_cart.
// It reads the authenticated user from context and delegates to the AddToCart command handler.
func (h *CommerceHandler) AddToCart(ctx *gin.Context) {
	var req dto.AddToCartRequestDTO
	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, carterrors.ErrUnauthorized)
		return
	}

	result, err := h.app.Commands.AddToCart.Handle(ctx, command.AddToCartCommand{
		UserID:      authUser.UserID(),
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		Price:       req.Price,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.AddToCartResponseDTO{
		CartID:         result.CartID,
		ItemID:         result.ItemID,
		TotalItemCount: result.TotalItemCount,
	})
}

// GetCart handles GET /cart.
// It reads the authenticated user from context and returns their cart.
func (h *CommerceHandler) GetCart(ctx *gin.Context) {
	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, carterrors.ErrUnauthorized)
		return
	}

	result, err := h.app.Queries.GetCart.Handle(ctx, query.GetCartQuery{
		UserID: authUser.UserID(),
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	// Serialize the Cart aggregate directly as the JSON response body.
	cart := result.Cart
	items := make([]dto.CartItemDTO, 0, cart.ItemCount())
	for _, item := range cart.Items() {
		items = append(items, dto.CartItemDTO{
			ID:          item.ID(),
			ProductID:   item.ProductID(),
			ProductName: item.ProductName(),
			Price:       item.Price(),
			Currency:    item.Currency().String(),
			AddedAt:     item.AddedAt(),
		})
	}

	utils.Response(ctx, dto.GetCartResponseDTO{
		ID:          cart.ID(),
		UserID:      cart.UserID(),
		Items:       items,
		ItemCount:   cart.ItemCount(),
		TotalAmount: cart.TotalAmount(),
		CreatedAt:   cart.CreatedAt(),
		UpdatedAt:   cart.UpdatedAt(),
	})
}

// RemoveFromCart handles DELETE /cart/:product_id.
// It reads the authenticated user from context and delegates to the RemoveFromCart command handler.
func (h *CommerceHandler) RemoveFromCart(ctx *gin.Context) {
	var req dto.RemoveFromCartRequestDTO
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, carterrors.ErrUnauthorized)
		return
	}

	result, err := h.app.Commands.RemoveFromCart.Handle(ctx, command.RemoveFromCartCommand{
		UserID:    authUser.UserID(),
		ProductID: req.ProductID,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.RemoveFromCartResponseDTO{
		CartID:         result.CartID,
		TotalItemCount: result.TotalItemCount,
	})
}
