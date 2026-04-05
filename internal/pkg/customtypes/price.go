// Package customtypes provides shared custom types used across all services.
package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// ErrInvalidPriceFormat is returned when a price string cannot be parsed
// or fails validation (empty, negative, or more than 2 decimal places).
var ErrInvalidPriceFormat = errors.New("invalid price format")

// Price represents a precise monetary value with up to 2 decimal places.
// It wraps shopspring/decimal.Decimal for arbitrary-precision arithmetic
// and is safe for financial calculations.
//
// DB storage: TEXT column (e.g. "10.99"). Never use floating-point types.
// Stripe compatibility: use Cents() / FromCents() for integer-cent operations.
type Price struct {
	decimal.Decimal
}

// NewPrice parses and validates a decimal string into a Price.
//
// Accepted formats: "10", "10.0", "10.99", "0.99", "1e2" (scientific notation).
//
// Rejected: empty string, malformed number, negative value, more than 2
// decimal places. Zero (0, 0.00) is accepted; enforce positivity at the
// domain layer.
func NewPrice(value string) (Price, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Price{}, ErrInvalidPriceFormat
	}

	d, err := decimal.NewFromString(value)
	if err != nil {
		return Price{}, ErrInvalidPriceFormat
	}

	if d.IsNegative() {
		return Price{}, ErrInvalidPriceFormat
	}

	// Reject more than 2 decimal places.
	// Exponent() returns the base-10 exponent; e.g. "1.999" → -3.
	if d.Exponent() < -2 {
		return Price{}, ErrInvalidPriceFormat
	}

	return Price{Decimal: d}, nil
}

// MustNewPrice constructs a Price, panicking on invalid input.
// Use only in tests and trusted static initialisation.
func MustNewPrice(value string) Price {
	p, err := NewPrice(value)
	if err != nil {
		panic("customtypes.MustNewPrice: " + err.Error())
	}
	return p
}

// FromCents constructs a Price from an integer-cent value (Stripe format).
// Example: FromCents(1099) → Price("10.99").
func FromCents(cents int64) Price {
	return Price{Decimal: decimal.NewFromInt(cents).Div(decimal.NewFromInt(100))}
}

// Cents converts the Price to integer cents (Stripe's canonical format).
// Monetary rounding is half-up. Examples:
//   - "10.99" → 1099
//   - "1.5"   → 150  (1.50)
//   - "0.009" → 1    (rounds up)
func (p Price) Cents() int64 {
	rounded := p.Decimal.Mul(decimal.NewFromInt(100)).Round(0)
	c := rounded.IntPart()
	return c
}

// IsZero returns true if the price is zero.
func (p Price) IsZero() bool { return p.Decimal.IsZero() }

// IsPositive returns true if the price is strictly greater than zero.
func (p Price) IsPositive() bool { return p.Decimal.IsPositive() }

// GreaterThan returns true if p > other.
func (p Price) GreaterThan(other Price) bool { return p.Decimal.GreaterThan(other.Decimal) }

// LessThan returns true if p < other.
func (p Price) LessThan(other Price) bool { return p.Decimal.LessThan(other.Decimal) }

// Equal returns true if p and other represent the same decimal value.
func (p Price) Equal(other Price) bool { return p.Decimal.Equal(other.Decimal) }

// Add returns p + other.
func (p Price) Add(other Price) Price {
	return Price{Decimal: p.Decimal.Add(other.Decimal)}
}

// Sub returns p - other.
func (p Price) Sub(other Price) Price {
	return Price{Decimal: p.Decimal.Sub(other.Decimal)}
}

// Mul returns p * multiplier. Takes decimal.Decimal to keep the type
// free of circular dependencies when building fee calculators.
func (p Price) Mul(multiplier decimal.Decimal) Price {
	return Price{Decimal: p.Decimal.Mul(multiplier)}
}

// String implements fmt.Stringer. Returns the plain decimal string
// (e.g. "10.9" or "10.99"), preserving decimal.Decimal's output format.
func (p Price) String() string { return p.Decimal.String() }

// GoString implements fmt.GoStringer. Returns a valid Go expression
// that produces the same value when evaluated.
func (p Price) GoString() string {
	return fmt.Sprintf("customtypes.MustNewPrice(%q)", p.Decimal.String())
}

// MarshalJSON implements json.Marshaler. The price is serialised as a
// quoted decimal string to avoid JavaScript floating-point representation
// bugs (e.g. 1.000000e+00).
func (p Price) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Decimal.String())
}

// UnmarshalJSON implements json.Unmarshaler. It accepts two input forms:
//
//   - Quoted string (preferred): "10.99"
//   - Bare number (forgiving):   10.99  (parsed via float64 then re-validated)
//
// In both cases the value is passed through NewPrice, which enforces
// max 2 decimal places and rejects negative values.
func (p *Price) UnmarshalJSON(data []byte) error {
	// Explicit null → zero value.
	if string(data) == "null" {
		*p = Price{}
		return nil
	}

	// Quoted string path: "10.99"
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		inner := string(data[1 : len(data)-1])
		// Unescape common escape sequences.
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		inner = strings.ReplaceAll(inner, `\\`, `\`)
		inner = strings.ReplaceAll(inner, `\n`, "\n")
		inner = strings.ReplaceAll(inner, `\t`, "\t")
		parsed, err := NewPrice(inner)
		if err != nil {
			return err
		}
		*p = parsed
		return nil
	}

	// Bare number path: parse via float64 then re-validate.
	var f float64
	if err := json.Unmarshal(data, &f); err != nil {
		return ErrInvalidPriceFormat
	}
	// Use decimal.Decimal to avoid float64 rounding noise, then re-parse
	// through NewPrice which normalises and enforces max 2dp.
	d := decimal.NewFromFloat(f)
	parsed, err := NewPrice(d.String())
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

// Value implements driver.Valuer. The price is serialised as a UTF-8
// byte slice for storage in a TEXT column.
func (p Price) Value() (driver.Value, error) {
	return []byte(p.Decimal.String()), nil
}

// Scan implements sql.Scanner. It reads a TEXT value from PostgreSQL
// and reconstructs the Price.
func (p *Price) Scan(src any) error {
	if src == nil {
		*p = Price{}
		return nil
	}

	switch v := src.(type) {
	case float64:
		p.Decimal = decimal.NewFromFloat(v)
	case int64:
		p.Decimal = decimal.NewFromInt(v)
	case []byte:
		dec, err := decimal.NewFromString(string(v))
		if err != nil {
			return err
		}
		p.Decimal = dec
	case string:
		dec, err := decimal.NewFromString(v)
		if err != nil {
			return err
		}
		p.Decimal = dec
	default:
		return fmt.Errorf("cannot scan type %T into Price", src)
	}

	return nil
}

// compile-time interface assertions
var (
	_ driver.Valuer                = Price{}
	_ interface{ Scan(any) error } = (*Price)(nil)
	_ json.Marshaler               = Price{}
	_ json.Unmarshaler             = (*Price)(nil)
)
