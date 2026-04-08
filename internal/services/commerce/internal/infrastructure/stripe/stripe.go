// Package stripe provides infrastructure adapters for Stripe API integration.
package stripe

import (
	"github.com/stripe/stripe-go/v85"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// Init sets the global Stripe API key. It must be called once at startup,
// before any Stripe API calls are made.
func Init(cfg config.StripeConfig) {
	stripe.Key = cfg.APIKey
}

type PaymentProvider struct {
	logger logger.Logger
}

func NewPaymentProvider(logger logger.Logger) *PaymentProvider {
	return &PaymentProvider{logger: logger}
}

var _ repository.PaymentProvider = (*PaymentProvider)(nil)
