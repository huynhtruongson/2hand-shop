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
