package stripe

import (
	"context"
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v85"
	"github.com/stripe/stripe-go/v85/checkout/session"

	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

func (s *PaymentProvider) CreateSession(ctx context.Context, params repository.CreateSessionParams) (*repository.SessionResult, error) {
	if params.Order == nil || len(params.Order.Items()) == 0 {
		return nil, errors.New("order is empty")
	}

	// Build Stripe line items from cart items.
	itemParams := make([]*stripe.CheckoutSessionLineItemParams, 0, len(params.Order.Items()))
	for _, item := range params.Order.Items() {
		amountCents := item.ToStripeAmountUnit()
		if amountCents <= 0 {
			continue
		}
		itemParams = append(itemParams, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(params.Currency),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.ProductName()),
				},
				UnitAmount: stripe.Int64(amountCents),
			},
			Quantity: stripe.Int64(1),
		})
	}
	// can contain shipping fee or tax

	metadata := map[string]string{
		"order_id": params.Order.ID(),
		"user_id":  params.UserID,
	}

	stripeParams := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems:  itemParams,
		SuccessURL: stripe.String(params.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(params.CancelURL),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
		Metadata: metadata,
	}

	sess, err := session.New(stripeParams)
	if err != nil {
		return nil, fmt.Errorf("stripe session create: %w", err)
	}

	return &repository.SessionResult{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

// GetSession retrieves a Checkout Session by its ID.
func (s *PaymentProvider) GetSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe session get: %w", err)
	}
	return sess, nil
}
