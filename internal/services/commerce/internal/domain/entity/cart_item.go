package entity

import (
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/customtypes"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/domain/valueobject"
)

// CartItem represents a product added to a cart.
type CartItem struct {
	id          string
	cartID      string
	productID   string
	productName string
	price       customtypes.Price
	currency    valueobject.Currency
	addedAt     time.Time
}

func NewCartItem(id, cartID, productID, productName string, price customtypes.Price, currency valueobject.Currency) CartItem {
	return CartItem{
		id:          id,
		cartID:      cartID,
		productID:   productID,
		productName: productName,
		price:       price,
		currency:    currency,
		addedAt:     time.Now().UTC(),
	}
}

func (i CartItem) ID() string                     { return i.id }
func (i CartItem) CartID() string                 { return i.cartID }
func (i CartItem) ProductID() string              { return i.productID }
func (i CartItem) ProductName() string            { return i.productName }
func (i CartItem) Price() customtypes.Price       { return i.price }
func (i CartItem) Currency() valueobject.Currency { return i.currency }
func (i CartItem) AddedAt() time.Time             { return i.addedAt }

// ToStripeUnit converts this Money to Stripe's smallest-unit integer.
//
// For 2-decimal currencies (USD, EUR, GBP) the result is cents:
//
//	MustMoney("10.99", "USD").ToStripeUnit() → 1099
//
// For zero-decimal currencies (JPY, KRW, VND) the result is the exact integer:
//
//	MustMoney("1500", "JPY").ToStripeUnit() → 1500
func (i CartItem) ToStripeAmountUnit() int64 {
	if i.currency.Decimals() == 0 {
		i := i.price.IntPart()
		return i
	}
	return i.price.Cents()
}
