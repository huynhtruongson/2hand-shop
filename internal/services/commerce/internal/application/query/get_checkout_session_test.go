package query

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stripe/stripe-go/v85"

	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// mockPaymentProvider is a test double for repository.PaymentProvider.
type mockPaymentProvider struct {
	getSessionFn func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error)
}

func (m *mockPaymentProvider) CreateSession(ctx context.Context, params repository.CreateSessionParams) (*repository.SessionResult, error) {
	return nil, nil
}

func (m *mockPaymentProvider) GetSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	if m.getSessionFn != nil {
		return m.getSessionFn(ctx, sessionID)
	}
	return nil, errors.New("GetSession not configured in mock")
}

func TestGetCheckoutSessionHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		query          GetCheckoutSessionQuery
		setup          func(t *testing.T) *mockPaymentProvider
		wantErr        bool
		errContains    string
		assertResponse func(t *testing.T, resp *GetCheckoutSessionResponse)
	}{
		{
			name:  "returns session data when Stripe returns a complete session",
			query: GetCheckoutSessionQuery{SessionID: "cs_test_abc123"},
			setup: func(t *testing.T) *mockPaymentProvider {
				provider := &mockPaymentProvider{}
				provider.getSessionFn = func(_ context.Context, sessionID string) (*stripe.CheckoutSession, error) {
					if sessionID != "cs_test_abc123" {
						t.Errorf("expected sessionID=cs_test_abc123, got %q", sessionID)
					}
					return &stripe.CheckoutSession{
						ID:            "cs_test_abc123",
						Status:        stripe.CheckoutSessionStatusComplete,
						AmountTotal:   4999,
						Currency:      stripe.CurrencyUSD,
					}, nil
				}
				return provider
			},
			wantErr: false,
			assertResponse: func(t *testing.T, resp *GetCheckoutSessionResponse) {
				if resp.SessionID != "cs_test_abc123" {
					t.Errorf("expected SessionID cs_test_abc123, got %q", resp.SessionID)
				}
				if resp.Status != string(stripe.CheckoutSessionStatusComplete) {
					t.Errorf("expected Status %q, got %q", stripe.CheckoutSessionStatusComplete, resp.Status)
				}
				if resp.Amount != 4999 {
					t.Errorf("expected Amount 4999, got %d", resp.Amount)
				}
				if resp.Currency != string(stripe.CurrencyUSD) {
					t.Errorf("expected Currency usd, got %q", resp.Currency)
				}
			},
		},
		{
			name:  "returns session with open status",
			query: GetCheckoutSessionQuery{SessionID: "cs_test_open"},
			setup: func(t *testing.T) *mockPaymentProvider {
				provider := &mockPaymentProvider{}
				provider.getSessionFn = func(_ context.Context, _ string) (*stripe.CheckoutSession, error) {
					return &stripe.CheckoutSession{
						ID:           "cs_test_open",
						Status:       stripe.CheckoutSessionStatusOpen,
						AmountTotal:  1999,
						Currency:     stripe.CurrencyEUR,
					}, nil
				}
				return provider
			},
			wantErr: false,
			assertResponse: func(t *testing.T, resp *GetCheckoutSessionResponse) {
				if resp.Status != string(stripe.CheckoutSessionStatusOpen) {
					t.Errorf("expected Status %q, got %q", stripe.CheckoutSessionStatusOpen, resp.Status)
				}
				if resp.Amount != 1999 {
					t.Errorf("expected Amount 1999, got %d", resp.Amount)
				}
				if resp.Currency != string(stripe.CurrencyEUR) {
					t.Errorf("expected Currency eur, got %q", resp.Currency)
				}
			},
		},
		{
			name:  "returns error when GetSession returns ErrNilSession",
			query: GetCheckoutSessionQuery{SessionID: "cs_missing"},
			setup: func(t *testing.T) *mockPaymentProvider {
				provider := &mockPaymentProvider{}
				provider.getSessionFn = func(_ context.Context, _ string) (*stripe.CheckoutSession, error) {
					return nil, errors.New("stripe: no such checkout session: cs_missing")
				}
				return provider
			},
			wantErr:     true,
			errContains: "no such checkout session",
		},
		{
			name:  "returns error when GetSession returns a context error",
			query: GetCheckoutSessionQuery{SessionID: "cs_ctx"},
			setup: func(t *testing.T) *mockPaymentProvider {
				provider := &mockPaymentProvider{}
				provider.getSessionFn = func(_ context.Context, _ string) (*stripe.CheckoutSession, error) {
					return nil, context.DeadlineExceeded
				}
				return provider
			},
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			provider := tc.setup(t)
			handler := NewGetCheckoutSessionHandler(provider)

			resp, err := handler.Handle(context.Background(), tc.query)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertResponse != nil {
				tc.assertResponse(t, resp)
			}
		})
	}
}
