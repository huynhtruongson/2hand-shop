package aggregate

import (
	"strings"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/entity"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/errors"
	commercevo "github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

type Order struct {
	id              string
	userID          string
	items           []entity.OrderItem
	refNumber       string
	subtotalAmount  customtypes.Price
	totalAmount     customtypes.Price
	currency        commercevo.Currency
	status          commercevo.OrderStatus
	shippingAddress *commercevo.ShippingAddress
	createdAt       time.Time
	updatedAt       time.Time
	deletedAt       *time.Time
}

func NewOrder(
	id, userID string,
	items []entity.OrderItem,
	totalAmount customtypes.Price,
	currency commercevo.Currency,
	shippingAddress *commercevo.ShippingAddress,
) (*Order, error) {
	subtotal := customtypes.FromCents(0)
	for _, item := range items {
		subtotal = subtotal.Add(item.Price())
	}

	o := &Order{
		id:              id,
		userID:          userID,
		items:           items,
		refNumber:       utils.GenerateRefNumber("ORD"),
		subtotalAmount:  subtotal,
		totalAmount:     totalAmount,
		currency:        currency,
		status:          commercevo.OrderStatusPending,
		shippingAddress: shippingAddress,
		createdAt:       time.Now().UTC(),
		updatedAt:       time.Now().UTC(),
	}
	if err := o.validate(); err != nil {
		return nil, err
	}
	return o, nil
}

// ── Getters ───────────────────────────────────────────────────────────────────

func (o *Order) ID() string                                   { return o.id }
func (o *Order) UserID() string                               { return o.userID }
func (o *Order) RefNumber() string                            { return o.refNumber }
func (o *Order) Items() []entity.OrderItem                    { return o.items }
func (o *Order) SubtotalAmount() customtypes.Price            { return o.subtotalAmount }
func (o *Order) TotalAmount() customtypes.Price               { return o.totalAmount }
func (o *Order) Currency() commercevo.Currency                { return o.currency }
func (o *Order) Status() commercevo.OrderStatus               { return o.status }
func (o *Order) ShippingAddress() *commercevo.ShippingAddress { return o.shippingAddress }
func (o *Order) CreatedAt() time.Time                         { return o.createdAt }
func (o *Order) UpdatedAt() time.Time                         { return o.updatedAt }

// ── Status transitions ───────────────────────────────────────────────────────

// Confirm moves the order to confirmed state (called after payment confirmation).
func (o *Order) Confirm() error {
	if !o.status.CanTransitionTo(commercevo.OrderStatusConfirmed) {
		return errors.ErrOrderInvalidStatusTransition.
			WithMeta("current_status", o.status.String()).
			WithMeta("target_status", "confirmed")
	}
	o.status = commercevo.OrderStatusConfirmed
	o.updatedAt = time.Now().UTC()
	return nil
}

// Ship moves the order to shipped state.
func (o *Order) Ship() error {
	if !o.status.CanTransitionTo(commercevo.OrderStatusShipped) {
		return errors.ErrOrderInvalidStatusTransition.
			WithMeta("current_status", o.status.String()).
			WithMeta("target_status", "shipped")
	}
	o.status = commercevo.OrderStatusShipped
	o.updatedAt = time.Now().UTC()
	return nil
}

// Deliver moves the order to delivered state.
func (o *Order) Deliver() error {
	if !o.status.CanTransitionTo(commercevo.OrderStatusDelivered) {
		return errors.ErrOrderInvalidStatusTransition.
			WithMeta("current_status", o.status.String()).
			WithMeta("target_status", "delivered")
	}
	o.status = commercevo.OrderStatusDelivered
	o.updatedAt = time.Now().UTC()
	return nil
}

// Cancel cancels the order.
func (o *Order) Cancel() error {
	if !o.status.CanTransitionTo(commercevo.OrderStatusCancelled) {
		return errors.ErrOrderInvalidStatusTransition.
			WithMeta("current_status", o.status.String()).
			WithMeta("target_status", "cancelled")
	}
	o.status = commercevo.OrderStatusCancelled
	o.updatedAt = time.Now().UTC()
	return nil
}

// UnmarshalOrderFromDB reconstructs an Order from persisted data, skipping validation.
func UnmarshalOrderFromDB(
	id, userID string,
	items []entity.OrderItem,
	refNumber string,
	subtotalAmount customtypes.Price,
	totalAmount customtypes.Price,
	currency commercevo.Currency,
	status commercevo.OrderStatus,
	shippingAddress *commercevo.ShippingAddress,
	createdAt, updatedAt time.Time,
) *Order {
	return &Order{
		id:              id,
		userID:          userID,
		items:           items,
		refNumber:       refNumber,
		subtotalAmount:  subtotalAmount,
		totalAmount:     totalAmount,
		currency:        currency,
		status:          status,
		shippingAddress: shippingAddress,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (o *Order) validate() error {
	switch {
	case strings.TrimSpace(o.id) == "":
		return errors.ErrValidation.WithDetail("id", "id is empty")
	case strings.TrimSpace(o.userID) == "":
		return errors.ErrValidation.WithDetail("buyer_id", "buyer_id is empty")
	case len(o.items) == 0:
		return errors.ErrValidation.WithDetail("items", "order must have at least one item")
	case !o.totalAmount.IsPositive():
		return errors.ErrValidation.WithDetail("total_amount", "total_amount must be positive")
	case !o.currency.IsValid():
		return errors.ErrValidation.WithDetail("currency", "currency is not a valid value")
	}
	return nil
}
