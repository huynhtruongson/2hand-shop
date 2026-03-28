package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Money pairs a precise Price with a Currency. It is the canonical monetary
// value object for the entire codebase and is safe for financial calculations
// with full Stripe compatibility.
type Money struct {
	price    Price
	currency Currency
}

// NewMoney constructs a Money value from a decimal amount string and a
// currency code.
//
// Example: NewMoney("10.99", "USD") → Money(price="10.99", currency="USD")
//
// amount is validated by NewPrice (max 2dp, non-negative). currency is
// validated by NewCurrency (must be a supported ISO 4217 code).
func NewMoney(amount, currencyCode string) (Money, error) {
	p, err := NewPrice(amount)
	if err != nil {
		return Money{}, err
	}
	c, err := NewCurrency(currencyCode)
	if err != nil {
		return Money{}, err
	}
	return Money{price: p, currency: c}, nil
}

// MustMoney constructs a Money, panicking on invalid input.
// Use only in tests and trusted static initialisation.
func MustMoney(amount, currencyCode string) Money {
	m, err := NewMoney(amount, currencyCode)
	if err != nil {
		panic("customtypes.MustMoney: " + err.Error())
	}
	return m
}

// FromStripeUnit constructs a Money from a Stripe smallest-unit integer.
// This is the inverse of ToStripeUnit().
//
// For 2-decimal currencies (USD, EUR, GBP) the unit is cents:
//   FromStripeUnit(1099, "USD") → Money("10.99", "USD")
//
// For zero-decimal currencies (JPY, KRW, VND) the unit IS the amount:
//   FromStripeUnit(1500, "JPY") → Money("1500", "JPY")
func FromStripeUnit(unit int64, currencyCode string) Money {
	cur := MustCurrency(currencyCode)
	if cur.Decimals() == 0 {
		return Money{
			price:    Price{Decimal: decimal.NewFromInt(unit)},
			currency: cur,
		}
	}
	return Money{
		price:    FromCents(unit),
		currency: cur,
	}
}

// ToStripeUnit converts this Money to Stripe's smallest-unit integer.
//
// For 2-decimal currencies (USD, EUR, GBP) the result is cents:
//   MustMoney("10.99", "USD").ToStripeUnit() → 1099
//
// For zero-decimal currencies (JPY, KRW, VND) the result is the exact integer:
//   MustMoney("1500", "JPY").ToStripeUnit() → 1500
func (m Money) ToStripeUnit() int64 {
	if m.currency.Decimals() == 0 {
		i := m.price.IntPart()
		return i
	}
	return m.price.Cents()
}

// Price returns the numeric price.
func (m Money) Price() Price { return m.price }

// Currency returns the currency code.
func (m Money) Currency() Currency { return m.currency }

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool { return m.price.IsZero() }

// IsPositive returns true if the amount is strictly greater than zero.
func (m Money) IsPositive() bool { return m.price.IsPositive() }

// String returns a human-readable representation, e.g. "10.99 USD".
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.price.String(), m.currency)
}

// MarshalJSON implements json.Marshaler. The money is serialised as a
// quoted string in the format "amount CURRENCY".
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// UnmarshalJSON implements json.Unmarshaler. It accepts a quoted string
// in the format "amount CURRENCY" (e.g. "10.99 USD").
func (m *Money) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*m = Money{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Expected format: "10.99 USD" — split on first space.
	idx := strings.IndexByte(s, ' ')
	if idx < 0 {
		return ErrInvalidPriceFormat
	}

	amount := strings.TrimSpace(s[:idx])
	currency := strings.TrimSpace(s[idx+1:])

	money, err := NewMoney(amount, currency)
	if err != nil {
		return err
	}
	*m = money
	return nil
}

// Value implements driver.Valuer. The money is serialised as a UTF-8
// byte slice in the format "amount|currency" for a TEXT column.
func (m Money) Value() (driver.Value, error) {
	return []byte(m.String()), nil
}

// Scan implements sql.Scanner. It reads a TEXT value from PostgreSQL
// and reconstructs the Money. The expected format matches Value(): "amount|currency".
func (m *Money) Scan(src any) error {
	if src == nil {
		*m = Money{}
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Money.Scan: expected []byte, got %T", src)
	}

	s := string(b)

	// Expected format: "amount|currency" or "amount CURRENCY"
	idx := strings.IndexByte(s, '|')
	if idx < 0 {
		idx = strings.LastIndexByte(s, ' ')
		if idx < 0 {
			return fmt.Errorf("Money.Scan: invalid format %q", s)
		}
	}

	amount := strings.TrimSpace(s[:idx])
	currency := strings.TrimSpace(s[idx+1:])

	money, err := NewMoney(amount, currency)
	if err != nil {
		return err
	}
	*m = money
	return nil
}

// compile-time interface assertions
var (
	_ driver.Valuer        = Money{}
	_ interface{ Scan(any) error } = (*Money)(nil)
	_ json.Marshaler       = Money{}
	_ json.Unmarshaler     = (*Money)(nil)
)
