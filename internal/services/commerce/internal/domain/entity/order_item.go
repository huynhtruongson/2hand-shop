package entity

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// OrderItem represents a line item in an order.
type OrderItem struct {
	id          string
	orderID     string
	productID   string
	productName string
	price       customtypes.Price
	currency    valueobject.Currency
	createdAt   time.Time
}

// NewOrderItem creates a new OrderItem.
func NewOrderItem(id, orderID, productID, productName string, price customtypes.Price, currency valueobject.Currency) OrderItem {
	return OrderItem{
		id:          id,
		orderID:     orderID,
		productID:   productID,
		productName: productName,
		price:       price,
		currency:    currency,
		createdAt:   time.Now().UTC(),
	}
}

func (i OrderItem) ID() string                     { return i.id }
func (i OrderItem) OrderID() string                { return i.orderID }
func (i OrderItem) ProductID() string              { return i.productID }
func (i OrderItem) ProductName() string            { return i.productName }
func (i OrderItem) Price() customtypes.Price       { return i.price }
func (i OrderItem) Currency() valueobject.Currency { return i.currency }
func (i OrderItem) CreatedAt() time.Time           { return i.createdAt }

// ToStripeUnit converts this Money to Stripe's smallest-unit integer.
//
// For 2-decimal currencies (USD, EUR, GBP) the result is cents:
//
//	MustMoney("10.99", "USD").ToStripeUnit() → 1099
//
// For zero-decimal currencies (JPY, KRW, VND) the result is the exact integer:
//
//	MustMoney("1500", "JPY").ToStripeUnit() → 1500
func (i OrderItem) ToStripeAmountUnit() int64 {
	if i.currency.Decimals() == 0 {
		i := i.price.IntPart()
		return i
	}
	return i.price.Cents()
}
