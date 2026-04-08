package query

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// GetCheckoutSessionQuery retrieves the status of a Stripe Checkout Session.
type GetCheckoutSessionQuery struct {
	SessionID string
}

// GetCheckoutSessionResponse is the result of GetCheckoutSessionHandler.
type GetCheckoutSessionResponse struct {
	SessionID string
	Status    string
	Amount    int64
	Currency  string
}

// GetCheckoutSessionHandler queries the Stripe API to get session status.
type GetCheckoutSessionHandler struct {
	stripeProvider repository.PaymentProvider
}

// NewGetCheckoutSessionHandler returns a new GetCheckoutSessionHandler.
func NewGetCheckoutSessionHandler(stripeProvider repository.PaymentProvider) *GetCheckoutSessionHandler {
	return &GetCheckoutSessionHandler{stripeProvider: stripeProvider}
}

// Handle retrieves the current status of a Checkout Session from Stripe.
func (h *GetCheckoutSessionHandler) Handle(ctx context.Context, q GetCheckoutSessionQuery) (*GetCheckoutSessionResponse, error) {
	sess, err := h.stripeProvider.GetSession(ctx, q.SessionID)
	if err != nil {
		return nil, err
	}
	return &GetCheckoutSessionResponse{
		SessionID: sess.ID,
		Status:    string(sess.Status),
		Amount:    sess.AmountTotal,
		Currency:  string(sess.Currency),
	}, nil
}
