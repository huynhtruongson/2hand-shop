package valueobject

import "errors"

type Currency struct {
	value string
}

var (
	CurrencyUSD = Currency{"USD"}
	CurrencyEUR = Currency{"EUR"}
	CurrencyGBP = Currency{"GBP"}
	CurrencyVND = Currency{"VND"}
	CurrencyJPY = Currency{"JPY"}
	CurrencyKRW = Currency{"KRW"}
)

type currencyInfo struct {
	decimals int
	symbol   string
}

// currencies maps each supported ISO 4217 code to its metadata.
// Zero-decimal currencies (JPY, KRW, VND) have decimals = 0 — Stripe's
// smallest unit IS the amount, with no subdivision.
var currencies = map[Currency]currencyInfo{
	CurrencyUSD: {2, "$"},
	CurrencyEUR: {2, "€"},
	CurrencyGBP: {2, "£"},
	CurrencyVND: {0, "₫"},
	CurrencyJPY: {0, "¥"},
	CurrencyKRW: {0, "₩"},
}

func (c Currency) String() string { return c.value }

// IsValid reports whether c is one of the five defined conditions.
func (c Currency) IsValid() bool {
	switch c.value {
	case "USD", "EUR", "GBP", "VND", "JPY", "KRW":
		return true
	}
	return false
}

// NewCurrencyFromString constructs a Currency from its string representation.
// It returns an error if the value is not a recognised currency.
func NewCurrencyFromString(value string) (Currency, error) {
	switch value {
	case "USD", "usd":
		return CurrencyUSD, nil
	case "EUR", "eur":
		return CurrencyEUR, nil
	case "GBP", "gbp":
		return CurrencyGBP, nil
	case "VND", "vnd":
		return CurrencyVND, nil
	case "JPY", "jpy":
		return CurrencyJPY, nil
	case "KRW", "krw":
		return CurrencyKRW, nil
	}
	return Currency{}, errors.New("invalid currency")
}

func (c Currency) Decimals() int {
	return currencies[c].decimals
}

// Symbol returns the display symbol for this currency, e.g. "$", "€", "₫".
func (c Currency) Symbol() string {
	return currencies[c].symbol
}
