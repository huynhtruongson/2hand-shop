package repository

import (
	"context"

	"github.com/stripe/stripe-go/v85"

	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
)

type PaymentProvider interface {
	CreateSession(ctx context.Context, params CreateSessionParams) (*SessionResult, error)

	GetSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error)
}

// CreateSessionParams wraps the data needed to build a Stripe Checkout Session.
type CreateSessionParams struct {
	Order      *aggregate.Order
	UserID     string
	SuccessURL string // URL to redirect after successful payment
	CancelURL  string // URL to redirect if user cancels
	Currency   string // lowercase ISO 4217 currency code, e.g. "usd"
}

// SessionResult is returned after a successful CreateSession call.
type SessionResult struct {
	SessionID string
	URL       string
}
