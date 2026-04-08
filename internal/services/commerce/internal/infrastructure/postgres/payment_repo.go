package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/postgressqlx"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/aggregate"
	carterrors "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/repository"
	commercevo "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// paymentModel mirrors the payments DB table.
type paymentModel struct {
	ID              string            `db:"id"`
	OrderID         string            `db:"order_id"`
	StripeSessionID sql.NullString    `db:"stripe_session_id"`
	RefNumber       string            `db:"ref_number"`
	TotalAmount     customtypes.Price `db:"total_amount"`
	Currency        string            `db:"currency"`
	Status          string            `db:"status"`
	CreatedAt       sql.NullTime      `db:"created_at"`
	UpdatedAt       sql.NullTime      `db:"updated_at"`
	DeletedAt       sql.NullTime      `db:"deleted_at"`
}

// toPayment reconstructs a domain Payment from a DB row.
func (m paymentModel) toPayment() (*aggregate.Payment, error) {

	currency, err := commercevo.NewCurrencyFromString(m.Currency)
	if err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("paymentModel.toPayment: parse currency")
	}

	status, err := commercevo.NewPaymentStatusFromString(m.Status)
	if err != nil {
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("paymentModel.toPayment: parse status")
	}

	var createdAt, updatedAt sql.NullTime
	if m.CreatedAt.Valid {
		createdAt = m.CreatedAt
	}
	if m.UpdatedAt.Valid {
		updatedAt = m.UpdatedAt
	}

	return aggregate.UnmarshalPaymentFromDB(
		m.ID, m.OrderID,
		m.StripeSessionID.String,
		m.RefNumber, m.TotalAmount, currency, status,
		createdAt.Time, updatedAt.Time,
	), nil
}

// toPaymentModel converts a domain Payment to a DB row model.
func toPaymentModel(p *aggregate.Payment) paymentModel {
	m := paymentModel{
		ID:          p.ID(),
		OrderID:     p.OrderID(),
		RefNumber:   p.RefNumber(),
		TotalAmount: p.Amount(),
		Currency:    p.Currency().String(),
		Status:      p.Status().String(),
	}
	if p.StripeSessionID() != "" {
		m.StripeSessionID = sql.NullString{String: p.StripeSessionID(), Valid: true}
	}
	if p.CreatedAt().IsZero() {
		m.CreatedAt = sql.NullTime{Time: p.CreatedAt(), Valid: false}
	} else {
		m.CreatedAt = sql.NullTime{Time: p.CreatedAt(), Valid: true}
	}
	if p.UpdatedAt().IsZero() {
		m.UpdatedAt = sql.NullTime{Time: p.UpdatedAt(), Valid: false}
	} else {
		m.UpdatedAt = sql.NullTime{Time: p.UpdatedAt(), Valid: true}
	}
	if p.DeletedAt() != nil {
		m.DeletedAt = sql.NullTime{Time: *p.DeletedAt(), Valid: true}
	}
	return m
}

// PaymentRepo implements repository.PaymentRepository using PostgreSQL.
type PaymentRepo struct{}

func NewPaymentRepo() *PaymentRepo {
	return &PaymentRepo{}
}

var _ repository.PaymentRepository = (*PaymentRepo)(nil)

// Save persists a new payment to the database.
func (r *PaymentRepo) Save(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	const insertPayment = `
		INSERT INTO payments (id, order_id, stripe_session_id, ref_number, total_amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	m := toPaymentModel(payment)
	_, err := q.ExecContext(ctx, insertPayment,
		m.ID, m.OrderID,
		m.StripeSessionID,
		m.RefNumber, m.TotalAmount,
		m.Currency, m.Status,
		m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return carterrors.ErrPaymentAlreadyExists
		}
		return carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.Save")
	}
	return nil
}

// Update updates an existing payment in the database.
func (r *PaymentRepo) Update(ctx context.Context, q postgressqlx.Querier, payment *aggregate.Payment) error {
	const updatePayment = `
		UPDATE payments SET
			status = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`

	m := toPaymentModel(payment)
	result, err := q.ExecContext(ctx, updatePayment, m.ID, m.Status, m.UpdatedAt)
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.Update")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.Update: rows affected")
	}
	if rows == 0 {
		return carterrors.ErrPaymentNotFound
	}
	return nil
}

// GetByID retrieves a payment by its aggregate ID.
func (r *PaymentRepo) GetByID(ctx context.Context, q postgressqlx.Querier, paymentID string) (*aggregate.Payment, error) {
	const query = `
		SELECT id, order_id, stripe_intent_id, stripe_session_id, ref_number, total_amount, currency, status, created_at, updated_at
		FROM payments
		WHERE id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m paymentModel
	err := q.QueryRowxContext(ctx, query, paymentID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, carterrors.ErrPaymentNotFound
		}
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.GetByID")
	}
	return m.toPayment()
}

// GetByOrderID retrieves a payment by its associated order ID.
func (r *PaymentRepo) GetByOrderID(ctx context.Context, q postgressqlx.Querier, orderID string) (*aggregate.Payment, error) {
	const query = `
		SELECT id, order_id, stripe_intent_id, stripe_session_id, ref_number, total_amount, currency, status, created_at, updated_at
		FROM payments
		WHERE order_id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m paymentModel
	err := q.QueryRowxContext(ctx, query, orderID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, carterrors.ErrPaymentNotFound
		}
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.GetByOrderID")
	}
	return m.toPayment()
}

// GetByStripeSessionID retrieves a payment by its Stripe Checkout Session ID.
func (r *PaymentRepo) GetByStripeSessionID(ctx context.Context, q postgressqlx.Querier, stripeSessionID string) (*aggregate.Payment, error) {
	const query = `
		SELECT id, order_id, stripe_intent_id, stripe_session_id, ref_number, total_amount, currency, status, created_at, updated_at
		FROM payments
		WHERE stripe_session_id = $1 AND deleted_at IS NULL
		LIMIT 1`

	var m paymentModel
	err := q.QueryRowxContext(ctx, query, stripeSessionID).StructScan(&m)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, carterrors.ErrPaymentNotFound
		}
		return nil, carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.GetByStripeSessionID")
	}
	return m.toPayment()
}

// List returns payments matching the given filter and pagination.
func (r *PaymentRepo) List(ctx context.Context, q postgressqlx.Querier, filter repository.ListPaymentsFilter, page postgressqlx.Page) ([]aggregate.Payment, int, error) {
	args := []any{}
	where := "WHERE deleted_at IS NULL"
	argIdx := 1

	if filter.OrderID != nil {
		where += " AND order_id = $" + itoa(argIdx)
		args = append(args, *filter.OrderID)
		argIdx++
	}
	if len(filter.Statuses) > 0 {
		where += " AND status = ANY($" + itoa(argIdx) + "::varchar[])"
		args = append(args, pq.Array(filter.Statuses))
	}

	countQuery := "SELECT COUNT(*) FROM payments " + where
	var total int
	if err := q.QueryRowxContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.List: count")
	}

	selectQuery := `
		SELECT id, order_id, stripe_intent_id, stripe_session_id, ref_number, total_amount, currency, status, created_at, updated_at
		FROM payments ` + where + " ORDER BY created_at DESC " + page.SQL()
	args = append(args, page.Limit, page.Offset)

	var rows []paymentModel
	if err := q.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return nil, 0, carterrors.ErrInternal.WithCause(err).WithInternal("PaymentRepo.List: select")
	}

	payments := make([]aggregate.Payment, 0, len(rows))
	for _, m := range rows {
		p, err := m.toPayment()
		if err != nil {
			return nil, 0, err
		}
		payments = append(payments, *p)
	}
	return payments, total, nil
}
