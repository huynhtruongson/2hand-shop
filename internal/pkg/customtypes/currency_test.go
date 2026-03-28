package customtypes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCurrency(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Currency
		wantErr require.ErrorAssertionFunc
	}{
		// Valid cases
		{name: "USD uppercase", input: "USD", want: "USD", wantErr: nil},
		{name: "EUR uppercase", input: "EUR", want: "EUR", wantErr: nil},
		{name: "GBP uppercase", input: "GBP", want: "GBP", wantErr: nil},
		{name: "VND uppercase", input: "VND", want: "VND", wantErr: nil},
		{name: "JPY uppercase", input: "JPY", want: "JPY", wantErr: nil},
		{name: "KRW uppercase", input: "KRW", want: "KRW", wantErr: nil},
		{name: "lowercase eur", input: "eur", want: "EUR", wantErr: nil},
		{name: "mixed case Usd", input: "Usd", want: "USD", wantErr: nil},
		{name: "whitespace trimmed", input: "  USD  ", want: "USD", wantErr: nil},

		// Error cases
		{name: "empty string", input: "", wantErr: require.Error},
		{name: "unsupported code", input: "XXX", wantErr: require.Error},
		{name: "unsupported lowercase", input: "usdxx", wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCurrency(tt.input)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				require.ErrorIs(t, err, ErrInvalidCurrency)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, c)
		})
	}
}

func TestMustCurrency(t *testing.T) {
	t.Run("valid input no panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			c := MustCurrency("usd")
			require.Equal(t, Currency("USD"), c)
		})
	})

	t.Run("invalid input panics", func(t *testing.T) {
		require.Panics(t, func() {
			MustCurrency("invalid")
		})
	})
}

func TestCurrency_Decimals(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{"USD", 2},
		{"EUR", 2},
		{"GBP", 2},
		{"VND", 0},
		{"JPY", 0},
		{"KRW", 0},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			c := MustCurrency(tt.code)
			require.Equal(t, tt.want, c.Decimals())
		})
	}
}

func TestCurrency_Symbol(t *testing.T) {
	tests := []struct {
		code  string
		want  string
	}{
		{"USD", "$"},
		{"EUR", "€"},
		{"GBP", "£"},
		{"VND", "₫"},
		{"JPY", "¥"},
		{"KRW", "₩"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			c := MustCurrency(tt.code)
			require.Equal(t, tt.want, c.Symbol())
		})
	}
}
