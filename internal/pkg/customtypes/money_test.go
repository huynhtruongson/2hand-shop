package customtypes

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		code    string
		want    string // expected Price string
		wantErr require.ErrorAssertionFunc
	}{
		// Valid cases
		{name: "USD 2dp", amount: "10.99", code: "USD", want: "10.99", wantErr: nil},
		{name: "EUR integer", amount: "100", code: "EUR", want: "100", wantErr: nil},
		{name: "VND zero-decimal", amount: "1500", code: "VND", want: "1500", wantErr: nil},
		{name: "JPY zero-decimal", amount: "150", code: "JPY", want: "150", wantErr: nil},
		{name: "KRW zero-decimal", amount: "50000", code: "KRW", want: "50000", wantErr: nil},
		{name: "GBP", amount: "0.01", code: "GBP", want: "0.01", wantErr: nil},
		{name: "zero USD", amount: "0", code: "USD", want: "0", wantErr: nil},

		// Error cases — invalid amount
		{name: "negative amount", amount: "-1", code: "USD", wantErr: require.Error},
		{name: "too many decimals", amount: "1.999", code: "USD", wantErr: require.Error},
		{name: "malformed amount", amount: "abc", code: "USD", wantErr: require.Error},
		{name: "empty amount", amount: "", code: "USD", wantErr: require.Error},

		// Error cases — invalid currency
		{name: "unsupported currency", amount: "10.99", code: "XXX", wantErr: require.Error},
		{name: "empty currency", amount: "10.99", code: "", wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMoney(tt.amount, tt.code)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, m.Price().String())
			require.Equal(t, tt.code, string(m.Currency()))
		})
	}
}

func TestMustMoney(t *testing.T) {
	t.Run("valid input no panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			m := MustMoney("10.99", "USD")
			require.Equal(t, "10.99", m.Price().String())
			require.Equal(t, Currency("USD"), m.Currency())
		})
	})

	t.Run("invalid amount panics", func(t *testing.T) {
		require.Panics(t, func() {
			MustMoney("invalid", "USD")
		})
	})

	t.Run("invalid currency panics", func(t *testing.T) {
		require.Panics(t, func() {
			MustMoney("10.99", "XXX")
		})
	})
}

func TestFromStripeUnit(t *testing.T) {
	tests := []struct {
		unit    int64
		code    string
		wantAmt string
	}{
		// 2-decimal currencies: unit is cents
		{unit: 1099, code: "USD", wantAmt: "10.99"},
		{unit: 0, code: "USD", wantAmt: "0"},
		{unit: 100, code: "EUR", wantAmt: "1"},
		{unit: 1, code: "GBP", wantAmt: "0.01"},
		// Zero-decimal currencies: unit is the exact amount
		{unit: 1500, code: "JPY", wantAmt: "1500"},
		{unit: 50000, code: "KRW", wantAmt: "50000"},
		{unit: 0, code: "VND", wantAmt: "0"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			m := FromStripeUnit(tt.unit, tt.code)
			require.Equal(t, tt.wantAmt, m.Price().String())
			require.Equal(t, tt.code, string(m.Currency()))
		})
	}
}

func TestMoney_ToStripeUnit(t *testing.T) {
	tests := []struct {
		amount string
		code   string
		want   int64
	}{
		// 2-decimal
		{amount: "10.99", code: "USD", want: 1099},
		{amount: "0", code: "USD", want: 0},
		{amount: "0.01", code: "EUR", want: 1},
		{amount: "1", code: "GBP", want: 100},
		// Zero-decimal
		{amount: "1500", code: "JPY", want: 1500},
		{amount: "50000", code: "KRW", want: 50000},
		{amount: "0", code: "VND", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			m := MustMoney(tt.amount, tt.code)
			require.Equal(t, tt.want, m.ToStripeUnit())
		})
	}
}

func TestMoney_StripeRoundTrip(t *testing.T) {
	cases := []struct {
		amount string
		code   string
	}{
		// 2-decimal
		{"10.99", "USD"},
		{"0", "USD"},
		{"1", "EUR"},
		{"0.01", "GBP"},
		// Zero-decimal
		{"1500", "JPY"},
		{"50000", "KRW"},
		{"0", "VND"},
	}

	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			original := MustMoney(tc.amount, tc.code)
			unit := original.ToStripeUnit()
			restored := FromStripeUnit(unit, tc.code)
			require.True(t, original.Price().Equal(restored.Price()),
				"round-trip failed for %s %s: got %s", tc.amount, tc.code, restored.Price().String())
			require.Equal(t, original.Currency(), restored.Currency())
		})
	}
}

func TestMoney_IsZero_IsPositive(t *testing.T) {
	tests := []struct {
		amount    string
		code     string
		isZero   bool
		isPositive bool
	}{
		{"0", "USD", true, false},
		{"0.00", "EUR", true, false},
		{"0.01", "GBP", false, true},
		{"1", "JPY", false, true},
	}

	for _, tt := range tests {
		m := MustMoney(tt.amount, tt.code)
		require.Equal(t, tt.isZero, m.IsZero(), "IsZero failed for %s", tt.amount)
		require.Equal(t, tt.isPositive, m.IsPositive(), "IsPositive failed for %s", tt.amount)
	}
}

func TestMoney_String(t *testing.T) {
	m := MustMoney("10.99", "USD")
	require.Equal(t, "10.99 USD", m.String())

	m2 := MustMoney("1500", "JPY")
	require.Equal(t, "1500 JPY", m2.String())
}

func TestMoney_MarshalJSON(t *testing.T) {
	m := MustMoney("10.99", "USD")
	data, err := m.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"10.99 USD"`, string(data))

	m2 := MustMoney("1500", "JPY")
	data2, err := m2.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `"1500 JPY"`, string(data2))
}

func TestMoney_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		amount  string
		code    string
		wantErr require.ErrorAssertionFunc
	}{
		// Valid
		{name: "USD 2dp", input: `"10.99 USD"`, amount: "10.99", code: "USD", wantErr: nil},
		{name: "JPY zero-decimal", input: `"1500 JPY"`, amount: "1500", code: "JPY", wantErr: nil},
		{name: "zero", input: `"0 USD"`, amount: "0", code: "USD", wantErr: nil},
		{name: "null becomes zero", input: `null`, amount: "0", code: "", wantErr: nil},

		// Error cases
		{name: "invalid amount", input: `"abc USD"`, wantErr: require.Error},
		{name: "invalid currency", input: `"10.99 XXX"`, wantErr: require.Error},
		{name: "bad JSON", input: `{`, wantErr: require.Error},
		{name: "bare number", input: `10.99`, wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Money
			err := m.UnmarshalJSON([]byte(tt.input))
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			require.NoError(t, err)
			if tt.input != "null" {
				require.Equal(t, tt.amount, m.Price().String())
				require.Equal(t, tt.code, string(m.Currency()))
			}
		})
	}
}

func TestMoney_Scan(t *testing.T) {
	tests := []struct {
		name    string
		src     any
		amount  string
		code    string
		wantErr require.ErrorAssertionFunc
	}{
		{name: "nil", src: nil, amount: "0", code: "", wantErr: nil},
		{name: "valid space separator", src: []byte("10.99 USD"), amount: "10.99", code: "USD", wantErr: nil},
		{name: "valid pipe separator", src: []byte("10.99|USD"), amount: "10.99", code: "USD", wantErr: nil},
		{name: "JPY", src: []byte("1500 JPY"), amount: "1500", code: "JPY", wantErr: nil},
		{name: "wrong type string", src: "10.99 USD", wantErr: require.Error},
		{name: "invalid amount", src: []byte("abc USD"), wantErr: require.Error},
		{name: "invalid currency", src: []byte("10.99 XXX"), wantErr: require.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Money
			err := m.Scan(tt.src)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			require.NoError(t, err)
			if tt.src != nil {
				require.Equal(t, tt.amount, m.Price().String())
				require.Equal(t, tt.code, string(m.Currency()))
			}
		})
	}
}

func TestMoney_Value(t *testing.T) {
	m := MustMoney("10.99", "USD")
	val, err := m.Value()
	require.NoError(t, err)

	b, ok := val.([]byte)
	require.True(t, ok)
	require.Equal(t, "10.99 USD", string(b))
}

func TestMoney_DB_RoundTrip(t *testing.T) {
	cases := []struct {
		amount string
		code   string
	}{
		{"10.99", "USD"},
		{"0", "EUR"},
		{"1500", "JPY"},
		{"50000", "KRW"},
		{"0", "VND"},
	}

	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			original := MustMoney(tc.amount, tc.code)
			val, err := original.Value()
			require.NoError(t, err)

			var restored Money
			err = restored.Scan(val)
			require.NoError(t, err)
			require.True(t, original.Price().Equal(restored.Price()))
			require.Equal(t, original.Currency(), restored.Currency())
		})
	}
}

func TestMoney_JSON_RoundTrip(t *testing.T) {
	cases := []struct {
		amount string
		code   string
	}{
		{"10.99", "USD"},
		{"0", "EUR"},
		{"1500", "JPY"},
		{"50000", "KRW"},
	}

	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			original := MustMoney(tc.amount, tc.code)
			data, err := json.Marshal(original)
			require.NoError(t, err)

			var restored Money
			err = json.Unmarshal(data, &restored)
			require.NoError(t, err)
			require.True(t, original.Price().Equal(restored.Price()))
			require.Equal(t, original.Currency(), restored.Currency())
		})
	}
}

func TestMoney_ImplementsInterfaces(t *testing.T) {
	var _ driver.Valuer        = Money{}
	var _ interface{ Scan(any) error } = (*Money)(nil)
	var _ json.Marshaler       = Money{}
	var _ json.Unmarshaler     = (*Money)(nil)
}
