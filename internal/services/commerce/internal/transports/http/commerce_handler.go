package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v85"

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

// CreateCheckoutSession handles POST /checkout/sessions.
func (h *CommerceHandler) CreateCheckoutSession(ctx *gin.Context) {
	var req dto.CreateCheckoutSessionRequestDTO
	if err := ctx.ShouldBind(&req); err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, carterrors.ErrUnauthorized)
		return
	}

	result, err := h.app.Commands.CreateCheckoutSession.Handle(ctx, command.CreateCheckoutSessionCommand{
		UserID:          authUser.UserID(),
		SuccessURL:      req.SuccessURL,
		CancelURL:       req.CancelURL,
		ShippingAddress: req.ShippingAddress,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.CreateCheckoutSessionResponseDTO{
		SessionID: result.SessionID,
		URL:       result.URL,
	})
}

// GetCheckoutSession handles GET /checkout/sessions.
func (h *CommerceHandler) GetCheckoutSession(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		utils.ResponseError(ctx, carterrors.ErrValidation.WithDetail("session_id", "session_id query parameter is required"))
		return
	}

	authUser, ok := auth.UserFromCtx(ctx)
	if !ok {
		utils.ResponseError(ctx, carterrors.ErrUnauthorized)
		return
	}

	_ = authUser // reserved for future user-level validation

	result, err := h.app.Queries.GetCheckoutSession.Handle(ctx, query.GetCheckoutSessionQuery{
		SessionID: sessionID,
	})
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.Response(ctx, dto.GetCheckoutSessionResponseDTO{
		SessionID: result.SessionID,
		Status:    result.Status,
		Amount:    result.Amount,
		Currency:  result.Currency,
	})
}

// HandleCheckoutWebhook handles POST /webhooks/stripe.
// No Cognito authentication — Stripe HMAC is the authentication mechanism.
func (h *CommerceHandler) HandleCheckoutWebhook(ctx *gin.Context) {
	rawEvent, exists := ctx.Get("stripe_event")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "stripe_event not found in context"})
		return
	}
	event, ok := rawEvent.(stripe.Event)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid stripe event type"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		sess, ok := event.Data.Object["object"].(*stripe.CheckoutSession)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse checkout session"})
			return
		}

		orderID := ""
		userID := ""
		if sess.Metadata != nil {
			orderID = sess.Metadata["order_id"]
			userID = sess.Metadata["user_id"]
		}

		if orderID == "" || userID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing order_id or user_id in session metadata"})
			return
		}

		cmd := command.CompleteCheckoutCommand{
			StripeSessionID: sess.ID,
			StripeEventID:   event.ID,
			OrderID:         orderID,
			UserID:          userID,
			AmountCents:     sess.AmountTotal,
			Currency:        string(sess.Currency),
		}

		if err := h.app.Commands.CompleteCheckout.Handle(ctx, cmd); err != nil {
			utils.ResponseError(ctx, err)
		}

	default:
		// Acknowledge unhandled event types without error.
	}

	ctx.JSON(http.StatusOK, gin.H{"received": true})
}
