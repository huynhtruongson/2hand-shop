package command

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v85"

	errpkg "github.com/huynhtruongson/2hand-shop/internal/pkg/errors"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
)

// mockResult is a minimal sql.Result implementation for test doubles.
type mockResult struct{}

func (mockResult) LastInsertId() (int64, error) { return 0, nil }
func (mockResult) RowsAffected() (int64, error) { return 0, nil }

// mockTX is a test double for postgressqlx.TX.
type mockTX struct{}

func (mockTX) GetContext(_ context.Context, _ interface{}, _ string, _ ...any) error { return nil }
func (mockTX) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return mockResult{}, nil
}
func (mockTX) NamedExecContext(_ context.Context, _ string, _ any) (sql.Result, error) {
	return mockResult{}, nil
}
func (mockTX) SelectContext(_ context.Context, _ interface{}, _ string, _ ...any) error { return nil }
func (mockTX) NamedQuery(_ string, _ any) (*sqlx.Rows, error)                              { return nil, nil }
func (mockTX) QueryRowxContext(_ context.Context, _ string, _ ...any) *sqlx.Row           { return nil }
func (mockTX) QueryxContext(_ context.Context, _ string, _ ...any) (*sqlx.Rows, error)  { return nil, nil }
func (mockTX) Rollback() error                                                            { return nil }
func (mockTX) Commit() error                                                              { return nil }

// mockCartRepo is a test double for repository.CartRepository.
type mockCartRepo struct {
	getByUserIDFn func(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error)
	saveFn        func(ctx context.Context, q postgressqlx.Querier, cart *aggregate.Cart) error
	deleteFn      func(ctx context.Context, q postgressqlx.Querier, userID string) error
}

func (m *mockCartRepo) GetByUserID(ctx context.Context, q postgressqlx.Querier, userID string) (*aggregate.Cart, error) {
	if m.getByUserIDFn != nil {
		return m.getByUserIDFn(ctx, q, userID)
	}
	return nil, errpkg.NewAppError(errpkg.KindNotFound, "CART_NOT_FOUND", "cart not found")
}

func (m *mockCartRepo) Save(ctx context.Context, q postgressqlx.Querier, cart *aggregate.Cart) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, q, cart)
	}
	return nil
}

func (m *mockCartRepo) Delete(ctx context.Context, q postgressqlx.Querier, userID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, q, userID)
	}
	return nil
}

// mockOrderRepo is a test double for repository.OrderRepository.
type mockOrderRepo struct {
	saveFn   func(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error
	updateFn func(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error
	getByIDFn func(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Order, error)
}

func (m *mockOrderRepo) Save(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, q, order)
	}
	return nil
}

func (m *mockOrderRepo) Update(ctx context.Context, q postgressqlx.Querier, order *aggregate.Order) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, q, order)
	}
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Order, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, q, orderID)
	}
	return nil, errpkg.NewAppError(errpkg.KindNotFound, "ORDER_NOT_FOUND", "order not found")
}

func (m *mockOrderRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListOrdersFilter, page postgressqlx.Page) ([]aggregate.Order, int, error) {
	return nil, 0, nil
}

// mockPaymentRepo is a test double for repository.PaymentRepository.
type mockPaymentRepo struct {
	saveFn               func(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error
	updateFn              func(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error
	getByIDFn            func(ctx context.Context, q postgressqlx.Querier, paymentID string) (*aggregate.Payment, error)
	getByStripeSessionFn  func(ctx context.Context, q postgressqlx.Querier, sessionID string) (*aggregate.Payment, error)
}

func (m *mockPaymentRepo) Save(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, q, payment)
	}
	return nil
}

func (m *mockPaymentRepo) Update(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, q, payment)
	}
	return nil
}

func (m *mockPaymentRepo) GetByID(ctx context.Context, q postgressqlx.Querier, paymentID string) (*aggregate.Payment, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, q, paymentID)
	}
	return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
}

func (m *mockPaymentRepo) GetByOrderID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Payment, error) {
	return nil, nil
}

func (m *mockPaymentRepo) GetByStripeSessionID(ctx context.Context, q postgressqlx.Querier, stripeSessionID string) (*aggregate.Payment, error) {
	if m.getByStripeSessionFn != nil {
		return m.getByStripeSessionFn(ctx, q, stripeSessionID)
	}
	return nil, errpkg.NewAppError(errpkg.KindNotFound, "PAYMENT_NOT_FOUND", "payment not found")
}

func (m *mockPaymentRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListPaymentsFilter, page postgressqlx.Page) ([]aggregate.Payment, int, error) {
	return nil, 0, nil
}

// mockPaymentProvider is a test double for repository.PaymentProvider.
type mockPaymentProvider struct {
	createSessionFn func(ctx context.Context, params repository.CreateSessionParams) (*repository.SessionResult, error)
	getSessionFn    func(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error)
}

func (m *mockPaymentProvider) CreateSession(ctx context.Context, params repository.CreateSessionParams) (*repository.SessionResult, error) {
	if m.createSessionFn != nil {
		return m.createSessionFn(ctx, params)
	}
	return &repository.SessionResult{SessionID: "session_test", URL: "https://test.stripe.com"}, nil
}

func (m *mockPaymentProvider) GetSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	if m.getSessionFn != nil {
		return m.getSessionFn(ctx, sessionID)
	}
	return nil, nil
}

// mockPublisher is a test double for the publisher interface.
type mockPublisher struct {
	publishFn func(ctx context.Context, msg interface{}) error
}

func (m *mockPublisher) PublishMessage(ctx context.Context, msg interface{}, _ ...interface{}) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, msg)
	}
	return nil
}

// mockLogger is a test double for logger.Logger.
type mockLogger struct {
	errFn func(msg string, keysAndValues ...any)
}

func (m *mockLogger) Debug(_ string, _ ...any) {}
func (m *mockLogger) Info(_ string, _ ...any)  {}
func (m *mockLogger) Warn(_ string, _ ...any)  {}
func (m *mockLogger) Error(msg string, keysAndValues ...any) {
	if m.errFn != nil {
		m.errFn(msg, keysAndValues...)
	}
}
func (m *mockLogger) Fatal(_ string, _ ...any) {}
func (m *mockLogger) With(_ ...any) interface{ Error(string, ...any) } {
	return m
}
