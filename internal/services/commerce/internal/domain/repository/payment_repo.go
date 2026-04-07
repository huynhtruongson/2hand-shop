package repository

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
)

// PaymentRepository defines the persistence contract for Payment aggregates.
type PaymentRepository interface {
	// Save persists a new Payment to the write database.
	Save(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error

	// Update updates an existing Payment in the write database.
	Update(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error

	// GetByID retrieves a payment by its aggregate ID.
	GetByID(ctx context.Context, q postgressqlx.Querier, paymentID string) (*aggregate.Payment, error)

	// GetByOrderID retrieves a payment by its associated order ID.
	GetByOrderID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Payment, error)

	// List returns payments matching the given filter and pagination, plus the total count.
	List(ctx context.Context, q postgressqlx.Querier, filter ListPaymentsFilter, page postgressqlx.Page) ([]aggregate.Payment, int, error)
}

type ListPaymentsFilter struct {
	OrderID  *string
	Statuses []string
}