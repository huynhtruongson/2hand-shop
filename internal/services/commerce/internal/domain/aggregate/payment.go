package aggregate

import (
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	commercevo "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

type Payment struct {
	id              string
	orderID         string
	stripeSessionID string
	refNumber       string
	totalAmount     customtypes.Price
	currency        commercevo.Currency
	status          commercevo.PaymentStatus
	createdAt       time.Time
	updatedAt       time.Time
	deletedAt       *time.Time
}

func NewPayment(
	id, orderID, stripeSessionID string,
	totalAmount customtypes.Price,
	currency commercevo.Currency,
) (*Payment, error) {
	p := &Payment{
		id:              id,
		orderID:         orderID,
		stripeSessionID: stripeSessionID,
		refNumber:       utils.GenerateRefNumber("PAY"),
		totalAmount:     totalAmount,
		currency:        currency,
		status:          commercevo.PaymentStatusPending,
		createdAt:       time.Now().UTC(),
		updatedAt:       time.Now().UTC(),
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (p *Payment) ID() string                       { return p.id }
func (p *Payment) OrderID() string                  { return p.orderID }
func (p *Payment) RefNumber() string                { return p.refNumber }
func (p *Payment) StripeSessionID() string          { return p.stripeSessionID }
func (p *Payment) Amount() customtypes.Price        { return p.totalAmount }
func (p *Payment) Currency() commercevo.Currency    { return p.currency }
func (p *Payment) Status() commercevo.PaymentStatus { return p.status }
func (p *Payment) CreatedAt() time.Time             { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time             { return p.updatedAt }
func (p *Payment) DeletedAt() *time.Time            { return p.deletedAt }

// ── Status transitions ───────────────────────────────────────────────────────

// Confirm marks the payment as successfully confirmed.
func (p *Payment) Confirm() error {
	if !p.status.CanTransitionTo(commercevo.PaymentStatusConfirmed) {
		return errors.ErrPaymentAlreadyConfirmed
	}
	p.status = commercevo.PaymentStatusConfirmed
	p.updatedAt = time.Now().UTC()
	return nil
}

// Fail marks the payment as failed.
func (p *Payment) Fail() error {
	if !p.status.CanTransitionTo(commercevo.PaymentStatusFailed) {
		return errors.ErrPaymentInvalidTransition
	}
	p.status = commercevo.PaymentStatusFailed
	p.updatedAt = time.Now().UTC()
	return nil
}

// Refund marks the payment as refunded.
func (p *Payment) Refund() error {
	if !p.status.CanTransitionTo(commercevo.PaymentStatusRefunded) {
		return errors.ErrPaymentInvalidTransition
	}
	p.status = commercevo.PaymentStatusRefunded
	p.updatedAt = time.Now().UTC()
	return nil
}

// UnmarshalPaymentFromDB reconstructs a Payment from persisted data, skipping validation.
func UnmarshalPaymentFromDB(
	id, orderID, stripeSessionID, refNumber string,
	totalAmount customtypes.Price,
	currency commercevo.Currency,
	status commercevo.PaymentStatus,
	createdAt, updatedAt time.Time,
) *Payment {
	return &Payment{
		id:              id,
		orderID:         orderID,
		stripeSessionID: stripeSessionID,
		refNumber:       refNumber,
		totalAmount:     totalAmount,
		currency:        currency,
		status:          status,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (p *Payment) validate() error {
	switch {
	case strings.TrimSpace(p.id) == "":
		return errors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(p.orderID) == "":
		return errors.ErrValidation.WithDetail("order_id", "order_id is empty")
	case !p.totalAmount.IsPositive():
		return errors.ErrValidation.WithDetail("amount", "amount must be positive")
	case !p.currency.IsValid():
		return errors.ErrValidation.WithDetail("currency", "currency is not a valid value")
	}
	return nil
}
