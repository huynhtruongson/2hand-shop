package customtypes

import (
	"errors"
	"strings"
)

// ErrInvalidCurrency is returned when an unsupported currency code is passed
// to NewCurrency.
var ErrInvalidCurrency = errors.New("unsupported currency code")

// Currency is an ISO 4217 currency code, e.g. "USD", "EUR", "VND".
type Currency string

// currencyInfo holds Stripe decimal places and the display symbol.
type currencyInfo struct {
	decimals int
	symbol   string
}

// currencies maps each supported ISO 4217 code to its metadata.
// Zero-decimal currencies (JPY, KRW, VND) have decimals = 0 — Stripe's
// smallest unit IS the amount, with no subdivision.
var currencies = map[Currency]currencyInfo{
	"USD": {2, "$"},
	"EUR": {2, "€"},
	"GBP": {2, "£"},
	"VND": {0, "₫"},
	"JPY": {0, "¥"},
	"KRW": {0, "₩"},
}

// NewCurrency validates and returns a Currency. The input is uppercased
// and trimmed before lookup, so "usd", " USD ", and "USD" are all accepted.
func NewCurrency(code string) (Currency, error) {
	c := Currency(strings.ToUpper(strings.TrimSpace(code)))
	if _, ok := currencies[c]; !ok {
		return "", ErrInvalidCurrency
	}
	return c, nil
}

// MustCurrency constructs a Currency, panicking on invalid input.
// Use only in tests and trusted static initialisation.
func MustCurrency(code string) Currency {
	c, err := NewCurrency(code)
	if err != nil {
		panic("customtypes.MustCurrency: " + err.Error())
	}
	return c
}

// Decimals returns the number of decimal places used by Stripe for this
// currency. Zero-decimal currencies (JPY, KRW, VND) return 0.
func (c Currency) Decimals() int {
	return currencies[c].decimals
}

// Symbol returns the display symbol for this currency, e.g. "$", "€", "₫".
func (c Currency) Symbol() string {
	return currencies[c].symbol
}
